package generator

import (
	"fmt"
	"heavybagworkout/internal/models"
	"time"
)

type combosForWorkPeriodGenerator interface {
	GenerateCombosForWorkPeriod(roundNumber, totalRounds int, workDuration time.Duration, pattern models.WorkoutPattern, previousMoveCount *int) models.Combo
}

// WorkoutGenerator produces full workouts using combo generation, configuration, and patterns.
type WorkoutGenerator struct {
	comboGenFactory func(includeDefensive bool) combosForWorkPeriodGenerator
}

// NewWorkoutGenerator creates a WorkoutGenerator with the default combo generator factory.
func NewWorkoutGenerator() *WorkoutGenerator {
	return &WorkoutGenerator{
		comboGenFactory: func(includeDefensive bool) combosForWorkPeriodGenerator {
			return NewComboGenerator(includeDefensive)
		},
	}
}

// NewWorkoutGeneratorWithFactory allows injection of a custom combo generator factory (useful for testing).
func NewWorkoutGeneratorWithFactory(factory func(includeDefensive bool) combosForWorkPeriodGenerator) *WorkoutGenerator {
	if factory == nil {
		factory = func(includeDefensive bool) combosForWorkPeriodGenerator {
			return NewComboGenerator(includeDefensive)
		}
	}
	return &WorkoutGenerator{
		comboGenFactory: factory,
	}
}

// GenerateWorkout produces a full workout based on the provided configuration and pattern.
func (wg *WorkoutGenerator) GenerateWorkout(config models.WorkoutConfig, pattern models.WorkoutPattern) (models.Workout, error) {
	if err := config.Validate(); err != nil {
		return models.Workout{}, fmt.Errorf("invalid workout config: %w", err)
	}
	if pattern.MinMoves <= 0 {
		pattern.MinMoves = 1
	}
	// Enforce maximum limit of 5 moves per combo
	if pattern.MaxMoves > 5 {
		pattern.MaxMoves = 5
	}
	if pattern.MaxMoves < pattern.MinMoves {
		pattern.MaxMoves = pattern.MinMoves
	}

	comboGen := wg.comboGenerator(pattern.IncludeDefensive)
	rounds := make([]models.WorkoutRound, 0, config.TotalRounds)

	for roundNumber := 1; roundNumber <= config.TotalRounds; roundNumber++ {
		// Get previous round's move count if available
		var previousMoveCount *int
		if roundNumber > 1 && len(rounds) > 0 {
			prevMoves := len(rounds[roundNumber-2].Combo.Moves)
			previousMoveCount = &prevMoves
		}

		combo := wg.distributeCombosAcrossWorkPeriod(comboGen, roundNumber, config, pattern, previousMoveCount)
		round := models.NewWorkoutRound(roundNumber, combo, config.WorkDuration, config.RestDuration)
		rounds = append(rounds, round)
	}

	return models.NewWorkout(config, rounds), nil
}

// CalculateTotalWorkoutDuration returns the total workout duration (sum of work + rest for all rounds).
func (wg *WorkoutGenerator) CalculateTotalWorkoutDuration(config models.WorkoutConfig) time.Duration {
	return (config.WorkDuration + config.RestDuration) * time.Duration(config.TotalRounds)
}

func (wg *WorkoutGenerator) comboGenerator(includeDefensive bool) combosForWorkPeriodGenerator {
	if wg == nil || wg.comboGenFactory == nil {
		return NewComboGenerator(includeDefensive)
	}
	return wg.comboGenFactory(includeDefensive)
}

func (wg *WorkoutGenerator) distributeCombosAcrossWorkPeriod(comboGen combosForWorkPeriodGenerator, roundNumber int, config models.WorkoutConfig, pattern models.WorkoutPattern, previousMoveCount *int) models.Combo {
	return comboGen.GenerateCombosForWorkPeriod(roundNumber, config.TotalRounds, config.WorkDuration, pattern, previousMoveCount)
}
