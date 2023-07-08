package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	bot     *tgbotapi.BotAPI
	db      *sql.DB
	updates chan tgbotapi.Update
}

func NewBot(botToken string, database *sql.DB) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %v", err)
	}

	return &Bot{
		bot:     bot,
		db:      database,
		updates: make(chan tgbotapi.Update),
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("failed to get Telegram updates: %v", err)
	}

	go func() {
		for update := range updates {
			if update.Message == nil && update.CallbackQuery == nil {
				continue
			}

			b.updates <- update
		}
	}()

	for update := range b.updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		var comand string
		var incomeMessage *tgbotapi.Message

		if update.CallbackQuery != nil {
			comand = update.CallbackQuery.Data
			incomeMessage = update.CallbackQuery.Message
		} else if update.Message != nil {
			comand = update.Message.Text
			incomeMessage = update.Message
		}

		//Сохраняем Телеграм юзера в БД
		err = SaveTgUser(b.db, incomeMessage.From)
		if err != nil {
			log.Printf("Failed to save user: %v", err)
		}

		switch comand {
		case "/Новый персонаж":
			b.handleCreateCharacter(incomeMessage)
		case "/start":
			b.handleStartCommand(incomeMessage)
		default:
			b.handleUnknownCommand(incomeMessage)
		}
	}

	return nil
}

func createKeyboardMarkup(buttons []string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, class := range buttons {
		button := tgbotapi.NewKeyboardButton(fmt.Sprintf(class))
		row := []tgbotapi.KeyboardButton{button}
		rows = append(rows, row)
	}

	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard: rows,
	}
}

func (b *Bot) waitForUserResponse(chatID int64) (*tgbotapi.Update, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case update := <-b.updates:
			if update.Message != nil && update.Message.Chat.ID == chatID {
				return &update, nil
			} else if update.CallbackQuery != nil && update.CallbackQuery.Message.Chat.ID == chatID {
				return &update, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("no user response received within the timeout")
		}
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func createInlineKeyboardMarkup(buttons []tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, button := range buttons {
		row := []tgbotapi.InlineKeyboardButton{button}
		rows = append(rows, row)
	}

	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
