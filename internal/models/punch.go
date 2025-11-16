package models

// Punch represents a boxing punch type
type Punch int

const (
	Jab Punch = iota + 1
	Cross
	LeadHook
	RearHook
	LeadUppercut
	RearUppercut
)

// String returns the string representation of a punch
func (p Punch) String() string {
	switch p {
	case Jab:
		return "Jab"
	case Cross:
		return "Cross"
	case LeadHook:
		return "Lead Hook"
	case RearHook:
		return "Rear Hook"
	case LeadUppercut:
		return "Lead Uppercut"
	case RearUppercut:
		return "Rear Uppercut"
	default:
		return "Unknown"
	}
}

// AllPunches returns a slice of all available punches
func AllPunches() []Punch {
	return []Punch{Jab, Cross, LeadHook, RearHook, LeadUppercut, RearUppercut}
}

// NameForStance returns the punch name based on the boxer's stance
// For orthodox (right-handed): left is lead, right is rear
// For southpaw (left-handed): right is lead, left is rear
func (p Punch) NameForStance(stance Stance) string {
	switch p {
	case Jab:
		return "jab"
	case Cross:
		return "cross"
	case LeadHook:
		if stance == Southpaw {
			return "right hook"
		}
		return "left hook"
	case RearHook:
		if stance == Southpaw {
			return "left hook"
		}
		return "right hook"
	case LeadUppercut:
		if stance == Southpaw {
			return "right uppercut"
		}
		return "left uppercut"
	case RearUppercut:
		if stance == Southpaw {
			return "left uppercut"
		}
		return "right uppercut"
	default:
		return "unknown"
	}
}
