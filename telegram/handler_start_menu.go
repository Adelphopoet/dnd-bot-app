package telegram

import (
	"log"

	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleStartCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Отправить стартовое меню
	button1 := tg_buttons.CreatePlayInlineButton()
	button2 := tg_buttons.CreateNewCharacterInlineButton()
	buttons := []tgbotapi.InlineKeyboardButton{button1, button2}
	replyMarkup := createInlineKeyboardMarkup(buttons)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Добро пожаловать!")
	msg.ReplyMarkup = replyMarkup
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
