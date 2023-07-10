package telegram

import (
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleLookAraundCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User, prev_command ...interface{}) {
	// Get the user ID
	userID := msgFrom.ID

	// Get user active character
	activeCharacter, err := game.GetActiveCharacter(b.db, int64(userID))
	if err != nil {
		log.Printf("Failed to get active character: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	log.Printf("Get active character: %v", activeCharacter.Name)

	// Check active character
	if activeCharacter == nil {
		b.sendMessage(message.Chat.ID, "У вас нет активного персонажа.")
		return
	}

	// Get current location
	currentLocation, err := activeCharacter.GetCurrentLocation()
	if err != nil {
		b.sendMessage(message.Chat.ID, "Не могу получить текущую локацию: "+err.Error())
		log.Printf("Failed to get active character: %v", err)
	}

	charactersInLocation, err := currentLocation.GetCharactersInLocation()
	if err != nil {
		log.Printf("Error during getting characters in location: ", err.Error())
		b.sendMessage(message.Chat.ID, "Не могу получить список персонажей: "+err.Error())
	}

	var otherCharacters []*game.Character
	for _, character := range charactersInLocation {
		if character.ID != activeCharacter.ID {
			otherCharacters = append(otherCharacters, character)
		}
	}

	//create keyboard to look up characters
	var buttons []tgbotapi.InlineKeyboardButton
	for _, character := range otherCharacters {
		button := tgbotapi.NewInlineKeyboardButtonData(character.Name, "/character_info "+character.Name)
		buttons = append(buttons, button)
	}

	var prevCommand string

	if len(prev_command) > 0 {
		switch c := prev_command[0].(type) {
		case string:
			prevCommand = c
		default:
			log.Println("Invalid prev message type type")
			prevCommand = "/ingame"
		}
	} else {
		prevCommand = "/ingame"
	}

	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Назад", prevCommand))
	replyMarkup := createInlineKeyboardMarkup(buttons)

	// Send visible characters list
	if len(otherCharacters) > 0 {
		b.sendMessage(message.Chat.ID, "Ты видишь персонажей:", replyMarkup)
	} else {
		b.sendMessage(message.Chat.ID, "В локации нет других персонажей.", replyMarkup)
	}
}
