package game

import (
	"database/sql"
	"fmt"
)

type LvlInfo struct {
	Lvl int
	Exp int
	Pm  int
}

func GetLvlInfo(db *sql.DB, lvl int) (*LvlInfo, error) {
	query := `
		SELECT
			lvl,
			exp,
			pm
		FROM game.bridge_lvl_exp_pm
		WHERE lvl = $1
	`
	var lvlInfo = &LvlInfo{}
	rows, err := db.Query(query, lvl)

	if rows == nil {
		return nil, nil // no lvl found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get lvl info: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&lvlInfo.Lvl, &lvlInfo.Exp, &lvlInfo.Pm)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lvl %v info: %v", lvl, err)
		}
	}
	return lvlInfo, nil
}
