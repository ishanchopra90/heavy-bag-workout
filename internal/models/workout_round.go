package models

import "time"

// WorkoutRound represents a single round of a workout
// A round consists of a work period (with one combo) and a rest period
type WorkoutRound struct {
	RoundNumber  int           // The round number (1-indexed)
	Combo        Combo         // The combo to perform during the work period
	WorkDuration time.Duration // Duration of the work period
	RestDuration time.Duration // Duration of the rest period
}

// NewWorkoutRound creates a new workout round with a single combo
func NewWorkoutRound(roundNumber int, combo Combo, workDuration, restDuration time.Duration) WorkoutRound {
	return WorkoutRound{
		RoundNumber:  roundNumber,
		Combo:        combo,
		WorkDuration: workDuration,
		RestDuration: restDuration,
	}
}

// TotalDuration returns the total duration of the round (work + rest)
func (wr WorkoutRound) TotalDuration() time.Duration {
	return wr.WorkDuration + wr.RestDuration
}
