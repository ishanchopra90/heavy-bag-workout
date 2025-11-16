package models

// DefensiveMove represents a defensive boxing move
type DefensiveMove int

const (
	LeftSlip DefensiveMove = iota
	RightSlip
	LeftRoll
	RightRoll
	PullBack
	Duck
)

// String returns the string representation of a defensive move
func (d DefensiveMove) String() string {
	switch d {
	case LeftSlip:
		return "Left Slip"
	case RightSlip:
		return "Right Slip"
	case LeftRoll:
		return "Left Roll"
	case RightRoll:
		return "Right Roll"
	case PullBack:
		return "Pull Back"
	case Duck:
		return "Duck"
	default:
		return "Unknown"
	}
}

// AllDefensiveMoves returns a slice of all available defensive moves
func AllDefensiveMoves() []DefensiveMove {
	return []DefensiveMove{LeftSlip, RightSlip, LeftRoll, RightRoll, PullBack, Duck}
}
