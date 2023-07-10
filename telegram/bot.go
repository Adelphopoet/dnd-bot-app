package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Adelphopoet/dnd-bot-app/game"
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
		var msgFrom *tgbotapi.User

		if update.CallbackQuery != nil {
			comand = update.CallbackQuery.Data
			incomeMessage = update.CallbackQuery.Message
			msgFrom = update.CallbackQuery.From

		} else if update.Message != nil {
			comand = update.Message.Text
			incomeMessage = update.Message
			msgFrom = update.Message.From
		}

		//Сохраняем Телеграм юзера в БД
		err = SaveTgUser(b.db, msgFrom)
		if err != nil {
			log.Printf("Failed to save user: %v", err)
		}
		b.handleCommand(comand, incomeMessage, msgFrom)
	}

	return nil
}

func (b *Bot) handleCommand(comand string, message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	//Clear command from args
	firstSpaceIndex := strings.Index(comand, " ")
	var commandWoArgs, args string
	if firstSpaceIndex != -1 {
		commandWoArgs = comand[:firstSpaceIndex]
		args = comand[firstSpaceIndex+1:]
	} else {
		commandWoArgs = comand
		args = ""
	}
	log.Printf("Command %v from user id %v with args %v", commandWoArgs, msgFrom.ID, args)

	//Log command to DB
	commandLog := game.NewUserCommandLog(b.db)
	err := commandLog.LogCommand(int64(msgFrom.ID), commandWoArgs, args)
	if err != nil {
		log.Printf("Failed to log command: %v", err)
	}

	switch commandWoArgs {
	case "/new_character":
		b.handleCreateCharacter(message, msgFrom)
	case "/start":
		b.handleStartCommand(message, msgFrom)
	case "/play":
		b.handleGameCommand(message, msgFrom)
	case "/go":
		b.handleMoveCommand(message, msgFrom)
	case "/look_araund":
		b.handleLookAraundCommand(message, msgFrom)
	case "/prev":
		b.handlePreviousCommand(message, msgFrom)
	case "/character_info":
		b.handleCharacterInfo(message, msgFrom, args)
	case "/ingame":
		b.HandleIngameMenu(message, msgFrom)
	case "/help":
		b.HandleHelpMenu(message, msgFrom)
	default:
		b.handleUnknownCommand(message, msgFrom)
	}
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

func (b *Bot) waitForUserResponse(chatID int64) (update *tgbotapi.Update, err error, was_deligated bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case update := <-b.updates:
			if update.Message != nil && update.Message.Chat.ID == chatID {
				if update.Message.IsCommand() {
					b.handleCommand(update.Message.Text, update.Message, update.Message.From)
					return nil, nil, true
				} else {
					return &update, nil, false
				}
			} else if update.CallbackQuery != nil && update.CallbackQuery.Message.Chat.ID == chatID {
				if strings.HasPrefix(update.CallbackQuery.Data, "/") {
					b.handleCommand(update.CallbackQuery.Data, update.CallbackQuery.Message, update.CallbackQuery.From)
					return nil, nil, true
				} else {
					return &update, nil, false
				}
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("no user response received within the timeout"), false
		}
	}
}

func (b *Bot) sendMessage(chatID int64, text string, keyboard ...interface{}) {
	if strings.HasPrefix(text, "/") {
		// ignore commands
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	if len(keyboard) > 0 {
		switch k := keyboard[0].(type) {
		case tgbotapi.ReplyKeyboardMarkup:
			msg.ReplyMarkup = k
		case tgbotapi.InlineKeyboardMarkup:
			msg.ReplyMarkup = k
		default:
			log.Println("Invalid keyboard type")
		}
	}
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
