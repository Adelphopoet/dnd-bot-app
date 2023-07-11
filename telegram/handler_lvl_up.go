package telegram

import (
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleLvlUp(message *tgbotapi.Message, msgFrom *tgbotapi.User, prev_command ...interface{}) {
	log.Printf("Start lvl up from user %v", msgFrom.FirstName)

	// Set up character
	activeCharacter, err := game.GetActiveCharacter(b.db, int64(msgFrom.ID))
	if err != nil {
		log.Printf("Error during get active character: ", err)
		b.sendMessage(message.Chat.ID, "Не могу получить текущего персонажа: "+err.Error())
		return
	}
	log.Printf("Chose character: %v", activeCharacter.Name)

	// Check character can lvl up
	canLvl, err := game.CheckCharacterCanLvlUp(activeCharacter)
	if err != nil {
		log.Printf("Error during check can lvl up: ", err)
		b.sendMessage(message.Chat.ID, "Ошибка при проверке возможности поднять уровень: "+err.Error())
	}

	// If character can't reach new lvl break
	if !canLvl {
		b.sendMessage(message.Chat.ID, "Ты слишком слаб для этого.")
		return
	}

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

	// Start lvlling
	_, err = activeCharacter.LvlUp(class)
	if err != nil {
		log.Fatalf("Error during lvling. UserID: %v, CharacterID: %v, ClassID: %v, error: %v", msgFrom.ID, activeCharacter.ID, class.ID, err.Error())
		b.sendMessage(message.Chat.ID, "Ошибка при поднятии уровня: "+err.Error())
		return
	}
	b.sendMessage(message.Chat.ID, "Уровень повышен!")
	b.handleCharacterInfo(message, msgFrom, activeCharacter.Name, "/lvlup")

}
