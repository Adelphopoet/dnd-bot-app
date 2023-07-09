package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleMoveCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Get the user ID
	userID := msgFrom.ID

	// Получить активного персонажа пользователя
	activeCharacter, err := game.GetActiveCharacter(b.db, int64(userID))
	if err != nil {
		log.Printf("Failed to get active character: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	// Check active character
	if activeCharacter == nil {
		b.sendMessage(message.Chat.ID, "У вас нет активного персонажа.")
		return
	}

	// Получить доступные пути для персонажа
	currentLocation := activeCharacter
	availablePaths, err := activeCharacter.GetAvailablePaths()
	if err != nil {
		log.Printf("Failed to get available paths: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get available paths. Please try again later.")
		return
	}

	// Создать кнопки клавиатуры для выбора локации
	var buttons []tgbotapi.InlineKeyboardButton
	for _, path := range availablePaths {
		button := tgbotapi.NewInlineKeyboardButtonData(path.LocationTo.Name, fmt.Sprintf("/ChangeLocation %d", path.LocationTo.ID))
		buttons = append(buttons, button)
	}
	keyboard := createInlineKeyboardMarkup(buttons)

	// Отправить сообщение с клавиатурой для выбора локации
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите новую локацию:")
	msg.ReplyMarkup = keyboard
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}
}
