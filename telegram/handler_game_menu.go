package telegram

import (
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleGameCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	userID := msgFrom.ID

	// Get all user characters list
	characters, err := game.GetAllUserCharacters(b.db, int64(userID))
	if err != nil {
		log.Printf("Failed to get user characters: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get user characters. Please try again later.")
		return
	}

	var buttons []tgbotapi.InlineKeyboardButton
	// Создать клавиатуру с кнопками для выбора персонажа

	for _, character := range characters {
		button := tgbotapi.NewInlineKeyboardButtonData(character.Name, character.Name)
		buttons = append(buttons, button)
	}
	buttons = append(buttons, tg_buttons.CreateNewCharacterInlineButton())

	// Create inline keyboards with characters
	replyMarkup := createInlineKeyboardMarkup(buttons)

	// Send chose character menu
	b.sendMessage(message.Chat.ID, "Выбери персонажа:", replyMarkup)

	// Wait for user response
	update, err, was_deligated := b.waitForUserResponse(message.Chat.ID)
	if was_deligated {
		return
	}
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}

	var characterName string
	if update.Message != nil {
		characterName = update.Message.Text
	} else if update.CallbackQuery != nil {
		characterName = update.CallbackQuery.Data
	}
	character, err := game.GetCharacterByName(b.db, characterName)
	if err != nil {
		log.Printf("Error is: %v", err)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при выборе персонажа: "+err.Error()))
	} else {
		game.SetMainCharacter(b.db, int64(msgFrom.ID), character.ID)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Выбран персонаж: "+character.Name))
	}
	b.HandleIngameMenu(message, msgFrom)

}
