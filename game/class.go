package game

import (
	"database/sql"
	"fmt"
	"time"

	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
	_ "github.com/lib/pq"
)

type Class struct {
	ID        int
	Name      string
	CreateTS  time.Time
	UpdateTS  time.Time
	DeleteTS  sql.NullTime
	IsDeleted bool
	db        *sql.DB
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

func GetClassByName(db *sql.DB, className string) (*Class, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts
		FROM game.dim_class
		WHERE "name" = ($1)
	`
	row := db.QueryRow(query, className)

	var class Class
	err := row.Scan(&class.ID, &class.Name, &class.CreateTS, &class.UpdateTS, &class.DeleteTS)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("class not found")
		}
		return nil, fmt.Errorf("failed to scan class: %v", err)
	}
	class.db = db
	return &class, nil
}

func GetClassById(db *sql.DB, id int) (*Class, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts
		FROM game.dim_class
		WHERE id = ($1)
	`
	row := db.QueryRow(query, id)

	var class Class
	err := row.Scan(&class.ID, &class.Name, &class.CreateTS, &class.UpdateTS, &class.DeleteTS)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("class not found")
		}
		return nil, fmt.Errorf("failed to scan class: %v", err)
	}
	class.db = db
	return &class, nil
}

func (c *Class) GetAttributeFormulaById(attributeID int) (string, error) {
	query := `
		SELECT formula
		FROM game.bridge_class_attribute_formula
		WHERE class_id = $1
		AND attribute_id = $2
		AND is_deleted = false
	`
	var formula string
	err := c.db.QueryRow(query, c.ID, attributeID).Scan(&formula)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No formula found
		}
		return "", fmt.Errorf("failed to get class attribute formula: %v", err)
	}

	return formula, nil
}

// Get class atribute formula by ID or Name
func (c *Class) GetAttributeFormulaByName(attName string) (*calculation.Formula, error) {
	var attFormula = &calculation.Formula{}

	att, err := GetAttributeByName(c.db, attName)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT formula
		FROM game.bridge_class_attribute_formula
		WHERE class_id = $1
		AND attribute_id = $2
		AND is_deleted = false
	`

	err = c.db.QueryRow(query, c.ID, att.ID).Scan(&attFormula.Expression)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("No formula found") // No formula found
		}
		return nil, fmt.Errorf("failed to get class attribute formula: %v", err)
	}

	return attFormula, nil
}
