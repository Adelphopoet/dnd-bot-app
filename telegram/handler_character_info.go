package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCharacterInfo(message *tgbotapi.Message, msgFrom *tgbotapi.User, characterName string, prev_command ...interface{}) {
	// Look up for character
	character, err := game.GetCharacterByName(b.db, characterName)
	if err != nil {
		log.Printf("Failed to get character: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	// Check character exist
	if character == nil {
		b.sendMessage(message.Chat.ID, "Персонаж не найден.")
		b.HandleIngameMenu(message, msgFrom)
		return
	}

	// Send character info
	info := fmt.Sprintf("Имя: %s\nID: %d\nСоздан: %s\nОбновлен: %s",
		character.Name, character.ID, character.CreateTS, character.UpdateTS)

	//create keyboard to look up characters
	var buttons []tgbotapi.InlineKeyboardButton

	var prevCommand string

	if len(prev_command) > 0 {
		switch c := prev_command[0].(type) {
		case string:
			prevCommand = c
		default:
			log.Println("Invalid prev message type type")
			prevCommand = "/look_araund"
		}
	} else {
		prevCommand = "/look_araund"
	}
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Назад", prevCommand))
	replyMarkup := createInlineKeyboardMarkup(buttons)
	b.sendMessage(message.Chat.ID, info, replyMarkup)
}
