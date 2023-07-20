package fighting

import (
	"fmt"

	"github.com/Adelphopoet/dnd-bot-app/game"
)

type Initiative struct {
	RollInitiative int
	DexBonus       int
}

// Get character initiative to fight turns
func GetCharacterInitiative(character *game.Character) (*Initiative, error) {
	// Roll for initiative
	baseFormula := &game.Formula{Expression: "d20"}
	rollInitiative, err := game.CalculateFormula(baseFormula)
	if err != nil {
		return nil, fmt.Errorf("Error during rolling initiative: %v", err)
	}

	// Get DEX bonus to resolve equal rolls conflicts
	dexAtt := character.GetAttributeValueByName("dex")
	dexBonus, err := game.GetAttributeBonus(dexAtt)
	if err != nil {
		return nil, fmt.Errorf("Error during rolling initiative: %v", err)
	}
	return &Initiative{RollInitiative: rollInitiative, DexBonus: dexBonus}, nil
}
