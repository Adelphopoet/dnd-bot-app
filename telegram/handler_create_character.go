package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCreateCharacter(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Get all game classes to chose
	classes, err := game.GetAllClasses(b.db)
	if err != nil {
		log.Printf("Failed to get classes: %v", err)
		b.sendMessage(message.Chat.ID, "Не получается загрузить список классов: "+err.Error())
		return
	}

	// Create choose class buttoms
	var buttons []tgbotapi.InlineKeyboardButton
	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(class.Name, class.Name)
		buttons = append(buttons, button)
	}

	// Send to user inline keyboards with classes
	replyMarkup := createInlineKeyboardMarkup(buttons)
	b.sendMessage(message.Chat.ID, "Выбери класс персонажа:", replyMarkup)

	// Wait for user response
	//update, err := b.waitForUserResponse(message.Chat.ID)
	update, err, was_deligated := b.waitForUserResponse(message.Chat.ID)
	if was_deligated {
		return
	}
	if err != nil {
		b.sendMessage(message.Chat.ID, err.Error())
		return
	}

	var className string
	if update.Message != nil {
		className = update.Message.Text
	} else if update.CallbackQuery != nil {
		className = update.CallbackQuery.Data
	}
	class, err := game.GetClassByName(b.db, className)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Произошла ошибка при выборе класса: "+err.Error())
		return
	}

	// Ask user to create character name
	b.sendMessage(message.Chat.ID, "Введи имя персонажа:")

	// Wait for user response
	update, err, was_deligated = b.waitForUserResponse(message.Chat.ID)
	if was_deligated {
		return
	}
	if err != nil {
		b.sendMessage(message.Chat.ID, err.Error())
		return
	}
	characterName := update.Message.Text

	// Creating character
	newCharacter := game.NewCharacter(b.db, characterName, []int{class.ID})
	err = newCharacter.Save()
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при создании персонажа: "+err.Error())
		return
	}

	// Send creation character success message
	response := fmt.Sprintf("Персонаж %s создан с классом %s.", newCharacter.Name, class.Name)
	b.sendMessage(message.Chat.ID, response)

	// Save user and character link
	game.SaveBridgeTgUserCharacter(b.db, int64(msgFrom.ID), newCharacter.ID, true)

	//Move character to first location
	location, err := game.GetLocationByID(b.db, 1)
	if err != nil {
		log.Printf("Failed to get location: %v", err)
	} else {
		newCharacter.SetLocation(location.ID)
		b.sendMessage(message.Chat.ID, "Персонаж "+newCharacter.Name+" появился в локации "+location.Name+".")
	}

	b.HandleIngameMenu(message, msgFrom)

}
