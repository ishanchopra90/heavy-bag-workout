package models

// WorkoutPatternType defines how combo complexity varies across rounds
type WorkoutPatternType string

const (
	PatternLinear   WorkoutPatternType = "linear"   // 1, 2, 3, 4, 5... (increasing)
	PatternPyramid  WorkoutPatternType = "pyramid"  // 1, 2, 3, 4, 3, 2, 1
	PatternRandom   WorkoutPatternType = "random"   // Random variation
	PatternConstant WorkoutPatternType = "constant" // Same complexity throughout
)

// WorkoutPattern defines how combos should vary across rounds
type WorkoutPattern struct {
	Type             WorkoutPatternType
	MinMoves         int  // Minimum moves per combo
	MaxMoves         int  // Maximum moves per combo
	IncludeDefensive bool // Whether to include defensive moves
}

// NewWorkoutPattern creates a new workout pattern
func NewWorkoutPattern(patternType WorkoutPatternType, minMoves, maxMoves int, includeDefensive bool) WorkoutPattern {
	return WorkoutPattern{
		Type:             patternType,
		MinMoves:         minMoves,
		MaxMoves:         maxMoves,
		IncludeDefensive: includeDefensive,
	}
}

// GetMovesPerRound calculates the number of moves for each round based on the pattern
func (wp WorkoutPattern) GetMovesPerRound(roundNumber, totalRounds int) int {
	switch wp.Type {
	case PatternLinear:
		// Linear progression: round 1 = min, round N = max
		if totalRounds == 1 {
			return wp.MinMoves
		}
		progress := float64(roundNumber-1) / float64(totalRounds-1)
		return wp.MinMoves + int(progress*float64(wp.MaxMoves-wp.MinMoves))

	case PatternPyramid:
		// Pyramid: peak in the middle
		midpoint := (totalRounds + 1) / 2
		if roundNumber <= midpoint {
			// Ascending
			if totalRounds == 1 {
				return wp.MinMoves
			}
			progress := float64(roundNumber-1) / float64(midpoint-1)
			return wp.MinMoves + int(progress*float64(wp.MaxMoves-wp.MinMoves))
		} else {
			// Descending
			descendingRound := totalRounds - roundNumber + 1
			if midpoint == 1 {
				return wp.MinMoves
			}
			progress := float64(descendingRound-1) / float64(midpoint-1)
			return wp.MinMoves + int(progress*float64(wp.MaxMoves-wp.MinMoves))
		}

	case PatternConstant:
		// Constant: average of min and max
		return (wp.MinMoves + wp.MaxMoves) / 2

	case PatternRandom:
		// Random: will be handled by LLM
		return (wp.MinMoves + wp.MaxMoves) / 2

	default:
		return wp.MinMoves
	}
}
