package generator

import (
	"heavybagworkout/internal/models"
	"math/rand"
	"testing"
	"time"
)

func TestComboGeneratorGenerateComboLength(t *testing.T) {
	gen := newComboGeneratorWithSource(false, rand.NewSource(42))
	combo := gen.GenerateCombo(2, 4)

	if combo.Length() < 2 || combo.Length() > 4 {
		t.Fatalf("combo length out of range: %d", combo.Length())
	}
}

func TestComboGeneratorGenerateComboIncludesDefensive(t *testing.T) {
	gen := newComboGeneratorWithSource(true, rand.NewSource(99))
	combo := gen.GenerateCombo(5, 5)

	hasDefensive := false
	for _, move := range combo.Moves {
		if move.IsDefensive() {
			hasDefensive = true
			break
		}
	}

	if !hasDefensive {
		t.Fatalf("expected combo to include defensive move: %+v", combo)
	}
}

func TestComboGeneratorHelpers(t *testing.T) {
	gen := newComboGeneratorWithSource(true, rand.NewSource(1))

	if combo := gen.GenerateSimpleCombo(); combo.Length() < 1 || combo.Length() > 3 {
		t.Fatalf("simple combo length out of range: %d", combo.Length())
	}

	if combo := gen.GenerateMediumCombo(); combo.Length() < 2 || combo.Length() > 4 {
		t.Fatalf("medium combo length out of range: %d", combo.Length())
	}

	if combo := gen.GenerateComplexCombo(); combo.Length() < 3 || combo.Length() > 6 {
		t.Fatalf("complex combo length out of range: %d", combo.Length())
	}
}

func TestComboStringRepresentation(t *testing.T) {
	combo := models.NewCombo([]models.Move{
		models.NewPunchMove(models.Jab),
		models.NewPunchMove(models.Cross),
		models.NewDefensiveMove(models.LeftSlip),
	})

	expected := "1, 2, Left Slip"
	if combo.String() != expected {
		t.Fatalf("expected %s, got %s", expected, combo.String())
	}
}

func TestComboGeneratorValidatePunchSequence(t *testing.T) {
	gen := NewComboGenerator(false)
	moves := []models.Move{
		models.NewPunchMove(models.Jab),
		models.NewPunchMove(models.Jab),
		models.NewPunchMove(models.Jab),
	}

	if gen.validatePunchSequence(moves) {
		t.Fatalf("expected punch sequence validation to fail for repeated punches")
	}
}

func TestComboGeneratorValidateDefensiveSequence(t *testing.T) {
	gen := NewComboGenerator(true)
	moves := []models.Move{
		models.NewDefensiveMove(models.LeftSlip),
		models.NewDefensiveMove(models.RightSlip),
		models.NewPunchMove(models.Cross),
	}

	if gen.validateDefensiveSequence(moves) {
		t.Fatalf("expected defensive sequence validation to fail for consecutive defensive moves")
	}
}

func TestGenerateCombosForWorkPeriod(t *testing.T) {
	gen := newComboGeneratorWithSource(true, rand.NewSource(5))
	pattern := models.NewWorkoutPattern(models.PatternLinear, 2, 5, true)

	workDuration := 20 * time.Second
	combo := gen.GenerateCombosForWorkPeriod(1, 5, workDuration, pattern, nil)

	// validate num ov moves between 2 and 5
	if combo.Length() < 2 || combo.Length() > 5 {
		t.Fatalf("combo length out of expected bounds: %d", combo.Length())
	}
}

func TestGenerateCombosForWorkPeriod_AlwaysOneCombo(t *testing.T) {
	gen := NewComboGenerator(true)
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 6, true)

	// Test with various work durations - should always return exactly 1 combo
	testCases := []struct {
		name         string
		workDuration time.Duration
		roundNumber  int
		totalRounds  int
	}{
		{"short duration", 6 * time.Second, 1, 3},
		{"medium duration", 20 * time.Second, 2, 5},
		{"long duration", 60 * time.Second, 3, 8},
		{"very long duration", 120 * time.Second, 1, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			combo := gen.GenerateCombosForWorkPeriod(tc.roundNumber, tc.totalRounds, tc.workDuration, pattern, nil)
			// validate combos not nil
			if combo.IsEmpty() {
				t.Fatalf("expected non-empty combo, got empty")
			}
		})
	}
}

func TestGenerateCombosForWorkPeriod_LinearPatternWithPreviousMoveCount(t *testing.T) {
	gen := newComboGeneratorWithSource(true, rand.NewSource(42))
	pattern := models.NewWorkoutPattern(models.PatternLinear, 3, 5, true)
	workDuration := 20 * time.Second

	// Test that linear pattern enforces non-decreasing move count
	testCases := []struct {
		name              string
		roundNumber       int
		totalRounds       int
		previousMoveCount int
		expectedMinMoves  int
	}{
		{"round 2 with previous 3 moves", 2, 5, 3, 3}, // Should be >= 3
		{"round 3 with previous 4 moves", 3, 5, 4, 4}, // Should be >= 4
		{"round 4 with previous 4 moves", 4, 5, 4, 4}, // Can stay same, should be >= 4
		{"round 5 with previous 5 moves", 5, 5, 5, 5}, // Should be >= 5
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prevCount := tc.previousMoveCount
			combo := gen.GenerateCombosForWorkPeriod(tc.roundNumber, tc.totalRounds, workDuration, pattern, &prevCount)

			if combo.IsEmpty() {
				t.Fatalf("expected non-empty combo, got empty")
			}

			// For linear pattern, current round must have >= previous round's moves
			if combo.Length() < tc.expectedMinMoves {
				t.Errorf("round %d: expected at least %d moves (previous had %d), got %d",
					tc.roundNumber, tc.expectedMinMoves, tc.previousMoveCount, combo.Length())
			}

			// Should still respect max moves
			if combo.Length() > pattern.MaxMoves {
				t.Errorf("round %d: expected at most %d moves, got %d",
					tc.roundNumber, pattern.MaxMoves, combo.Length())
			}
		})
	}
}

func TestGenerateCombosForWorkPeriod_PyramidPatternWithPreviousMoveCount(t *testing.T) {
	gen := newComboGeneratorWithSource(true, rand.NewSource(99))
	pattern := models.NewWorkoutPattern(models.PatternPyramid, 2, 6, true)
	workDuration := 20 * time.Second

	// Test pyramid pattern: ascending phase should be >= previous, descending should be <= previous
	testCases := []struct {
		name              string
		roundNumber       int
		totalRounds       int
		previousMoveCount int
		expectedMinMoves  int
		expectedMaxMoves  int
		phase             string // "ascending" or "descending"
	}{
		// Ascending phase (rounds 1-3 of 5): should be >= previous
		{"round 2 ascending", 2, 5, 2, 2, 6, "ascending"},
		{"round 3 ascending", 3, 5, 4, 4, 6, "ascending"},
		// Descending phase (rounds 4-5 of 5): should be <= previous
		{"round 4 descending", 4, 5, 6, 2, 6, "descending"},
		{"round 5 descending", 5, 5, 4, 2, 4, "descending"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prevCount := tc.previousMoveCount
			combo := gen.GenerateCombosForWorkPeriod(tc.roundNumber, tc.totalRounds, workDuration, pattern, &prevCount)

			if combo.IsEmpty() {
				t.Fatalf("expected non-empty combo, got empty")
			}

			if tc.phase == "ascending" {
				// Ascending: should be >= previous
				if combo.Length() < tc.expectedMinMoves {
					t.Errorf("round %d (ascending): expected at least %d moves (previous had %d), got %d",
						tc.roundNumber, tc.expectedMinMoves, tc.previousMoveCount, combo.Length())
				}
			} else if tc.phase == "descending" {
				// Descending: should be <= previous
				if combo.Length() > tc.expectedMaxMoves {
					t.Errorf("round %d (descending): expected at most %d moves (previous had %d), got %d",
						tc.roundNumber, tc.expectedMaxMoves, tc.previousMoveCount, combo.Length())
				}
			}

			// Should still respect overall min/max bounds
			if combo.Length() < pattern.MinMoves || combo.Length() > pattern.MaxMoves {
				t.Errorf("round %d: combo length %d out of bounds [%d, %d]",
					tc.roundNumber, combo.Length(), pattern.MinMoves, pattern.MaxMoves)
			}
		})
	}
}
