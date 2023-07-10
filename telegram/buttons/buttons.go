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
	buttonText := "Играть"
	callbackData := "/play"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}
