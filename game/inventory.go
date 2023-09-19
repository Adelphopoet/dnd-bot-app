package game

import (
	"database/sql"
	"fmt"
)

type InventoryItem struct {
	Item    *Item
	Quanty  int
	IsUsing bool
}

func GetInventaryItemsByCharacterID(db *sql.DB, characterID int) ([]*InventoryItem, error) {
	query := `
	select item_id, quanty, is_using
	from game.fact_inventary
	where character_id = $1
	`

	rows, err := db.Query(query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item list: %v", err)
	}
	defer rows.Close()

	var inventaryItems []*InventoryItem

	for rows.Next() {
		var itemId, quanty int
		var isUsing bool

		err := rows.Scan(&itemId, &quanty, &isUsing)
		if err != nil {
			return nil, fmt.Errorf("failed to scan items: %v", err)
		}

		item, err := GetItemById(db, itemId)
		if err != nil {
			return nil, fmt.Errorf("Can't get item: %v", err)
		}

		inventaryItem := &InventoryItem{Item: item, Quanty: quanty, IsUsing: isUsing}
		inventaryItems = append(inventaryItems, inventaryItem)
	}
	return inventaryItems, nil
}
