package game

import (
	"database/sql"
	"fmt"
	"time"
)

type Attribute struct {
	ID          int
	Name        string
	Description string
	CreateTS    time.Time
	UpdateTS    time.Time
	DeleteTS    *time.Time
	IsDeleted   bool
}

type AttributeValue struct {
	StringValue  string
	NumericValue int
	BoolValue    bool
	FormulaValue *Formula
}

func GetAttributeByName(db *sql.DB, attributeName string) (*Attribute, error) {
	query := `
		SELECT id, name, description, create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_attribute
		WHERE name = $1
	`
	var attribute Attribute
	err := db.QueryRow(query, attributeName).Scan(&attribute.ID, &attribute.Name, &attribute.Description, &attribute.CreateTS, &attribute.UpdateTS, &attribute.DeleteTS, &attribute.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute by name: %v", err)
	}
	return &attribute, nil
}

func GetAttributeByID(db *sql.DB, attributeID int) (*Attribute, error) {
	query := `
		SELECT id, name, description, create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_attribute
		WHERE id = $1
	`
	var attribute Attribute
	err := db.QueryRow(query, attributeID).Scan(&attribute.ID, &attribute.Name, &attribute.Description, &attribute.CreateTS, &attribute.UpdateTS, &attribute.DeleteTS, &attribute.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute by ID: %v", err)
	}
	return &attribute, nil
}

func GetAttributeBonus(attibuteValue *AttributeValue) (int, error) {
	bonus := 0
	if attibuteValue == nil {
		return bonus, fmt.Errorf("Epty attribute value")
	}
	baseFormula := &Formula{Expression: fmt.Sprintf("(%d-10)/2", attibuteValue.NumericValue)}
	bonus, err := CalculateFormula(baseFormula)
	if err != nil {
		return bonus, err
	}
	return bonus, nil
}
