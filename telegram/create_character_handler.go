package telegram

import (
	"fmt"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/game"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCreateCharacter(message *tgbotapi.Message) {
	// Получить все доступные классы
	classes, err := game.GetAllClasses(b.db)
	if err != nil {
		log.Printf("Failed to get classes: %v", err)
		b.sendMessage(message.Chat.ID, "Failed to get classes. Please try again later.")
		return
	}

	// Создать кнопки для выбора класса
	var buttons []tgbotapi.InlineKeyboardButton
	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(class.Name, class.Name)
		buttons = append(buttons, button)
	}

	// Создать инлайн-клавиатуру с кнопками выбора класса
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons)

	// Отправить сообщение с клавиатурой выбора классов
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выбери класс персонажа:")
	msg.ReplyMarkup = replyMarkup
	_, err = b.bot.Send(msg)

	// Ожидать ответа пользователя с выбранным классом
	update, err := b.waitForUserResponse(message.Chat.ID)
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Время вышло."))
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
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при выборе класса: "+err.Error()))
		return
	}

	// Отправить сообщение для ввода имени персонажа
	msg = tgbotapi.NewMessage(message.Chat.ID, "Введи имя персонажа:")
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	// Ожидать ответа пользователя с именем персонажа
	update, err = b.waitForUserResponse(message.Chat.ID)
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Время вышло."))
		return
	}
	characterName := update.Message.Text

	// Создать нового персонажа
	newCharacter := game.NewCharacter(b.db, characterName, []int{class.ID})
	err = newCharacter.Save()
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ошибка при создании персонажа: "+err.Error()))
		return
	}

	// Отправить сообщение о создании персонажа
	response := fmt.Sprintf("Персонаж %s создан с классом %s.", newCharacter.Name, class.Name)
	b.sendMessage(message.Chat.ID, response)
}
