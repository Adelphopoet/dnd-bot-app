package telegram

import (
	"fmt"
	"log"
	"sort"

	"github.com/Adelphopoet/dnd-bot-app/game"
	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
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

	// Setup start charecteristic
	baseFormula := &calculation.Formula{Expression: "d6"}

	var atts []*game.Attribute
	for _, attName := range [6]string{"str", "dex", "int", "con", "wis", "cha"} {
		att, err := game.GetAttributeByName(b.db, attName)
		if err != nil {
			log.Fatalf("Error during getting att by name %v, %v", att, err)
			b.sendMessage(message.Chat.ID, "Ошибка получения характеристики "+err.Error())
			return
		} else {
			atts = append(atts, att)
		}
	}

	//roll for max 3 from 4 d6 dice for each characteristics
	b.sendMessage(message.Chat.ID, "Ролим характеристики!")
	for _, att := range atts {
		var tmpValList []int
		for i := 0; i < 4; i++ {
			diceRes, err := calculation.CalculateFormula(baseFormula)
			if err != nil {
				log.Fatalf("Can't roll dice: %v", err)
				b.sendMessage(message.Chat.ID, "Ошибка вычисления броска костей: "+err.Error())
				return
			}
			tmpValList = append(tmpValList, diceRes)
		}
		sort.Ints(tmpValList)
		tmpValList = tmpValList[1:4]
		diceSum := 0
		for i := 0; i < 3; i++ {
			diceSum += tmpValList[i]
		}
		err = newCharacter.SetAttributeValue(att, &game.AttributeValue{NumericValue: diceSum})
		if err != nil {
			log.Fatalf("Can't set up characteristic: %v", err)
			b.sendMessage(message.Chat.ID, "Ошибка сохранения характеристики: "+err.Error())
			return
		}
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Характеристике: %s, Установлено значение: %d\n", att.Name, diceSum))
	}

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

	// Lvl up from 0 to 1
	_, err = newCharacter.LvlUp(class)
	if err != nil {
		log.Fatal("New characterID: %v from userID %v can't lvlup: %v", newCharacter.ID, msgFrom.ID, err.Error())
	}

	// At last send in game menu
	b.HandleIngameMenu(message, msgFrom)

}
