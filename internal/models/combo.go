package models

import "fmt"

// Combo represents a sequence of punches and/or defensive moves
type Combo struct {
	Moves []Move
}

// NewCombo creates a new combo with the given moves
func NewCombo(moves []Move) Combo {
	return Combo{
		Moves: moves,
	}
}

// String returns the string representation of a combo
// Format: "1-2, Left Slip, 3-4" or "1-5-4, Duck"
func (c Combo) String() string {
	if len(c.Moves) == 0 {
		return ""
	}

	result := ""
	for i, move := range c.Moves {
		if i > 0 {
			result += ", "
		}

		if move.IsPunch() && move.Punch != nil {
			// For punches, show the number (1-6)
			result += fmt.Sprintf("%d", int(*move.Punch))
		} else {
			// For defensive moves, show the full name
			result += move.String()
		}
	}

	return result
}

// Length returns the number of moves in the combo
func (c Combo) Length() int {
	return len(c.Moves)
}

// IsEmpty returns true if the combo has no moves
func (c Combo) IsEmpty() bool {
	return len(c.Moves) == 0
}
