package game

import (
	"database/sql"
	"fmt"

	"github.com/Adelphopoet/dnd-bot-app/game/weapon"
)

type Item struct {
	ID     int
	Weapon *weapon.Weapon
	Name   string
	db     *sql.DB
}

func GetItemById(db *sql.DB, itemID int) (*Item, error) {
	query := `
	select id, name, weapon_id
	from game.dim_item
	where id = $1
	`
	var id int
	var name string
	var weaponID *int

	err := db.QueryRow(query, itemID).Scan(&id, &name, &weaponID)

	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %v", err)
	}

	item := &Item{ID: id, Name: name, db: db}

	if weaponID != nil {
		weapon, err := weapon.GetWeaponById(db, *weaponID)
		if err != nil {
			return nil, err
		} else {
			item.Weapon = weapon
		}
	}
	return item, nil
}
