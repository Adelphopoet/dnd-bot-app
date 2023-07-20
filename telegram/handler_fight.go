package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	"github.com/Adelphopoet/dnd-bot-app/game/action"
	fighting "github.com/Adelphopoet/dnd-bot-app/game/fight"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleFight(message *tgbotapi.Message, msgFrom *tgbotapi.User, opponentName string, prev_command ...interface{}) {
	// Look up for character
	activeCharacter, err := game.GetActiveCharacter(b.db, int64(msgFrom.ID))
	if err != nil {
		log.Printf("Failed to get active character: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	// Check character exist
	if activeCharacter == nil {
		b.sendMessage(message.Chat.ID, "Персонаж не найден.")
		b.HandleIngameMenu(message, msgFrom)
		return
	}

	// Look for opponent
	opponentCharacter, err := game.GetCharacterByName(b.db, opponentName)
	if err != nil {
		log.Printf("Failed to get opponent character: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	if opponentCharacter == nil {
		b.sendMessage(message.Chat.ID, "Персонаж не найден.")
		b.HandleIngameMenu(message, msgFrom)
		return
	}

	// Create or get existing fight
	fight, err := fighting.NewFight(b.db, activeCharacter, opponentCharacter)
	if err != nil {
		log.Printf("Failed to get fight: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка создания боя: "+err.Error())
		return
	}

	info := fmt.Sprintf("В бой!\nID битвы %d", fight.ID)
	b.sendMessage(message.Chat.ID, info)

	newRound, err := fight.CreateOrGetRound()
	if err != nil {
		log.Printf("Failed to create round: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка создания раунда: "+err.Error())
		return
	}
	b.sendMessage(message.Chat.ID, fmt.Sprintf("Начат новый раунд ID: %d", newRound.ID))
	activeTurn, err := fight.WhoseTurn()
	if err != nil {
		log.Printf("Failed to create round: %v", err)
		b.sendMessage(message.Chat.ID, "Ошибка получении текущего хода: "+err.Error())
		return
	}
	activePlayer := activeTurn.Character
	fmt.Printf("\n!!!проверям чей ход")
	if activeCharacter.ID != activePlayer.ID {
		fmt.Printf("\n!!!Не наш ход. Я = %d, ходит = %d", activeCharacter.ID, activePlayer.ID)
		b.sendMessage(message.Chat.ID, "Ходит: "+activePlayer.Name)
	} else {
		// Get all infight actions
		filter := map[string]interface{}{"name": "available in fight", "bool_value": true}
		avaibleActions, err := action.GetActionsByFilter(b.db, filter)
		fmt.Printf("\n!!!Avaibles: %v", avaibleActions)
		if err != nil {
			log.Printf("Failed to get infight actions: %v", err)
			b.sendMessage(message.Chat.ID, "Ошибка получении доступных действий: "+err.Error())
			return
		}

		// Create choose class buttoms
		var buttons []tgbotapi.InlineKeyboardButton
		for _, act := range avaibleActions {
			commandProp, err := act.GetValueByPropertyName("chat command")

			if err != nil {
				log.Fatal("Failed to get action %v command: %v", act.Name, err)
				b.sendMessage(message.Chat.ID, "Не могу получить значение команды "+act.Name+": "+err.Error())
				return
			}

			if commandProp == nil {
				log.Fatal("Failed to get action %v command: %v: No command found.")
				b.sendMessage(message.Chat.ID, "Не могу получить значение команды "+act.Name+": "+err.Error())
				return
			}

			fmt.Printf("\nCommand is :%v\n\n\n\n", commandProp)
			button := tgbotapi.NewInlineKeyboardButtonData(act.Name, commandProp.Values.StringValue)
			buttons = append(buttons, button)
		}
		replyMarkup := createInlineKeyboardMarkup(buttons)

		b.sendMessage(message.Chat.ID, "Ваш ход", replyMarkup)
	}

}
