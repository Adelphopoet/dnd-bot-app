package telegram

import (
	"database/sql"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TgUser struct {
	ID           int
	Username     string
	FirstName    string
	LastName     string
	IsBot        bool
	LanguageCode string
	CreateTS     time.Time
	UpdateTS     time.Time
	DeleteTS     sql.NullTime
}

func SaveTgUser(db *sql.DB, user *tgbotapi.User) error {
	query := `
		INSERT INTO game.dim_tg_user (id, username, first_name, last_name, is_bot, language_code, create_ts, update_ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (id) DO UPDATE
		SET username = excluded.username, first_name = excluded.first_name, last_name = excluded.last_name,
			is_bot = excluded.is_bot, language_code = excluded.language_code, update_ts = excluded.update_ts
	`
	_, err := db.Exec(query, user.ID, user.UserName, user.FirstName, user.LastName, user.IsBot, user.LanguageCode, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save or update Telegram user: %v", err)
	}

	return nil
}
