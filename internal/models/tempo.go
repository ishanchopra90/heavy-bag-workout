package models

import (
	"strings"
	"time"
)

// Tempo represents the workout tempo (speed of combo intervals)
type Tempo int

const (
	TempoSlow      Tempo = iota // 5 seconds between beeps
	TempoMedium                 // 4 seconds between beeps
	TempoFast                   // 3 seconds between beeps
	TempoSuperfast              // 2 seconds between beeps
	TempoUnknown                // Invalid/unknown tempo
)

// String returns the string representation of the tempo
func (t Tempo) String() string {
	switch t {
	case TempoSlow:
		return "slow"
	case TempoMedium:
		return "medium"
	case TempoFast:
		return "fast"
	case TempoSuperfast:
		return "superfast"
	case TempoUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// DisplayName returns the display name for the tempo (capitalized)
func (t Tempo) DisplayName() string {
	switch t {
	case TempoSlow:
		return "Slow"
	case TempoMedium:
		return "Medium"
	case TempoFast:
		return "Fast"
	case TempoSuperfast:
		return "Superfast"
	case TempoUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// Duration returns the time.Duration for the tempo
func (t Tempo) Duration() time.Duration {
	switch t {
	case TempoSlow:
		return 5 * time.Second
	case TempoMedium:
		return 4 * time.Second
	case TempoFast:
		return 3 * time.Second
	case TempoSuperfast:
		return 2 * time.Second
	case TempoUnknown:
		return 5 * time.Second // Default to slow for unknown
	default:
		return 5 * time.Second // Default to slow
	}
}

// ParseTempo parses a tempo string and returns the corresponding Tempo value
// Returns TempoUnknown if the string is invalid
// Empty string defaults to TempoSlow
func ParseTempo(s string) Tempo {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "slow", "":
		return TempoSlow
	case "medium":
		return TempoMedium
	case "fast":
		return TempoFast
	case "superfast":
		return TempoSuperfast
	default:
		return TempoUnknown // Invalid tempo
	}
}

// AllTempos returns a slice of all available tempo values
func AllTempos() []Tempo {
	return []Tempo{TempoSlow, TempoMedium, TempoFast, TempoSuperfast}
}

// MaxMovesLimit returns the maximum number of moves allowed per combo for this tempo
func (t Tempo) MaxMovesLimit() int {
	switch t {
	case TempoSlow:
		return 5
	case TempoMedium:
		return 4
	case TempoFast:
		return 3
	case TempoSuperfast:
		return 2
	default:
		return 5 // Default to slow tempo limit
	}
}
