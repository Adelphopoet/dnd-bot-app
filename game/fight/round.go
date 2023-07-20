package fighting

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Adelphopoet/dnd-bot-app/game"
	"github.com/Adelphopoet/dnd-bot-app/game/action"
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
		update_ts = DEFAULT
	WHERE 
		id = $6
	RETURNING id
	`
	_, err := rr.db.Exec(queryUpdateRound, rr.ID, rr.Character.ID, rr.TurnRound, rr.Action.ID, rr.Details)
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
	row := &RoundRow{ID: roundRowID, Character: character, db: r.db, TurnRound: turn}
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
		action_details--,
		--create_ts,
		--update_ts,
		--delete_ts
	FROM game.fact_fight_round_row
	WHERE round_id = $1
	`
	rows, err := r.db.Query(query, r.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query round: %v", err)
	}
	defer rows.Close()

	var roundRows []*RoundRow

	for rows.Next() {
		var id int
		var characterID int
		var turn int

		var actionID *int // to avaid sql NULL use *
		var act *action.Action
		var details sql.NullString
		var detailsJson json.RawMessage

		err := rows.Scan(&id, &characterID, &turn, &actionID, &details)

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
		roundRows = append(roundRows, &RoundRow{ID: id, Character: character, TurnRound: turn, Action: act, Details: detailsJson})
	}

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
