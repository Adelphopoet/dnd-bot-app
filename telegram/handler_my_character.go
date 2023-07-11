package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tg_buttons "github.com/Adelphopoet/dnd-bot-app/telegram/buttons"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleMyCharacter(message *tgbotapi.Message, msgFrom *tgbotapi.User, prev_command ...interface{}) {
	// Look up for character
	character, err := game.GetActiveCharacter(b.db, int64(msgFrom.ID))
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

	for _, class := range character.CharacterClass {
		info += fmt.Sprintf("Класс: %s, Уровень: %d\n", class.Class.Name, class.Lvl)
	}

	// Create buttons
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
		prevCommand = "/play"
	}

	var buttons []tgbotapi.InlineKeyboardButton
	buttons = append(
		buttons,
		tgbotapi.NewInlineKeyboardButtonData("Назад", prevCommand),
		tg_buttons.CreateIngameInlineButton(),
	)

	// Check character can lvling
	canLvl, err := game.CheckCharacterCanLvlUp(character)
	if err != nil {
		log.Fatalf("Error during check can leveling. UserID: %v, CharacterID: %v, error: %v", msgFrom.ID, character.ID, err.Error())
	}

	// If character has enought EXP, add lvling button to keyboard
	if canLvl {
		buttons = append(
			buttons,
			tg_buttons.CreateLvlupInlineButton(),
		)
	}
	replyMarkup := createInlineKeyboardMarkup(buttons)
	b.sendMessage(message.Chat.ID, info, replyMarkup)
}
