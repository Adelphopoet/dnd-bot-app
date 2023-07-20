package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) HandleAttack(message *tgbotapi.Message, msgFrom *tgbotapi.User, args string, prev_command ...interface{}) {
	// Look up for character
}
