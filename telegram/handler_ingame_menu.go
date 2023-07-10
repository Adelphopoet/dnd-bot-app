package telegram

import (
	"log"

	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) HandleIngameMenu(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Получить идентификатор пользователя
	userID := msgFrom.ID

	buttons := []tgbotapi.InlineKeyboardButton{
		tg_buttons.CreateMoveInlineButton(),
		//tgbotapi.NewInlineKeyboardButtonData("Опция 2", "option2"),
		//tgbotapi.NewInlineKeyboardButtonData("Опция 3", "option3"),
	}
	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			buttons,
		},
	}

	msg := tgbotapi.NewMessage(int64(userID), "Выберите опцию:")
	msg.ReplyMarkup = keyboard

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
