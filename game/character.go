package game

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Character struct {
	ID       int
	Name     string
	CreateTS time.Time
	UpdateTS time.Time
	DeleteTS sql.NullTime
	db       *sql.DB
}

func NewCharacter(db *sql.DB, name string) *Character {
	return &Character{
		Name: name,
		db:   db,
	}
}

func (c *Character) Save() error {
	query := `
		INSERT INTO game.dim_character ("name")
		VALUES ($1)
		RETURNING id, create_ts, update_ts, delete_ts
	`
	err := c.db.QueryRow(query, c.Name).Scan(&c.ID, &c.CreateTS, &c.UpdateTS, &c.DeleteTS)
	if err != nil {
		return fmt.Errorf("failed to save character: %v", err)
	}
	return nil
}
