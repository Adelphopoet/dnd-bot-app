package telegram

import (
	"log"
	"strconv"

	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCalculateFormula(message *tgbotapi.Message, msgFrom *tgbotapi.User, formula string, prev_command ...interface{}) {
	log.Printf("Start calculate %v from user %v", formula, msgFrom.FirstName)
	formulaRes, err := calculation.CalculateFormula(&calculation.Formula{Expression: formula})

	if err != nil {
		log.Printf("Error during calculate formula: %v", err)
		b.sendMessage(int64(msgFrom.ID), "Ошибка: "+err.Error())
	} else {
		b.sendMessage(int64(msgFrom.ID), strconv.Itoa(formulaRes))
	}

}
