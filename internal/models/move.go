package models

// MoveType represents the type of move (punch or defensive)
type MoveType int

const (
	MoveTypePunch MoveType = iota
	MoveTypeDefensive
)

// Move represents either a punch or a defensive move
// TODO: add other types of moves like planks, step in/outs etc.
type Move struct {
	Type      MoveType
	Punch     *Punch
	Defensive *DefensiveMove
}

// NewPunchMove creates a new Move from a Punch
func NewPunchMove(punch Punch) Move {
	return Move{
		Type:  MoveTypePunch,
		Punch: &punch,
	}
}

// NewDefensiveMove creates a new Move from a DefensiveMove
func NewDefensiveMove(defensive DefensiveMove) Move {
	return Move{
		Type:      MoveTypeDefensive,
		Defensive: &defensive,
	}
}

// String returns the string representation of a move
func (m Move) String() string {
	switch m.Type {
	case MoveTypePunch:
		if m.Punch != nil {
			return m.Punch.String()
		}
	case MoveTypeDefensive:
		if m.Defensive != nil {
			return m.Defensive.String()
		}
	}
	return "Unknown"
}

// IsPunch returns true if the move is a punch
func (m Move) IsPunch() bool {
	return m.Type == MoveTypePunch
}

// IsDefensive returns true if the move is a defensive move
func (m Move) IsDefensive() bool {
	return m.Type == MoveTypeDefensive
}
