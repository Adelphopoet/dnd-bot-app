package fighting

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Adelphopoet/dnd-bot-app/game"
	"github.com/Adelphopoet/dnd-bot-app/game/action"
	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
)

type Round struct {
	ID   int
	Rows []*RoundRow
	db   *sql.DB
}

type RoundRow struct {
	ID        int
	Character *game.Character
	TurnRound int
	db        *sql.DB
	Action    *action.Action
	Details   json.RawMessage
	IsEnded   bool
}

func (rr *RoundRow) Update() error {
	queryUpdateRound := `
	UPDATE game.fact_fight_round_row
	SET
		round_id = $1,
		character_id = $2,
		turn =  = $3,
		action_id = $4,
		action_details = $5,
		update_ts = DEFAULT,
		is_ended = $6
	WHERE 
		id = $6
	RETURNING id
	`
	_, err := rr.db.Exec(queryUpdateRound, rr.ID, rr.Character.ID, rr.TurnRound, rr.Action.ID, rr.Details, rr.IsEnded)
	if err != nil {
		return fmt.Errorf("failed to update fight round row: %v", err)
	}
	return nil
}

func (r *Round) NewRoundRow(fightID int, character *game.Character, turn int) (*RoundRow, error) {
	queryNewRound := `
	INSERT INTO game.fact_fight_round_row (
		round_id,
		character_id,
		turn
	) VALUES ($1, $2, $3)
	RETURNING id
	`
	var roundRowID int
	err := r.db.QueryRow(queryNewRound, r.ID, character.ID, turn).Scan(&roundRowID)
	if err != nil {
		return nil, fmt.Errorf("Error during scan round row id: %v", err)
	}
	row := &RoundRow{ID: roundRowID, Character: character, db: r.db, TurnRound: turn, IsEnded: false}
	r.Rows = append(r.Rows, row)
	return row, nil
}

func (r *Round) GetRoundRows() ([]*RoundRow, error) {
	query := `
	SELECT 
		id, 
		character_id, 
		turn,
		action_id,
		action_details,
		is_ended
		--create_ts,
		--update_ts,
		--delete_ts
	FROM game.fact_fight_round_row
	WHERE round_id = $1
	`
	fmt.Println(query, r.ID)
	rows, err := r.db.Query(query, r.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query round: %v", err)
	}
	defer rows.Close()

	var roundRows []*RoundRow

	for rows.Next() {
		var id, characterID, turn int

		var actionID *int // to avaid sql NULL use *
		var act *action.Action
		var details sql.NullString
		var detailsJson json.RawMessage
		var isEnded bool

		err := rows.Scan(&id, &characterID, &turn, &actionID, &details, &isEnded)

		if err != nil {
			return nil, fmt.Errorf("failed to scan round rows: %v", err)
		}

		if details.Valid {
			err = json.Unmarshal([]byte(details.String), &detailsJson)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON details: %v", err)
			}
		}

		if actionID != nil {
			act, err = action.GetActionByIDorName(r.db, actionID)
			if err != nil {
				return nil, fmt.Errorf("failed to get action: %v", err)
			}
		}

		character, err := game.GetCharacterByID(r.db, characterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get character: %v", err)
		}
		roundRows = append(roundRows, &RoundRow{ID: id, Character: character, TurnRound: turn, Action: act, Details: detailsJson, IsEnded: isEnded})
	}
	r.Rows = roundRows

	return roundRows, nil
}

// Return characters whos turn now
func (f *Fight) WhoseTurn() (*Turn, error) {
	if len(f.Turns) == 0 || f.Turns == nil {
		return nil, fmt.Errorf("no turns available")
	}

	minTurnNum := f.Turns[0].Turn
	minTurn := f.Turns[0]

	for _, turn := range f.Turns {
		if turn.Turn < minTurnNum {
			minTurnNum = turn.Turn
			minTurn = turn
		}
	}

	return minTurn, nil
}

func GetRoundsByFightID(fightID int, db *sql.DB) ([]*Round, error) {
	query := `
	select id
	from game.fact_fight_round
	where fight_id  = $1
	and is_deleted = false
	`

	rows, err := db.Query(query, fightID)
	if err != nil {
		return nil, fmt.Errorf("failed to query rounds by fight: %v", err)
	}
	defer rows.Close()

	var rounds []*Round
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rounds: %v", err)
		}

		round := &Round{ID: id, db: db}
		_, err = round.GetRoundRows()
		if err != nil {
			return nil, fmt.Errorf("Can't get round rows: %v", err)
		}
		rounds = append(rounds, round)
	}
	return rounds, nil
}

func (r *Round) GetLastActiveRoundRow() (*RoundRow, error) {
	if r.Rows == nil {
		return nil, nil
	}
	if len(r.Rows) == 0 {
		return nil, nil
	}

	minRoundRowTurn := -1
	var minRoundRow *RoundRow

	for _, row := range r.Rows {
		if (row.TurnRound < minRoundRowTurn || minRoundRowTurn == -1) && row.IsEnded != true {
			minRoundRowTurn = row.TurnRound
			minRoundRow = row
		}
	}

	if minRoundRowTurn != -1 {
		return minRoundRow, nil
	} else {
		return nil, nil
	}

}

func (rr *RoundRow) DoAttack(opponentCharacter *game.Character, act action.Action) error {
	//Get weapon dmg
	var itemDmgFormula *calculation.Formula
	if act.Name == "Атака оружием" {
		acitveWeapon, err := rr.Character.GetActiveWeapon()
		if err != nil {
			return fmt.Errorf("Error during get active weapin: %v", err)
		}

		itemDmgFormula = acitveWeapon.DamageFormula

		weaponAttr := acitveWeapon.WeaponType.MainCharacteristic

		attackRoll, err := rr.Character.Roll(&calculation.Formula{Expression: fmt.Sprintf("d20 + %v", weaponAttr)})
		if err != nil {
			return fmt.Errorf("Can't roll attack: %v", err)
		}

		if attackRoll >= 12 { // Допилить получение брони
			damageRoll, err := rr.Character.Roll(itemDmgFormula)
			if err != nil {
				return fmt.Errorf("Can't roll attack: %v", err)
			} else {
				fmt.Printf("%d", damageRoll)
			}
		}

	}
	return nil
}
