package game

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Class struct {
	ID        int
	Name      string
	CreateTS  time.Time
	UpdateTS  time.Time
	DeleteTS  sql.NullTime
	IsDeleted bool
}

func NewClass(db *sql.DB, name string) (*Class, error) {
	query := `
		INSERT INTO game.dim_class ("name")
		VALUES ($1)
		RETURNING id, create_ts, update_ts, delete_ts
	`
	class := &Class{Name: name}
	err := db.QueryRow(query, class.Name).Scan(&class.ID, &class.CreateTS, &class.UpdateTS, &class.DeleteTS)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %v", err)
	}
	return class, nil
}

func GetAllClasses(db *sql.DB) ([]Class, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts
		FROM game.dim_class
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %v", err)
	}
	defer rows.Close()

	var classes []Class
	for rows.Next() {
		var class Class
		err := rows.Scan(&class.ID, &class.Name, &class.CreateTS, &class.UpdateTS, &class.DeleteTS)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class: %v", err)
		}
		classes = append(classes, class)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in rows: %v", err)
	}

	return classes, nil
}
