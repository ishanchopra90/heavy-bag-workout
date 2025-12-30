package gui

import (
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"time"
)

// startAnimationSequence starts the timer-based animation sequence for a combo
// This is called when a beep plays during work period
func (a *App) startAnimationSequence() {
	// Check if we're still in the workout period
	// If not, switch to rest period
	if a.currentPeriod == types.PeriodWork {
		elapsed := time.Since(a.workoutStartTime)
		if elapsed >= a.workoutPeriodDuration {
			// Workout period has ended, switch to rest
			a.currentPeriod = types.PeriodRest
			a.startRestPeriodAnimation()
			return
		}
	}

	// Only start animation sequence during work period
	if a.currentPeriod != types.PeriodWork {
		return
	}

	if a.currentCombo.IsEmpty() || len(a.currentCombo.Moves) == 0 {
		return
	}

	// Stop any existing animation timer
	a.stopAnimationSequence()

	// Get time per move based on tempo
	timePerMove := a.getTimePerMove()

	// Get tempo interval (time between beeps)
	tempoInterval := a.selectedTempo.Duration()
	if a.selectedTempo == models.TempoSuperfast {
		tempoInterval = 1 * time.Second
	}

	// Calculate total combo time
	totalComboTime := time.Duration(len(a.currentCombo.Moves)) * timePerMove

	// Calculate idle duration (remainder until next beep)
	idleDuration := tempoInterval - totalComboTime
	if idleDuration < 0 {
		idleDuration = 0 // Safety check
	}

	// Set up animation sequence goroutine
	a.animationTimerStop = make(chan struct{})
	a.currentMoveIndex = 0

	// Play beep at the start of the combo sequence
	if a.audioHandler != nil {
		a.audioHandler.PlayBeep()
	}

	// Start with first move
	a.startNextMove(timePerMove, tempoInterval, idleDuration)
}

// startNextMove starts the timer for the current move in the combo
func (a *App) startNextMove(timePerMove, tempoInterval, idleDuration time.Duration) {
	if a.currentPeriod != types.PeriodWork {
		return
	}

	if a.currentMoveIndex >= len(a.currentCombo.Moves) {
		// All moves completed, start idle animation
		a.startIdleAnimation(idleDuration, tempoInterval)
		return
	}

	// Get the current move
	move := a.currentCombo.Moves[a.currentMoveIndex]
	animationState := a.getAnimationStateForMove(move)

	// Set the animation
	if a.characterSprite != nil {
		a.characterSprite.SetAnimation(animationState)
	}

	// Show "go!" indicator
	a.showGo = true
	if a.window != nil {
		a.window.Invalidate()
	}

	// Hide "go!" after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		if a.currentPeriod == types.PeriodWork {
			a.showGo = false
			if a.window != nil {
				a.window.Invalidate()
			}
		}
	}()

	// Create timer for this move
	a.animationTimer = time.AfterFunc(timePerMove, func() {
		select {
		case <-a.animationTimerStop:
			return // Sequence was stopped
		default:
			// Move to next move
			a.currentMoveIndex++
			a.startNextMove(timePerMove, tempoInterval, idleDuration)
		}
	})
}

// startIdleAnimation starts the idle animation timer after combo completes
func (a *App) startIdleAnimation(idleDuration, tempoInterval time.Duration) {
	if a.currentPeriod != types.PeriodWork {
		return
	}

	// Set to idle animation
	if a.characterSprite != nil {
		a.characterSprite.SetAnimation(AnimationStateIdle)
	}

	if a.window != nil {
		a.window.Invalidate()
	}

	// If there's idle time, create timer for it
	if idleDuration > 0 {
		a.animationTimer = time.AfterFunc(idleDuration, func() {
			select {
			case <-a.animationTimerStop:
				return // Sequence was stopped
			default:
				// Idle completed, play beep and start next combo
				// Beep plays at the start of the next combo sequence
				a.startAnimationSequence()
			}
		})
	} else {
		// No idle time, immediately start next combo (beep will play at start)
		a.startAnimationSequence()
	}
}

// startRestPeriodAnimation starts the rest period animation
func (a *App) startRestPeriodAnimation() {
	// Stop any existing animation sequence
	a.stopAnimationSequence()

	// Set to idle animation (non-looping)
	if a.characterSprite != nil {
		a.characterSprite.SetAnimation(AnimationStateIdle)
	}

	// Get rest duration from current round
	if a.currentRound > 0 && a.currentRound <= len(a.workout.Rounds) {
		round := a.workout.Rounds[a.currentRound-1]
		restDuration := round.RestDuration

		// Create timer for rest period
		a.animationTimerStop = make(chan struct{})
		a.animationTimer = time.AfterFunc(restDuration, func() {
			select {
			case <-a.animationTimerStop:
				return // Rest period was stopped
			default:
				// Rest period completed - timer will call OnPeriodStart for next work period
				// Animation will be updated by OnPeriodStart
			}
		})
	}

	if a.window != nil {
		a.window.Invalidate()
	}
}

// stopAnimationSequence stops the current animation sequence
func (a *App) stopAnimationSequence() {
	if a.animationTimer != nil {
		a.animationTimer.Stop()
		a.animationTimer = nil
	}
	if a.animationTimerStop != nil {
		// Signal stop without closing (channel will be recreated on next start)
		select {
		case a.animationTimerStop <- struct{}{}:
		default:
			// Channel already has a value or is closed, ignore
		}
		a.animationTimerStop = nil
	}
}

// getTimePerMove returns the time per move based on tempo
func (a *App) getTimePerMove() time.Duration {
	switch a.selectedTempo {
	case models.TempoSlow:
		return 400 * time.Millisecond
	case models.TempoMedium:
		return 400 * time.Millisecond
	case models.TempoFast:
		return 400 * time.Millisecond
	case models.TempoSuperfast:
		return 400 * time.Millisecond
	default:
		return 400 * time.Millisecond
	}
}
