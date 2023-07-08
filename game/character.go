package game

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Character struct {
	ID        int
	Name      string
	CreateTS  time.Time
	UpdateTS  time.Time
	DeleteTS  sql.NullTime
	db        *sql.DB
	IsDeleted bool
	ClassIDs  []int
}

func NewCharacter(db *sql.DB, name string, classIDs []int) *Character {
	return &Character{
		Name:     name,
		ClassIDs: classIDs,
		db:       db,
	}
}

func (c *Character) Save() error {
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	query := `
		INSERT INTO game.dim_character ("name")
		VALUES ($1)
		RETURNING id, create_ts, update_ts, delete_ts
	`
	err = tx.QueryRow(query, c.Name).Scan(&c.ID, &c.CreateTS, &c.UpdateTS, &c.DeleteTS)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save character: %v", err)
	}

	// Связывание персонажа с классами в таблице bridge_character_class
	for _, classID := range c.ClassIDs {
		query = `
			INSERT INTO game.bridge_character_class (character_id, class_id)
			VALUES ($1, $2)
		`
		_, err = tx.Exec(query, c.ID, classID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save character class: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (c *Character) Delete() error {
	query := `
		UPDATE game.dim_character
		SET is_deleted = true, update_ts = current_timestamp
		WHERE id = 1;
	`
	err := c.db.QueryRow(query, c.Name).Scan(&c.ID, &c.CreateTS, &c.UpdateTS, &c.DeleteTS)
	if err != nil {
		return fmt.Errorf("failed to delete character: %v", err)
	}
	c.IsDeleted = true // Устанавливаем флаг IsDeleted в true
	return nil
}

func (c *Character) Load() error {
	query := `
		SELECT "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_character
		WHERE id = $1
	`
	err := c.db.QueryRow(query, c.ID).Scan(&c.Name, &c.CreateTS, &c.UpdateTS, &c.DeleteTS, &c.IsDeleted)
	if err != nil {
		return fmt.Errorf("failed to load character: %v", err)
	}
	return nil
}
