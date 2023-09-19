package action

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
)

type Action struct {
	Name          string
	CreateTS      time.Time
	UpdateTS      time.Time
	DeleteTS      *time.Time
	IsDeleted     bool
	ID            int
	ActionType    *ActionType
	ActionFormula *ActionFormula
	db            *sql.DB
	Properties    []*ActionProperty
}

func GetActionByIDorName(db *sql.DB, identifier interface{}) (*Action, error) {
	var query string
	var args []interface{}
	switch identifier := identifier.(type) {
	case int:
		query = `
			SELECT 
				name,
				create_ts,
				update_ts,
				delete_ts,
				is_deleted,
				id,
				action_type_id
			FROM game.dim_action
			WHERE id = $1
			AND is_deleted = False
		`
		args = []interface{}{identifier}
	case string:
		query = `
			SELECT 
				name,
				create_ts,
				update_ts,
				delete_ts,
				is_deleted,
				id,
				action_type_id
			FROM game.dim_action
			WHERE LOWER(name) = LOWER($1)
			AND is_deleted = False
		`
		args = []interface{}{identifier}
	default:
		return nil, fmt.Errorf("invalid identifier type")
	}
	var action Action
	var actionTypeID int
	err := db.QueryRow(query, args...).Scan(&action.Name, &action.CreateTS, &action.UpdateTS, &action.DeleteTS, &action.IsDeleted, &action.ID, &actionTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get action by ID: %v", err)
	}

	actionType, err := GetActionTypeByID(db, actionTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actionType by ID: %v", err)
	}
	action.ActionType = actionType
	action.db = db

	_, err = action.GetActionFormula()
	if err != nil {
		return nil, fmt.Errorf("failed to get ActionFormula: %v", err)
	}

	_, err = action.GetActionProperties()
	if err != nil {
		return nil, fmt.Errorf("failed to get ActionProperties: %v", err)
	}

	return &action, nil
}

func (a *Action) GetActionFormula() (*ActionFormula, error) {
	fmt.Printf("\n!!!Start check formula by actionID: %d", a.ID)
	query := `
	SELECT
		hit_formula,
		damage_formula,
		base_formula
	FROM game.bridge_action_formula
	WHERE action_id = $1
	`

	hitFormula := &calculation.Formula{}
	dmgFormula := &calculation.Formula{}
	baseFormula := &calculation.Formula{}
	err := a.db.QueryRow(query, a.ID).Scan(&hitFormula.Expression, &dmgFormula.Expression, &baseFormula.Expression)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No formula
		}
		return nil, fmt.Errorf("failed to get action formula: %v", err)
	}

	a.ActionFormula = &ActionFormula{DamageFormula: dmgFormula, HitFormula: hitFormula, BaseFormula: baseFormula}
	return a.ActionFormula, nil
}

func (a *Action) GetActionProperties() ([]*ActionProperty, error) {
	query := `
	SELECT 
		property_id
	FROM game.bridge_action_property_value
	WHERE 
		action_id = $1
	`

	rows, err := a.db.Query(query, a.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to scan character class: %v", err)
	}
	defer rows.Close()

	var properties []*ActionProperty

	for rows.Next() {
		var propertyID int
		err := rows.Scan(&propertyID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan propertyID: %v", err)
		}

		property, err := GetActionPropertyByIDorName(a.db, propertyID)
		if err != nil {
			return nil, fmt.Errorf("failed get property by id: %v", err)
		}

		_, err = property.GetActionPropertyValue(a.ID)
		if err != nil {
			return nil, fmt.Errorf("failed get property value: %v", err)
		}
		properties = append(properties, property)
	}
	a.Properties = properties
	return properties, nil
}

// Get current action property value by it's name
func (a *Action) GetValueByPropertyName(propertyName string) (*ActionProperty, error) {
	for _, property := range a.Properties {
		if property.Name == propertyName {
			_, _ = property.GetActionPropertyValue(a.ID)
			return property, nil
		}
	}
	return nil, nil
}

// Filtring actions by user data where keys = property name and values = property value
func GetActionsByFilter(db *sql.DB, filter map[string]interface{}) ([]*Action, error) {
	// Создаем SQL-запрос и массив параметров
	query := `
	SELECT 
		action_id
	FROM 
		game.bridge_action_property_value pv
	JOIN 
		game.dim_property p
		ON p.id = pv.property_id
	WHERE
	`
	params := []interface{}{}
	conditions := []string{}

	// Check dict and generate SQL
	for columnName, filterValue := range filter {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", columnName, len(params)+1))
		params = append(params, filterValue)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no filter values provided")
	}

	query += " " + strings.Join(conditions, " AND ")

	// Execute SQL with user filtres
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var actions []*Action
	for rows.Next() {
		var actionID int
		err := rows.Scan(&actionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan actionID: %v", err)
		}

		act, err := GetActionByIDorName(db, actionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get action: %v", err)
		}
		actions = append(actions, act)
	}

	return actions, nil
}
