package models

import (
	"fmt"
	"strings"
)

// MoveMapping provides numeric mappings for punches and defensive moves
// Punches: 1-6, Defensive Moves: 7-12
type MoveMapping struct {
	PunchMappings       map[int]Punch
	DefensiveMappings   map[int]DefensiveMove
	ReversePunchMap     map[Punch]int
	ReverseDefensiveMap map[DefensiveMove]int
}

// NewMoveMapping creates a new move mapping with standard mappings
func NewMoveMapping() MoveMapping {
	pm := make(map[int]Punch)
	dm := make(map[int]DefensiveMove)
	rpm := make(map[Punch]int)
	rdm := make(map[DefensiveMove]int)

	// Punches: 1-6
	punches := AllPunches()
	for i, punch := range punches {
		num := i + 1
		pm[num] = punch
		rpm[punch] = num
	}

	// Defensive Moves: 7-12
	defensiveMoves := AllDefensiveMoves()
	for i, move := range defensiveMoves {
		num := i + 7
		dm[num] = move
		rdm[move] = num
	}

	return MoveMapping{
		PunchMappings:       pm,
		DefensiveMappings:   dm,
		ReversePunchMap:     rpm,
		ReverseDefensiveMap: rdm,
	}
}

// GetMoveFromNumber converts a numeric move (1-12) to a Move
func (mm MoveMapping) GetMoveFromNumber(num int) (Move, bool) {
	if punch, ok := mm.PunchMappings[num]; ok {
		return NewPunchMove(punch), true
	}
	if defensive, ok := mm.DefensiveMappings[num]; ok {
		return NewDefensiveMove(defensive), true
	}
	return Move{}, false
}

// GetNumberFromMove converts a Move to its numeric representation
func (mm MoveMapping) GetNumberFromMove(move Move) (int, bool) {
	if move.IsPunch() && move.Punch != nil {
		if num, ok := mm.ReversePunchMap[*move.Punch]; ok {
			return num, true
		}
	}
	if move.IsDefensive() && move.Defensive != nil {
		if num, ok := mm.ReverseDefensiveMap[*move.Defensive]; ok {
			return num, true
		}
	}
	return 0, false
}

// GetMappingDescription returns a string description of all mappings for LLM context
func (mm MoveMapping) GetMappingDescription() string {
	desc := "Punch Mappings:\n"
	for num, punch := range mm.PunchMappings {
		desc += fmt.Sprintf("  %d = %s\n", num, punch.String())
	}
	desc += "\nDefensive Move Mappings:\n"
	for num, move := range mm.DefensiveMappings {
		desc += fmt.Sprintf("  %d = %s\n", num, move.String())
	}
	return desc
}

// GetMappingDescriptionWithStance returns a string description of all mappings with stance-specific punch names
func (mm MoveMapping) GetMappingDescriptionWithStance(stance Stance) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Boxer Stance: %s\n", stance.String()))
	sb.WriteString("\n")

	sb.WriteString("Punch Mappings (with stance-specific names):\n")
	for num, punch := range mm.PunchMappings {
		stanceName := punch.NameForStance(stance)
		sb.WriteString(fmt.Sprintf("  %d = %s (technical: %s)\n", num, stanceName, punch.String()))
	}

	sb.WriteString("\nDefensive Move Mappings:\n")
	for num, move := range mm.DefensiveMappings {
		sb.WriteString(fmt.Sprintf("  %d = %s\n", num, move.String()))
	}

	return sb.String()
}
