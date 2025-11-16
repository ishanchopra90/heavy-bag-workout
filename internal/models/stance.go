package models

// Stance represents a boxing stance
type Stance int

const (
	Orthodox Stance = iota // Right-handed boxer (left foot forward)
	Southpaw               // Left-handed boxer (right foot forward)
)

// String returns the string representation of the stance
func (s Stance) String() string {
	switch s {
	case Orthodox:
		return "orthodox"
	case Southpaw:
		return "southpaw"
	default:
		return "unknown"
	}
}
