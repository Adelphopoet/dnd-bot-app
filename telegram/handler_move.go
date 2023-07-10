package telegram

import (
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleMoveCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Get the user ID
	userID := msgFrom.ID

	// Get user active character
	activeCharacter, err := game.GetActiveCharacter(b.db, int64(userID))
	if err != nil {
		log.Printf("Failed to get active character: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	// If user doesn't have active character lets suggest to create new
	if activeCharacter == nil {
		// Create choose class buttoms
		var buttons []tgbotapi.InlineKeyboardButton
		button := tg_buttons.CreateNewCharacterInlineButton()
		buttons = append(buttons, button)
		b.sendMessage(message.Chat.ID, "У тебя нет персонажа.", createInlineKeyboardMarkup(buttons))
		return
	}
	log.Printf("Get active character: %v", activeCharacter.Name)

	// Check active character
	if activeCharacter == nil {
		b.sendMessage(message.Chat.ID, "У вас нет активного персонажа.")
		return
	}

	// Get character avaible move pathes
	currentLocation, err := activeCharacter.GetCurrentLocation()
	if err != nil {
		b.sendMessage(message.Chat.ID, "Не могу получить текущую локацию: "+err.Error())
		log.Printf("Failed to get active character: %v", err)
	}
	availablePaths, err := currentLocation.GetAvailablePathes()
	if err != nil {
		log.Printf("Failed to get available paths: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	if len(availablePaths) == 0 {
		b.sendMessage(int64(msgFrom.ID), "Отсюда нет выхода.")
		b.HandleIngameMenu(message, msgFrom)
	}

	// Create select location keyboard
	var buttons []tgbotapi.InlineKeyboardButton
	for _, path := range availablePaths {
		button := tgbotapi.NewInlineKeyboardButtonData(path.Name, path.Name)
		buttons = append(buttons, button)
	}
	keyboard := createInlineKeyboardMarkup(buttons)

	// Send keyboard
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите новую локацию:")
	msg.ReplyMarkup = keyboard
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}
	// Wait for user response
	update, err, was_deligated := b.waitForUserResponse(message.Chat.ID)
	if was_deligated {
		return
	}
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	var destinationLocationName string
	if update.Message != nil {
		destinationLocationName = update.Message.Text
	} else if update.CallbackQuery != nil {
		destinationLocationName = update.CallbackQuery.Data
	} else {
		b.sendMessage(message.Chat.ID, "Нет пункта назначения")
		return
	}

	//Search destination
	log.Printf("Start search destination location by name %v", destinationLocationName)
	destination, err := game.GetLocationByName(b.db, destinationLocationName)
	if err != nil {
		log.Printf("Failed to get location: %v", err)
		b.sendMessage(message.Chat.ID, "Не удалось получить точку назначения: "+err.Error())
	}
	log.Printf("Destination location is %v", activeCharacter.Name)

	// Move character to destination locaton
	err = activeCharacter.SetLocation(destination.ID)
	if err != nil {
		log.Printf("Failed to set location: %v", err)
		b.sendMessage(message.Chat.ID, "Не удалось переместиться: "+err.Error())
	}

	b.sendMessage(message.Chat.ID, "Персонаж "+activeCharacter.Name+" переместился в локацию "+destination.Name)
	b.HandleIngameMenu(message, msgFrom)
}
