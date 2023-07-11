package telegram

import (
	"log"

	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) HandleIngameMenu(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Get user id
	userID := msgFrom.ID

	buttons := []tgbotapi.InlineKeyboardButton{
		tg_buttons.CreateMyCharacterInlineButton(),
		tg_buttons.CreateMoveInlineButton(),
		tg_buttons.CreateLookAroundInlineButton(),
		tg_buttons.CreatePrevInlineButton(),
		tg_buttons.CreateHelpInlineButton(),
		tg_buttons.CreateStartInlineButton(),
	}
	keyboard := createInlineKeyboardMarkup(buttons)

	msg := tgbotapi.NewMessage(int64(userID), "Выберите опцию:")
	msg.ReplyMarkup = keyboard

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
