package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(botToken string) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %v", err)
	}

	return &Bot{
		bot: bot,
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

	// Создать новый персонаж
	character := game.NewCharacter(name)

	// Сохранить персонаж в базе данных
	err := character.Save()
	if err != nil {
		log.Printf("Failed to save character: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to create character. Please try again later.")
		return
	}

	// Отправить ответ пользователю
	response := fmt.Sprintf("Character %s has been created.", character.Name)
	b.sendMessage(message.Chat.ID, response)
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
