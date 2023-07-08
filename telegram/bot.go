package telegram

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	bot *tgbotapi.BotAPI
	db  *sql.DB
}

func NewBot(botToken string, database *sql.DB) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %v", err)
	}

	return &Bot{
		bot: bot,
		db:  database,
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("failed to get Telegram updates: %v", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/create_character":
			b.handleCreateCharacter(update.Message)
		default:
			b.handleUnknownCommand(update.Message)
		}
	}

	return nil
}

func (b *Bot) handleCreateCharacter(message *tgbotapi.Message) {
	// Получить имя персонажа от пользователя
	name := message.Text

	// Получить все доступные классы
	classes, err := game.GetAllClasses(b.db)
	if err != nil {
		log.Printf("Failed to get classes: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get classes. Please try again later.")
		return
	}

	// Создать клавиатуру с кнопками для выбора классов
	keyboard := createClassKeyboardMarkup(classes)

	// Отправить сообщение с клавиатурой выбора классов
	msg := tgbotapi.NewMessage(message.Chat.ID, "Choose a class for your character:")
	msg.ReplyMarkup = keyboard
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	// Ожидать ответа пользователя с выбранным классом
	b.waitForClassSelection(message.Chat.ID, name, classes)
}

func (b *Bot) waitForClassSelection(chatID int64, name string, classes []game.Class) {
	for {
		// Ожидать ответа пользователя
		update, err := b.waitForUserResponse(chatID)
		if err != nil {
			log.Printf("Failed to wait for user response: %v", err)
			b.sendMessage(chatID, "Failed to create character. Please try again later.")
			return
		}

		// Получить выбранный класс из ответа пользователя
		classID, err := getClassIDFromUserResponse(update, classes)
		if err != nil {
			log.Printf("Failed to get class ID: %v", err)
			b.sendMessage(chatID, "Invalid class selection. Please try again.")
			continue
		}

		// Создать новый персонаж с выбранным классом
		character := game.NewCharacter(b.db, name, []int{classID})

		// Сохранить персонаж в базе данных
		err = character.Save()
		if err != nil {
			log.Printf("Failed to save character: %v", err)
			b.sendMessage(chatID, "Failed to create character. Please try again later.")
			return
		}

		// Отправить ответ пользователю
		response := fmt.Sprintf("Character %s has been created with class ID %d.", character.Name, classID)
		b.sendMessage(chatID, response)
		return
	}
}

func (b *Bot) waitForUserResponse(chatID int64) (*tgbotapi.Update, error) {
	// Ожидать ответа пользователя
	updates := tgbotapi.NewUpdate(0)
	updates.Timeout = 60

	update, err := b.bot.GetUpdatesChan(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to get Telegram updates: %v", err)
	}

	for u := range update {
		if u.Message == nil {
			continue
		}

		if u.Message.Chat.ID == chatID {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("no user response received")
}

func getClassIDFromUserResponse(update *tgbotapi.Update, classes []game.Class) (int, error) {
	// Получить выбранный класс из ответа пользователя
	text := update.Message.Text
	parts := strings.Split(text, "-")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid class selection")
	}

	classID, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, fmt.Errorf("invalid class ID: %v", err)
	}

	// Проверить, что выбранный класс существует
	exists := false
	for _, class := range classes {
		if class.ID == classID {
			exists = true
			break
		}
	}

	if !exists {
		return 0, fmt.Errorf("invalid class ID")
	}

	return classID, nil
}

func createClassKeyboardMarkup(classes []game.Class) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, class := range classes {
		button := tgbotapi.NewKeyboardButton(fmt.Sprintf("%d - %s", class.ID, class.Name))
		row := []tgbotapi.KeyboardButton{button}
		rows = append(rows, row)
	}

	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard: rows,
	}
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) {
	b.sendMessage(message.Chat.ID, "Unknown command. Please try again.")
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
