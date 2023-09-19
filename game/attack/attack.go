package attack

import (
	"database/sql"

	"github.com/Adelphopoet/dnd-bot-app/game"
	"github.com/Adelphopoet/dnd-bot-app/game/action"
)

type Attack struct {
	CharacterFrom *game.Character
	CharacterTo   *game.Character
	db            *sql.DB
	Action        *action.Action
	Item          *game.Item
}
