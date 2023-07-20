package action

import (
	"github.com/Adelphopoet/dnd-bot-app/game"
)

type ActionFormula struct {
	HitFormula    *game.Formula
	DamageFormula *game.Formula
	BaseFormula   *game.Formula
}
