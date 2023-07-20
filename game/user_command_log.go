package game

import (
	"database/sql"
	"fmt"
	"time"
)

type CommandLogEntry struct {
	Command     string
	Args        string
	LogTS       time.Time
	CharacterID int
}

type UserCommandLog struct {
	DB *sql.DB
}

func NewUserCommandLog(db *sql.DB) *UserCommandLog {
	return &UserCommandLog{
		DB: db,
	}
}

func (u *UserCommandLog) LogCommand(userID int64, command string, args string) error {
	query := `
		INSERT INTO game.fact_user_command_log (tg_user_id, command, args, character_id)
		VALUES ($1, $2, $3, (SELECT MAX(character_id)
			                FROM game.bridge_tg_user_character 
							WHERE user_id = $1 
							AND is_deleted = False))
	`
	_, err := u.DB.Exec(query, userID, command, args)
	if err != nil {
		return fmt.Errorf("failed to log command with userId %d, command %v, args%v. Error: %v", userID, command, args, err)
	}
	return nil
}

func (u *UserCommandLog) GetCommandHistory(userID int64) ([]*CommandLogEntry, error) {
	query := `
		SELECT command, 
		       args, 
			   log_ts, 
			   character_id
		FROM game.fact_user_command_log
		WHERE tg_user_id = $1
		ORDER BY log_ts DESC
	`
	rows, err := u.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command history: %v", err)
	}
	defer rows.Close()

	var entries []*CommandLogEntry
	for rows.Next() {
		entry := &CommandLogEntry{}
		err := rows.Scan(&entry.Command, &entry.Args, &entry.LogTS, &entry.CharacterID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan command history entry: %v", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in command history rows: %v", err)
	}

	return entries, nil
}

func (u *UserCommandLog) GetUserPreviousCommand(userID int64) (string, error) {
	query := `
		SELECT command, args
		FROM (
		SELECT command, args, ROW_NUMBER() OVER (PARTITION BY tg_user_id ORDER BY log_ts DESC) rn
		FROM game.fact_user_command_log
		WHERE tg_user_id = $1
		AND command <> '/prev'
		AND log_ts >= current_date) t 
		WHERE rn = 2
	`
	var command, args string
	err := u.DB.QueryRow(query, userID).Scan(&command, &args)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No prev
		}
		return "", fmt.Errorf("failed to get user previous command: %v", err)
	}

	if args != "" {
		return fmt.Sprintf("%s %s", command, args), nil
	}

	return command, nil
}
