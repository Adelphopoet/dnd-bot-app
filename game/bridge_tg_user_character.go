package game

import (
	"database/sql"
	"fmt"
	"time"
)

type BridgeTgUserCharacter struct {
	UserID          int64
	CharacterID     int
	IsMainCharacter bool
	CreateTS        time.Time
	UpdateTS        time.Time
	DeleteTS        sql.NullTime
}

func SaveBridgeTgUserCharacter(db *sql.DB, userID int64, characterID int, isMainCharacter bool) error {
	// Обновление всех записей для данного пользователя
	updateQuery := `
		UPDATE game.bridge_tg_user_character
		SET is_main_character = false
		WHERE user_id = $1
	`
	_, err := db.Exec(updateQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to update bridge tg user character: %v", err)
	}

	// Вставка новой записи для данного пользователя и персонажа
	insertQuery := `
		INSERT INTO game.bridge_tg_user_character (user_id, character_id, is_main_character)
		VALUES ($1, $2, $3)
	`
	_, err = db.Exec(insertQuery, userID, characterID, isMainCharacter)
	if err != nil {
		return fmt.Errorf("failed to save bridge tg user character: %v", err)
	}

	err = SetMainCharacter(db, userID, characterID)
	if err != nil {
		return fmt.Errorf("failed to set main character: %v", err)
	}

	return nil
}

func SetMainCharacter(db *sql.DB, userID int64, characterID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Обновление текущей строки
	query := `
		UPDATE game.bridge_tg_user_character
		SET is_main_character = true
		WHERE user_id = $1 AND character_id = $2
	`
	_, err = tx.Exec(query, userID, characterID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update bridge tg user character: %v", err)
	}

	// Обновление остальных строк
	query = `
		UPDATE game.bridge_tg_user_character
		SET is_main_character = false
		WHERE user_id = $1 AND character_id != $2
	`
	_, err = tx.Exec(query, userID, characterID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update other bridge tg user characters: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
