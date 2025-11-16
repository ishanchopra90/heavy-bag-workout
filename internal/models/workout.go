package models

import "time"

// Workout represents a complete workout session
type Workout struct {
	Config WorkoutConfig  // Configuration used to generate this workout
	Rounds []WorkoutRound // All rounds in the workout
}

// NewWorkout creates a new workout with the given configuration and rounds.
// The TotalRounds in the config will be automatically set to match the number of rounds provided,
// ensuring they are tightly coupled.
func NewWorkout(config WorkoutConfig, rounds []WorkoutRound) Workout {
	// Sync the TotalRounds in config to match the actual number of rounds
	config.TotalRounds = len(rounds)
	return Workout{
		Config: config,
		Rounds: rounds,
	}
}

// TotalDuration calculates the total duration of the workout
func (w Workout) TotalDuration() time.Duration {
	var total time.Duration
	for _, round := range w.Rounds {
		total += round.TotalDuration()
	}
	return total
}

// RoundCount returns the number of rounds in the workout
func (w Workout) RoundCount() int {
	return len(w.Rounds)
}

// IsEmpty returns true if the workout has no rounds
func (w Workout) IsEmpty() bool {
	return len(w.Rounds) == 0
}
