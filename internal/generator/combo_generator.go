package generator

import (
	"heavybagworkout/internal/models"
	"math/rand"
	"time"
)

// ComboGenerator generates random boxing combos
type ComboGenerator struct {
	rng              *rand.Rand
	includeDefensive bool
}

const (
	maxSequentialIdenticalPunches = 2
)

// newComboGeneratorWithSource creates a combo generator with a custom random source (used for testing).
func newComboGeneratorWithSource(includeDefensive bool, source rand.Source) *ComboGenerator {
	if source == nil {
		source = rand.NewSource(time.Now().UnixNano())
	}
	return &ComboGenerator{
		rng:              rand.New(source),
		includeDefensive: includeDefensive,
	}
}

// NewComboGenerator creates a new combo generator
func NewComboGenerator(includeDefensive bool) *ComboGenerator {
	return newComboGeneratorWithSource(includeDefensive, nil)
}

// GenerateCombo generates a random combo with the specified number of moves
// minMoves and maxMoves define the range of moves in the combo
func (cg *ComboGenerator) GenerateCombo(minMoves, maxMoves int) models.Combo {
	if minMoves < 1 {
		minMoves = 1
	}
	if maxMoves < minMoves {
		maxMoves = minMoves
	}

	numMoves := minMoves
	if maxMoves > minMoves {
		numMoves = minMoves + cg.rng.Intn(maxMoves-minMoves+1)
	}

	for attempt := 0; attempt < 10; attempt++ {
		moves := make([]models.Move, 0, numMoves)

		for i := 0; i < numMoves; i++ {
			// Decide whether to add a punch or defensive move
			shouldAddDefensive := cg.includeDefensive && cg.rng.Float32() < 0.3 // 30% chance for defensive move

			if shouldAddDefensive {
				defensiveMoves := models.AllDefensiveMoves()
				defensiveMove := defensiveMoves[cg.rng.Intn(len(defensiveMoves))]
				moves = append(moves, models.NewDefensiveMove(defensiveMove))
			} else {
				punches := models.AllPunches()
				punch := punches[cg.rng.Intn(len(punches))]
				moves = append(moves, models.NewPunchMove(punch))
			}
		}

		if cg.isValidCombo(moves) {
			return models.NewCombo(moves)
		}
	}

	// Fallback to random combo even if validation fails repeatedly
	fallbackMoves := make([]models.Move, 0, numMoves)
	for i := 0; i < numMoves; i++ {
		punches := models.AllPunches()
		punch := punches[cg.rng.Intn(len(punches))]
		fallbackMoves = append(fallbackMoves, models.NewPunchMove(punch))
	}
	return models.NewCombo(fallbackMoves)
}

// GenerateSimpleCombo generates a simple combo (typically 1-3 moves, mostly punches)
func (cg *ComboGenerator) GenerateSimpleCombo() models.Combo {
	return cg.GenerateCombo(1, 3)
}

// GenerateMediumCombo generates a medium combo (typically 2-4 moves)
func (cg *ComboGenerator) GenerateMediumCombo() models.Combo {
	return cg.GenerateCombo(2, 4)
}

// GenerateComplexCombo generates a complex combo (typically 3-6 moves)
func (cg *ComboGenerator) GenerateComplexCombo() models.Combo {
	return cg.GenerateCombo(3, 6)
}

// GenerateCombosForWorkPeriod creates exactly one combo per work period.
// The combo complexity is influenced by the workout pattern.
// previousMoveCount is an optional parameter (nil for first round) that indicates the number of moves
// in the previous round, used to enforce pattern constraints (e.g., non-decreasing for linear).
func (cg *ComboGenerator) GenerateCombosForWorkPeriod(roundNumber, totalRounds int, workDuration time.Duration, pattern models.WorkoutPattern, previousMoveCount *int) models.Combo {
	if roundNumber < 1 {
		roundNumber = 1
	}
	if totalRounds < 1 {
		totalRounds = 1
	}

	// Always generate exactly 1 combo per round
	targetMoves := pattern.GetMovesPerRound(roundNumber, totalRounds)
	if targetMoves < 1 {
		targetMoves = 1
	}

	minMoves := targetMoves - 1
	if minMoves < pattern.MinMoves {
		minMoves = pattern.MinMoves
	}
	maxMoves := targetMoves + 1
	if maxMoves > pattern.MaxMoves {
		maxMoves = pattern.MaxMoves
	}
	if maxMoves < minMoves {
		maxMoves = minMoves
	}

	// Apply pattern constraints based on previous round
	if previousMoveCount != nil {
		switch pattern.Type {
		case models.PatternLinear:
			// For linear pattern: current must be >= previous (non-decreasing)
			if minMoves < *previousMoveCount {
				minMoves = *previousMoveCount
			}
			// Ensure minMoves doesn't exceed maxMoves
			if minMoves > maxMoves {
				maxMoves = minMoves
			}

		case models.PatternPyramid:
			// For pyramid pattern: check if we're ascending or descending
			prevExpectedMoves := pattern.GetMovesPerRound(roundNumber-1, totalRounds)
			if targetMoves > prevExpectedMoves {
				// Ascending phase: current must be >= previous
				if minMoves < *previousMoveCount {
					minMoves = *previousMoveCount
				}
				if minMoves > maxMoves {
					maxMoves = minMoves
				}
			} else if targetMoves < prevExpectedMoves {
				// Descending phase: current must be <= previous
				if maxMoves > *previousMoveCount {
					maxMoves = *previousMoveCount
				}
				if maxMoves < minMoves {
					minMoves = maxMoves
				}
			}
			// If targetMoves == prevExpectedMoves, we're at the peak, allow variation
		}
	}

	// Generate exactly one combo
	return cg.GenerateCombo(minMoves, maxMoves)
}

func (cg *ComboGenerator) isValidCombo(moves []models.Move) bool {
	return cg.validatePunchSequence(moves) && cg.validateDefensiveSequence(moves)
}

// validatePunchSequence checks if a sequence of moves has too many identical punch moves in a row.
// It returns false if the number of sequential identical punches exceeds maxSequentialIdenticalPunches.
//
// Examples (assuming maxSequentialIdenticalPunches == 3):
//
//	Jab, Jab, Jab     --> true   (3 in a row is OK)
//	Jab, Jab, Jab, Jab --> false (4 in a row is too many)
//	Jab, Cross, Jab   --> true   (different punches)
//	Jab, Slip, Jab    --> true   (defensive move breaks streak)
//
// Defensive moves reset the repeat count and punch tracking.
func (cg *ComboGenerator) validatePunchSequence(moves []models.Move) bool {
	var lastPunch *models.Punch
	repeatCount := 0

	for _, move := range moves {
		if move.IsPunch() && move.Punch != nil {
			if lastPunch != nil && *move.Punch == *lastPunch {
				repeatCount++
				if repeatCount >= maxSequentialIdenticalPunches {
					return false
				}
			} else {
				val := *move.Punch
				lastPunch = &val
				repeatCount = 1
			}
		} else {
			// reset count when encountering defensive move
			lastPunch = nil
			repeatCount = 0
		}
	}
	return true
}

// validateDefensiveSequence checks for realistic placement of defensive moves in a combo.
//   - Two defensive moves should not appear consecutively (e.g., "Slip, Slip" is invalid)
//   - Every defensive move must be adjacent to at least one punch
//     (e.g., "Jab, Slip, Cross" is valid, "Slip, Duck" is invalid)
//
// Examples:
//
//	Jab, Slip, Cross        // valid (defensive between punches)
//	Slip, Jab, Cross        // valid (defensive at start next to punch)
//	Jab, Slip, Slip         // invalid (two defensive in a row)
//	Jab, Slip               // valid
//	Slip, Duck              // invalid (no punch next to defensive)
//	Slip, Jab, Slip         // valid
func (cg *ComboGenerator) validateDefensiveSequence(moves []models.Move) bool {
	previousWasDefensive := false

	for i, move := range moves {
		if move.IsDefensive() {
			// Two defensive moves in a row is invalid
			if previousWasDefensive {
				return false
			}

			// Defensive moves must be adjacent to at least one punch
			hasPunchAround := false
			if i > 0 && moves[i-1].IsPunch() {
				hasPunchAround = true
			}
			if i < len(moves)-1 && moves[i+1].IsPunch() {
				hasPunchAround = true
			}
			if !hasPunchAround {
				return false
			}

			previousWasDefensive = true
		} else {
			previousWasDefensive = false
		}
	}

	return true
}
