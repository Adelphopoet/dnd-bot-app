package fighting

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/Adelphopoet/dnd-bot-app/game"
)

type Fight struct {
	Characters []*game.Character
	ID         int
	CreateTS   time.Time
	UpdateTS   time.Time
	DeleteTS   sql.NullTime
	IsEnded    bool
	db         *sql.DB
	Turns      []*Turn
}

func GetFightByID(db *sql.DB, id int) (*Fight, error) {
	query := `
	SELECT
		create_ts,
		update_ts,
		delete_ts,
		id,
		is_ended
	FROM game.fact_fight
	WHERE id = $1
	AND is_deleted = False
	`
	fight := &Fight{}
	err := db.QueryRow(query, id).Scan(&fight.CreateTS, &fight.UpdateTS, &fight.DeleteTS, &fight.ID, &fight.IsEnded)

	if err != nil {
		return nil, fmt.Errorf("error to get fight: %v", err)
	}

	fight.db = db

	// Add characters into struct
	_, _, err = fight.GetFightCharacters()
	if err != nil {
		return nil, fmt.Errorf("error to get characters: %v", err)
	}

	return fight, nil
}

func (f *Fight) GetFightCharacters() ([]*game.Character, []*Turn, error) {
	query := `
		SELECT character_id, turn
		FROM game.bridge_fight_character
		WHERE fight_id = $1
	`
	rows, err := f.db.Query(query, f.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get fight characters by fight ID: %v", err)
	}
	defer rows.Close()

	var fightCharacters []*game.Character
	var fightTurns []*Turn
	for rows.Next() {
		var characterId int
		var turn int

		err := rows.Scan(&characterId, &turn)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to scan fight character: %v", err)
		}

		character, err := game.GetCharacterByID(f.db, characterId)
		if err != nil {
			return nil, nil, fmt.Errorf("Can't get charater by ID: %v", err)
		}

		fightTurn := &Turn{Character: character, Turn: turn}
		fightTurns = append(fightTurns, fightTurn)
		fightCharacters = append(fightCharacters, character)
	}

	// Тестим
	f.Turns = fightTurns
	return fightCharacters, fightTurns, nil
}

func NewFight(db *sql.DB, charaterFrom *game.Character, characterTo *game.Character) (*Fight, error) {
	//Chek fight already started
	sqlSearch := `
	select 
		id
	from (
		select 
			f.id, 
			count(1) cnt
		from game.bridge_fight_character fc
		join game.fact_fight f
			on f.id = fc.fight_id
		where 
			character_id in ($1, $2)
			and f.is_deleted = false 
			and f.is_ended = false
		group by 1
	) t
		where cnt >= 2
	limit 1
	`
	var existingFightId int
	rows, err := db.Query(sqlSearch, characterTo.ID, charaterFrom.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing fights: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&existingFightId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan existing fights: %v", err)
		}
		existingFight, err := GetFightByID(db, existingFightId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan get fight by id: %v", err)
		} else {
			return existingFight, nil // If we already have one not ended fight, return it
		}
	}

	// Setup turns
	// Rolling for initiative
	var turns []*Turn
	attackerInit, err := GetCharacterInitiative(charaterFrom)
	if err != nil {
		return nil, fmt.Errorf("Error during roll for init by attacker: %v", err)
	}
	defenderInit, err := GetCharacterInitiative(characterTo)
	if err != nil {
		return nil, fmt.Errorf("Error during roll for init by defender: %v", err)
	}

	var firstTurn, secondTurn *Turn
	switch {
	case attackerInit.RollInitiative > defenderInit.RollInitiative:
		firstTurn = &Turn{Character: charaterFrom, Turn: 1}
		secondTurn = &Turn{Character: characterTo, Turn: 2}
	case attackerInit.RollInitiative < defenderInit.RollInitiative:
		firstTurn = &Turn{Character: characterTo, Turn: 1}
		secondTurn = &Turn{Character: charaterFrom, Turn: 2}
	case attackerInit.DexBonus == defenderInit.DexBonus:
		randomNum := rand.Intn(2) // Random 0-1
		if randomNum == 0 {
			firstTurn = &Turn{Character: charaterFrom, Turn: 1}
			secondTurn = &Turn{Character: characterTo, Turn: 2}
		} else {
			firstTurn = &Turn{Character: characterTo, Turn: 1}
			secondTurn = &Turn{Character: charaterFrom, Turn: 2}
		}
	default:
		firstTurn = &Turn{Character: characterTo, Turn: 1}
		secondTurn = &Turn{Character: charaterFrom, Turn: 2}
	}
	turns = append(turns, firstTurn, secondTurn)

	fight := &Fight{
		Characters: []*game.Character{charaterFrom, characterTo},
		IsEnded:    false,
		Turns:      turns,
		db:         db,
	}
	err = fight.Save()
	if err != nil {
		return nil, fmt.Errorf("Can't save fight: %v", err)
	}

	return fight, nil
}

func (f *Fight) Save() error {
	tx, err := f.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	queryFight := `
		INSERT INTO game.fact_fight ("is_ended")
		VALUES ($1)
		RETURNING id, create_ts, update_ts, delete_ts, is_ended
	`
	err = tx.QueryRow(queryFight, f.IsEnded).Scan(&f.ID, &f.CreateTS, &f.UpdateTS, &f.DeleteTS, &f.IsEnded)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save fight: %v", err)
	}

	queryCharacters := `
		INSERT INTO game.bridge_fight_character (fight_id, character_id, turn)
		VALUES ($1, $2, $3)
		ON CONFLICT (fight_id, character_id) DO UPDATE
		SET turn = EXCLUDED.turn,
		update_ts = current_timestamp
	`
	for _, turn := range f.Turns {
		_, err = tx.Exec(queryCharacters, f.ID, turn.Character.ID, turn.Turn)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save characters in fight: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit save fight transaction: %v", err)
	}

	return nil
}

func (f *Fight) GetLastActiveRound() (*Round, error) {
	query := `
	SELECT cast(coalesce(max(fr.id), 0) as int) id
	FROM game.fact_fight_round_row frr
	JOIN game.fact_fight_round fr
	ON fr.id = frr.round_id
	WHERE 
	frr.action_id is null
	AND fr.fight_id = $1
	`
	var roundID int
	err := f.db.QueryRow(query, f.ID).Scan(&roundID)
	if err != nil {
		return nil, fmt.Errorf("Error during get max round id: %v", err)
	}

	// No round found
	if roundID == 0 {
		return nil, nil
	}

	round := &Round{ID: roundID, db: f.db}
	_, err = round.GetRoundRows()
	if err != nil {
		return nil, fmt.Errorf("Error during get round rows: %v", err)
	}

	return round, nil
}

func (f *Fight) CreateOrGetRound() (*Round, error) {
	// Check we have already active round
	activeRound, err := f.GetLastActiveRound()
	if err != nil {
		return nil, err
	}

	if activeRound != nil {
		_, err = activeRound.GetRoundRows()
		if err != nil {
			return nil, err
		}
		return activeRound, nil
	}

	// Create new round
	queryNewRound := `
	INSERT INTO game.fact_fight_round (fight_id)
	VALUES ($1)
	RETURNING id
	`
	var roundID int
	err = f.db.QueryRow(queryNewRound, f.ID).Scan(&roundID)
	if err != nil {
		return nil, fmt.Errorf("failed to scan fight round: %v", err)
	}

	round := &Round{
		ID: roundID,
		db: f.db,
	}

	for _, turn := range f.Turns {
		_, err = round.NewRoundRow(f.ID, turn.Character, turn.Turn)
		if err != nil {
			return nil, err
		}
	}
	return round, nil
}
