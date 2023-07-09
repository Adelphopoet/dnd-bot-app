package telegram

import (
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleGameCommand(message *tgbotapi.Message, msgFrom *tgbotapi.User) {
	// Получить идентификатор пользователя
	userID := msgFrom.ID
	// Получить список персонажей пользователя
	characters, err := game.GetAllUserCharacters(b.db, int64(userID))
	if err != nil {
		log.Printf("Failed to get user characters: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get user characters. Please try again later.")
		return
	}

	// Создать клавиатуру с кнопками для выбора персонажа
	var buttons []tgbotapi.InlineKeyboardButton
	for _, character := range characters {
		button := tgbotapi.NewInlineKeyboardButtonData(character.Name, character.Name)
		buttons = append(buttons, button)
	}
	//Добавим кнопку создания нового персонажа
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Новый персонаж", "/Новый персонаж"))
	// Создать инлайн-клавиатуру с кнопками выбора класса
	replyMarkup := createInlineKeyboardMarkup(buttons)

	// Отправить сообщение с клавиатурой выбора классов
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выбери персонажа:")
	msg.ReplyMarkup = replyMarkup
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Printf("Error is: %v", err)
	}
	// Ожидать ответа пользователя с выбранным классом
	update, err := b.waitForUserResponse(message.Chat.ID)
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Время вышло."))
		return
	}
	var characterName string
	if update.Message != nil {
		characterName = update.Message.Text
	} else if update.CallbackQuery != nil {
		characterName = update.CallbackQuery.Data
	}
	character, err := game.GetCharacterByName(b.db, characterName)
	if err != nil {
		log.Printf("Error is: %v", err)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при выборе персонажа: "+err.Error()))
	} else {
		game.SetMainCharacter(b.db, int64(msgFrom.ID), character.ID)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Выбран персонаж: "+character.Name))
	}

}
