package weapon

import (
	"database/sql"
	"fmt"

	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
)

type Weapon struct {
	ID            int
	db            *sql.DB
	WeaponType    *WeaponType
	DamageFormula *calculation.Formula
	weight        int
}

type WeaponType struct {
	db                 *sql.DB
	ID                 int
	Name               string
	MainCharacteristic string
}

func GetWeaponById(db *sql.DB, weaponID int) (*Weapon, error) {
	query := `
	select id, name, cost, damage_formula, weight, weapon_type_id, damage_type_id
	from game.dim_weapon
	where id = $1
	`
	var id, cost, weight, weaponTypeId, damageTypeID int
	var name, damageFormula string

	err := db.QueryRow(query, weaponID).Scan(&id, &name, &cost, &damageFormula, &weight, &weaponTypeId, &damageTypeID)

	if err != nil {
		return nil, fmt.Errorf("failed to scan weapon: %v", err)
	}

	weaponType, err := GetWeaponTypeById(db, weaponTypeId)
	if err != nil {
		return nil, fmt.Errorf("Can't get weapon type by id: %v", err)
	}

	weapon := &Weapon{ID: id, db: db, weight: weight, WeaponType: weaponType, DamageFormula: &calculation.Formula{Expression: damageFormula}}

	return weapon, nil
}

func GetWeaponTypeById(db *sql.DB, weaponID int) (*WeaponType, error) {
	query := `
	select id, name, main_characteristic
	from game.dim_weapon_type
	where id = $1
	`
	var id int
	var name, MainCharacteristic string

	err := db.QueryRow(query, weaponID).Scan(&id, &name, &MainCharacteristic)

	if err != nil {
		return nil, fmt.Errorf("failed to scan weapon: %v", err)
	}

	weaponType := &WeaponType{ID: id, db: db, Name: name, MainCharacteristic: MainCharacteristic}
	return weaponType, nil
}
