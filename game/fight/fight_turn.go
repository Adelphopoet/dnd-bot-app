package fighting

import (
	"github.com/Adelphopoet/dnd-bot-app/game"
)

type Turn struct {
	Character *game.Character
	Turn      int
}
