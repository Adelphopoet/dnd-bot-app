package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) {
	b.sendMessage(message.Chat.ID, "Unknown command. Please try again.")
}
