package action

import (
	"database/sql"
	"fmt"
	"time"

	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
)

type ActionProperty struct {
	ID          int
	Name        string
	Description string
	CreateTS    time.Time
	UpdateTS    time.Time
	DeleteTS    *time.Time
	IsDeleted   bool
	db          *sql.DB
	Values      *PropertyValue
}

type PropertyValue struct {
	StringValue  string
	NumericValue int
	BoolValue    bool
	FormulaValue *calculation.Formula
}

func (p *ActionProperty) GetActionPropertyValue(actionID int) (*PropertyValue, error) {
	query := `
		SELECT
			string_value,
			formula_value,
			bool_value,
			numeric_value
		FROM game.bridge_action_property_value
		WHERE action_id = $1
		AND property_id = $2
	`
	value := &PropertyValue{FormulaValue: &calculation.Formula{}}

	err := p.db.QueryRow(query, actionID, p.ID).Scan(
		&value.StringValue,
		&value.FormulaValue.Expression,
		&value.BoolValue,
		&value.NumericValue,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Property Value not found
		}
		return nil, fmt.Errorf("failed to get action property value: %v", err)
	}
	p.Values = value
	return value, nil

}

func GetActionPropertyByIDorName(db *sql.DB, identifier interface{}) (*ActionProperty, error) {
	var args []interface{}
	var query string
	switch identifier := identifier.(type) {
	case int:
		query = `
			SELECT
				id,
				"name",
				create_ts,
				update_ts,
				delete_ts,
				is_deleted,
				description
			FROM game.dim_property
			WHERE is_deleted = False
			AND id = $1
			`
		args = []interface{}{identifier}
	case string:
		query = `
			SELECT
				id,
				"name",
				create_ts,
				update_ts,
				delete_ts,
				is_deleted,
				description
			FROM game.dim_property
			WHERE is_deleted = False
			AND LOWER("name") = LOWER($1)
			`
		args = []interface{}{identifier}
	default:
		return nil, fmt.Errorf("invalid identifier type")
	}

	property := &ActionProperty{}
	err := db.QueryRow(query, args...).Scan(
		&property.ID,
		&property.Name,
		&property.CreateTS,
		&property.UpdateTS,
		&property.DeleteTS,
		&property.IsDeleted,
		&property.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Property not found
		}
		return nil, fmt.Errorf("failed to get action property by ID: %v", err)
	}

	property.db = db
	return property, nil
}
