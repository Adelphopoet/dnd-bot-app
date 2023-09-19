package action

import (
	calculation "github.com/Adelphopoet/dnd-bot-app/game/claculation"
)

type ActionFormula struct {
	HitFormula    *calculation.Formula
	DamageFormula *calculation.Formula
	BaseFormula   *calculation.Formula
}
