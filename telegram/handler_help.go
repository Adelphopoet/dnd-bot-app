package telegram

import (
	"log"

	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) HandleHelpMenu(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Get user id
	userID := msgFrom.ID

	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonURL("Перейти в GitHub проекта", "https://github.com/Adelphopoet/dnd-bot-app"),
		tg_buttons.CreateIngameInlineButton(),
	}
	keyboard := createInlineKeyboardMarkup(buttons)

	msg := tgbotapi.NewMessage(int64(userID), "Выберите опцию:")
	msg.ReplyMarkup = keyboard

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
