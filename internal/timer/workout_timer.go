package timer

import (
	"fmt"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"time"
)

// TimerDisplayHandler handles display updates for the timer
type TimerDisplayHandler interface {
	OnTimerUpdate(remaining time.Duration, periodType types.PeriodType, roundNumber int)
	OnPeriodStart(periodType types.PeriodType, roundNumber int, duration time.Duration)
	OnPeriodEnd(periodType types.PeriodType, roundNumber int)
	OnWorkoutStart(totalRounds int)
	OnWorkoutComplete()
}

// AudioCueHandler handles audio cues for period transitions
type AudioCueHandler interface {
	PlayBeep()
	PlayPeriodTransition(periodType types.PeriodType)
	PlayWorkoutStart()
	PlayWorkoutComplete()
	PlayComboCallout(combo models.Combo, stance models.Stance)
}

// WorkoutTimer manages the execution of a workout with work and rest periods
type WorkoutTimer struct {
	workout           models.Workout
	currentRound      int
	currentPeriod     types.PeriodType
	workTimer         *WorkPeriodTimer
	restTimer         *RestPeriodTimer
	displayHandler    TimerDisplayHandler
	audioHandler      AudioCueHandler
	stance            models.Stance // Stance for combo callouts
	onRoundComplete   func(roundNumber int)
	onWorkoutComplete func()
}

// NewWorkoutTimer creates a new workout timer
func NewWorkoutTimer(workout models.Workout) *WorkoutTimer {
	return &WorkoutTimer{
		workout:       workout,
		currentRound:  0,
		currentPeriod: types.PeriodWork,
		stance:        models.Orthodox, // Default stance
	}
}

// SetStance sets the boxer's stance for combo callouts
func (wt *WorkoutTimer) SetStance(stance models.Stance) {
	wt.stance = stance
}

// SetDisplayHandler sets the display handler for timer updates
func (wt *WorkoutTimer) SetDisplayHandler(handler TimerDisplayHandler) {
	wt.displayHandler = handler
}

// SetAudioHandler sets the audio handler for period transitions
func (wt *WorkoutTimer) SetAudioHandler(handler AudioCueHandler) {
	wt.audioHandler = handler
}

// OnRoundComplete sets a callback for when a round completes
func (wt *WorkoutTimer) OnRoundComplete(callback func(roundNumber int)) {
	wt.onRoundComplete = callback
}

// OnWorkoutComplete sets a callback for when the workout completes
func (wt *WorkoutTimer) OnWorkoutComplete(callback func()) {
	wt.onWorkoutComplete = callback
}

// Start begins the workout timer
func (wt *WorkoutTimer) Start() error {
	if len(wt.workout.Rounds) == 0 {
		return fmt.Errorf("workout has no rounds")
	}

	wt.currentRound = 1
	wt.currentPeriod = types.PeriodWork

	if wt.displayHandler != nil {
		wt.displayHandler.OnWorkoutStart(len(wt.workout.Rounds))
	}

	if wt.audioHandler != nil {
		wt.audioHandler.PlayWorkoutStart()
	}

	return wt.startWorkPeriod()
}

// Pause pauses the current timer
func (wt *WorkoutTimer) Pause() {
	if wt.workTimer != nil && wt.workTimer.State() == StateRunning {
		wt.workTimer.Pause()
	}
	if wt.restTimer != nil && wt.restTimer.State() == StateRunning {
		wt.restTimer.Pause()
	}
}

// Resume resumes the paused timer
func (wt *WorkoutTimer) Resume() error {
	if wt.workTimer != nil && wt.workTimer.State() == StatePaused {
		return wt.workTimer.Start()
	}
	if wt.restTimer != nil && wt.restTimer.State() == StatePaused {
		return wt.restTimer.Start()
	}
	return nil
}

// Stop stops the workout timer
func (wt *WorkoutTimer) Stop() {
	if wt.workTimer != nil {
		wt.workTimer.Stop()
	}
	if wt.restTimer != nil {
		wt.restTimer.Stop()
	}
	wt.currentRound = 0
}

// CurrentRound returns the current round number (1-indexed, 0 if not started)
func (wt *WorkoutTimer) CurrentRound() int {
	return wt.currentRound
}

// CurrentPeriod returns the current period type
func (wt *WorkoutTimer) CurrentPeriod() types.PeriodType {
	return wt.currentPeriod
}

// RemainingTime returns the remaining time for the current period
func (wt *WorkoutTimer) RemainingTime() time.Duration {
	if wt.currentPeriod == types.PeriodWork && wt.workTimer != nil {
		return wt.workTimer.Remaining()
	}
	if wt.currentPeriod == types.PeriodRest && wt.restTimer != nil {
		return wt.restTimer.Remaining()
	}
	return 0
}

// startWorkPeriod starts a work period for the current round
func (wt *WorkoutTimer) startWorkPeriod() error {
	if wt.currentRound > len(wt.workout.Rounds) {
		wt.completeWorkout()
		return nil
	}

	round := wt.workout.Rounds[wt.currentRound-1]
	wt.workTimer = NewWorkPeriodTimer(round.WorkDuration)

	// Set up callbacks
	wt.workTimer.OnTick(func(remaining time.Duration) {
		if wt.displayHandler != nil {
			wt.displayHandler.OnTimerUpdate(remaining, types.PeriodWork, wt.currentRound)
		}
	}).OnComplete(func() {
		wt.onWorkPeriodComplete()
	})

	if wt.displayHandler != nil {
		wt.displayHandler.OnPeriodStart(types.PeriodWork, wt.currentRound, round.WorkDuration)
	}

	if wt.audioHandler != nil {
		wt.audioHandler.PlayPeriodTransition(types.PeriodWork)
		// Call out the combo for this round
		wt.audioHandler.PlayComboCallout(round.Combo, wt.stance)
	}

	return wt.workTimer.Start()
}

// onWorkPeriodComplete handles the completion of a work period
func (wt *WorkoutTimer) onWorkPeriodComplete() {
	if wt.displayHandler != nil {
		wt.displayHandler.OnPeriodEnd(types.PeriodWork, wt.currentRound)
	}

	// Start rest period
	wt.currentPeriod = types.PeriodRest
	wt.startRestPeriod()
}

// startRestPeriod starts a rest period for the current round
func (wt *WorkoutTimer) startRestPeriod() error {
	round := wt.workout.Rounds[wt.currentRound-1]
	wt.restTimer = NewRestPeriodTimer(round.RestDuration)

	// Track last beep time to avoid multiple beeps in the same second
	lastBeepSecond := -1

	// Set up callbacks
	wt.restTimer.OnTick(func(remaining time.Duration) {
		if wt.displayHandler != nil {
			wt.displayHandler.OnTimerUpdate(remaining, types.PeriodRest, wt.currentRound)
		}

		// Play beep in the last 3 seconds of rest period (3, 2, 1 seconds remaining)
		if wt.audioHandler != nil {
			remainingSeconds := int(remaining.Seconds())
			// Only beep if we're in the last 3 seconds and haven't beeped for this second yet
			if remainingSeconds <= 3 && remainingSeconds > 0 && remainingSeconds != lastBeepSecond {
				wt.audioHandler.PlayBeep()
				lastBeepSecond = remainingSeconds
			}
		}
	}).OnComplete(func() {
		wt.onRestPeriodComplete()
	})

	if wt.displayHandler != nil {
		wt.displayHandler.OnPeriodStart(types.PeriodRest, wt.currentRound, round.RestDuration)
	}

	if wt.audioHandler != nil {
		wt.audioHandler.PlayPeriodTransition(types.PeriodRest)
	}

	return wt.restTimer.Start()
}

// onRestPeriodComplete handles the completion of a rest period
func (wt *WorkoutTimer) onRestPeriodComplete() {
	if wt.displayHandler != nil {
		wt.displayHandler.OnPeriodEnd(types.PeriodRest, wt.currentRound)
	}

	// Notify round completion
	if wt.onRoundComplete != nil {
		wt.onRoundComplete(wt.currentRound)
	}

	// Move to next round
	wt.currentRound++
	wt.currentPeriod = types.PeriodWork

	// Start next work period or complete workout
	if wt.currentRound <= len(wt.workout.Rounds) {
		wt.startWorkPeriod()
	} else {
		wt.completeWorkout()
	}
}

// completeWorkout handles workout completion
func (wt *WorkoutTimer) completeWorkout() {
	if wt.displayHandler != nil {
		wt.displayHandler.OnWorkoutComplete()
	}

	if wt.audioHandler != nil {
		wt.audioHandler.PlayWorkoutComplete()
	}

	if wt.onWorkoutComplete != nil {
		wt.onWorkoutComplete()
	}
}
