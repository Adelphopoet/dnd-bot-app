package buttons

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// Help menu buttom
func CreateHelpInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Помощь"
	callbackData := "/help"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}

// Help menu buttom
func CreateNewCharacterInlineButton() tgbotapi.InlineKeyboardButton {
	buttonText := "Новый персонаж"
	callbackData := "/Новый персонаж"
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	return button
}
