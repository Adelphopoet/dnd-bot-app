package telegram

import (
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handlePreviousCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Get the user ID
	userID := int64(msgFrom.ID)

	userLogs := game.NewUserCommandLog(b.db)
	prevCommand, err := userLogs.GetUserPreviousCommand(userID)
	if err != nil {
		log.Printf("Error during get prev user log: ", err)
		b.sendMessage(int64(msgFrom.ID), err.Error())
	}
	b.handleCommand(prevCommand, message, msgFrom)
}
