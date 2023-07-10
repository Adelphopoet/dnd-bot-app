package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	b.sendMessage(message.Chat.ID, "Непонятная команда. Попробуй /start для начала игры.")
}
