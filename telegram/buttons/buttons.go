package tg_buttons

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// Help menu buttom
func CreateHelpInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Помощь"
	callbackData := "/help"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Create new character buttom
func CreateNewCharacterInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Новый персонаж"
	callbackData := "/new_character"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Move buttom
func CreateMoveInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Идти"
	callbackData := "/go"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Play game buttom
func CreatePlayInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Начать игру"
	callbackData := "/play"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Look around buttom
func CreateLookAroundInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Осмотреться"
	callbackData := "/look_araund"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Previous menu button
func CreatePrevInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Назад"
	callbackData := "/prev"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Start menu button
func CreateStartInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "В главное меню"
	callbackData := "/start"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Ingame menu button
func CreateIngameInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "К игре"
	callbackData := "/ingame"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}
