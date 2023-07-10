package game

import (
	"database/sql"
	"fmt"
	"time"
)

type Location struct {
	ID        int
	Name      string
	CreateTS  time.Time
	UpdateTS  time.Time
	DeleteTS  sql.NullTime
	IsDeleted bool
	db        *sql.DB
}

func NewLocation(name string, db *sql.DB) *Location {
	return &Location{
		Name: name,
		db:   db,
	}
}

func (l *Location) Save() error {
	query := `
		INSERT INTO game.dim_location ("name")
		VALUES ($1)
		RETURNING id, create_ts, update_ts, delete_ts
	`
	err := l.db.QueryRow(query, l.Name).Scan(&l.ID, &l.CreateTS, &l.UpdateTS, &l.DeleteTS)
	if err != nil {
		return fmt.Errorf("failed to save location: %v", err)
	}
	return nil
}

func GetLocationByName(db *sql.DB, name string) (*Location, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_location
		WHERE "name" = $1
		AND is_deleted = False
	`
	location := &Location{}
	err := db.QueryRow(query, name).Scan(&location.ID, &location.Name, &location.CreateTS, &location.UpdateTS, &location.DeleteTS, &location.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Location not found
		}
		return nil, fmt.Errorf("failed to get location by name: %v", err)
	}
	return location, nil
}

func GetLocationByID(db *sql.DB, id int) (*Location, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_location
		WHERE id = $1
		AND is_deleted = False
	`
	location := &Location{}
	err := db.QueryRow(query, id).Scan(&location.ID, &location.Name, &location.CreateTS, &location.UpdateTS, &location.DeleteTS, &location.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Location not found
		}
		return nil, fmt.Errorf("failed to get location by ID: %v", err)
	}
	location.db = db
	return location, nil
}

func GetAllLocations(db *sql.DB) ([]Location, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_location
		WHERE is_deleted = False
		ORDER BY update_ts ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %v", err)
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var location Location
		err := rows.Scan(&location.ID, &location.Name, &location.CreateTS, &location.UpdateTS, &location.DeleteTS, &location.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan location: %v", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in rows: %v", err)
	}

	return locations, nil
}

func (l *Location) GetAvailablePathes() ([]*Location, error) {
	query := `
		SELECT lc.location_to_id, dl."name", dl.create_ts, dl.update_ts, dl.delete_ts, dl.is_deleted
		FROM game.bridge_location_path AS lc
		INNER JOIN game.dim_location AS dl ON lc.location_to_id = dl.id
		WHERE lc.location_from_id = $1
		AND lc.is_deleted = false
		AND dl.is_deleted = false
	`
	var pathes []*Location
	rows, err := l.db.Query(query, l.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available pathes: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		path := &Location{}
		err := rows.Scan(&path.ID, &path.Name, &path.CreateTS, &path.UpdateTS, &path.DeleteTS, &path.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan path: %v", err)
		}
		pathes = append(pathes, path)
	}

	return pathes, nil
}
