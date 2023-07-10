package game

import (
	"database/sql"
	"fmt"
	"strings"
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("Персонаж с таким именем уже существует.")
		}
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

func GetAllUserCharacters(db *sql.DB, userID int64) ([]*Character, error) {
	query := `
		SELECT character_id
		FROM game.bridge_tg_user_character
		WHERE user_id = $1
		ORDER BY is_main_character DESC, update_ts ASC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user characters: %v", err)
	}
	defer rows.Close()

	var characters []*Character
	if rows != nil {
		isMainCharacter := true // Флаг для определения главного персонажа
		for rows.Next() {
			var characterID int
			err := rows.Scan(&characterID)
			if err != nil {
				return nil, fmt.Errorf("failed to scan user character: %v", err)
			}

			character, err := GetCharacterByID(db, characterID)
			if err != nil {
				return nil, fmt.Errorf("failed to load user character: %v", err)
			}

			if isMainCharacter {
				isMainCharacter = false
			}

			characters = append(characters, character)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error in rows: %v", err)
		}
	}

	return characters, nil
}

func GetActiveCharacter(db *sql.DB, userID int64) (*Character, error) {
	query := `
		SELECT c.id
		FROM game.dim_character c
		JOIN game.bridge_tg_user_character b ON c.id = b.character_id
		WHERE b.user_id = $1 
		AND b.is_main_character = true 
		AND c.is_deleted = false
		AND b.is_deleted = false
	`
	var characterID int

	err := db.QueryRow(query, userID).Scan(&characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет активного персонажа
		}
		return nil, fmt.Errorf("failed to get active character: %v", err)
	}

	character, err := GetCharacterByID(db, characterID)
	if err != nil {
		return nil, nil // No actual character
	}

	return character, nil
}

func GetCharacterByID(db *sql.DB, characterID int) (*Character, error) {
	query := `
		SELECT "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_character
		WHERE id = $1
	`
	var character Character
	err := db.QueryRow(query, characterID).Scan(&character.Name, &character.CreateTS, &character.UpdateTS, &character.DeleteTS, &character.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %v", err)
	}
	character.db = db
	character.ID = characterID
	return &character, nil
}

func GetCharacterByName(db *sql.DB, characterName string) (*Character, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_character
		WHERE "name" = $1
		AND is_deleted = False
	`
	var character Character
	err := db.QueryRow(query, characterName).Scan(&character.ID, &character.Name, &character.CreateTS, &character.UpdateTS, &character.DeleteTS, &character.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %v", err)
	}
	return &character, nil
}

func (c *Character) SetLocation(locationID int) error {
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Удаление предыдущей записи о местоположении персонажа, если есть
	_, err = tx.Exec("DELETE FROM game.bridge_character_location WHERE character_id = $1", c.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete previous character location: %v", err)
	}

	// Вставка новой записи о местоположении персонажа
	_, err = tx.Exec("INSERT INTO game.bridge_character_location (character_id, location_id) VALUES ($1, $2)", c.ID, locationID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set character location: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (c *Character) GetCurrentLocation() (*Location, error) {
	fmt.Printf("Start GetCurrentLocation with character %v", c.ID)
	query := `
		SELECT l.id
		FROM game.dim_location l
		JOIN game.bridge_character_location b ON l.id = b.location_id
		WHERE b.character_id = $1
	`
	var locationID int
	err := c.db.QueryRow(query, c.ID).Scan(&locationID)
	location, err := GetLocationByID(c.db, locationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет текущей локации для персонажа
		}
		return nil, fmt.Errorf("failed to get current location: %v", err)
	}

	return location, nil
}
