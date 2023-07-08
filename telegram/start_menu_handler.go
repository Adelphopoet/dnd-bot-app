package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	// Отправить стартовое меню

	button1 := tgbotapi.NewInlineKeyboardButtonData("Играть", "/Играть")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Новый персонаж", "/Новый персонаж")
	button3 := tgbotapi.NewInlineKeyboardButtonData("tbd", "/option3")
	buttons := []tgbotapi.InlineKeyboardButton{button1, button2, button3}
	replyMarkup := createInlineKeyboardMarkup(buttons)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome! Choose an option:")
	msg.ReplyMarkup = replyMarkup
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
