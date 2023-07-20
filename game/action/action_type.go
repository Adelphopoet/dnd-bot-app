package action

import (
	"database/sql"
	"fmt"
	"time"
)

type ActionType struct {
	ID        int
	Name      string
	CreateTS  time.Time
	UpdateTS  time.Time
	DeleteTS  *time.Time
	IsDeleted bool
}

func GetActionTypeByID(db *sql.DB, id int) (*ActionType, error) {
	query := `
		SELECT id, name, create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_action_type
		WHERE id = $1
		AND is_deleted = False
	`
	var actionType ActionType
	err := db.QueryRow(query, id).Scan(&actionType.ID, &actionType.Name, &actionType.CreateTS, &actionType.UpdateTS, &actionType.DeleteTS, &actionType.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get action type by ID: %v", err)
	}
	return &actionType, nil
}

func GetActionTypeByName(db *sql.DB, name string) (*ActionType, error) {
	query := `
		SELECT id, name, create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_action_type
		WHERE name = $1
		AND is_deleted = False
	`
	var actionType ActionType
	err := db.QueryRow(query, name).Scan(&actionType.ID, &actionType.Name, &actionType.CreateTS, &actionType.UpdateTS, &actionType.DeleteTS, &actionType.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get action type by name: %v", err)
	}
	return &actionType, nil
}
