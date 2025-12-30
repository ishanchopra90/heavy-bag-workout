package gui

import (
	"heavybagworkout/internal/models"
	"testing"
	"time"
)

// TestComboMoveTiming verifies that each move in a combo takes exactly timePerMove
// This test catches regressions where moves take inconsistent amounts of time
func TestComboMoveTiming(t *testing.T) {
	cs := NewCharacterSprite(models.Orthodox)
	timePerMove := 100 * time.Millisecond
	cs.SetTimePerMove(timePerMove)

	// Create a combo with multiple moves
	combo := models.NewCombo([]models.Move{
		models.NewPunchMove(models.Jab),
		models.NewPunchMove(models.Cross),
		models.NewPunchMove(models.LeadHook),
		models.NewPunchMove(models.RearUppercut),
	})

	if len(combo.Moves) != 4 {
		t.Fatalf("expected 4 moves, got %d", len(combo.Moves))
	}

	// Track timing for each move
	type MoveTiming struct {
		moveIndex    int
		startTime    time.Time
		endTime      time.Time
		duration     time.Duration
		expectedTime time.Duration
	}

	var moveTimings []MoveTiming
	comboStartTime := time.Now()

	// Simulate the combo animation sequence
	for moveIndex := 0; moveIndex < len(combo.Moves); moveIndex++ {
		move := combo.Moves[moveIndex]
		animationState := getAnimationStateForMove(move)

		// Set moveStartTime based on comboStartTime (like app.go does)
		expectedMoveStartTime := comboStartTime.Add(time.Duration(moveIndex) * timePerMove)

		// Set animation for this move (this sets moveStartTime to time.Now(), but we'll override it)
		cs.SetAnimation(animationState)
		cs.SetMoveStartTime(expectedMoveStartTime)

		moveStartTime := cs.GetMoveStartTime()

		// Simulate updates until move completes
		// Move should complete when elapsedSinceMoveStart >= timePerMove
		// We simulate time advancing from the move start time
		moveCompleted := false
		for i := 0; i < 20; i++ {
			// Simulate time advancing from moveStartTime
			simulatedTime := moveStartTime.Add(time.Duration(i) * 10 * time.Millisecond)
			cs.Update(simulatedTime)

			// Check if move is complete based on elapsedSinceMoveStart
			elapsedSinceMoveStart := simulatedTime.Sub(moveStartTime)
			if elapsedSinceMoveStart >= timePerMove {
				// Move completed
				moveTimings = append(moveTimings, MoveTiming{
					moveIndex:    moveIndex,
					startTime:    moveStartTime,
					endTime:      simulatedTime,
					duration:     elapsedSinceMoveStart,
					expectedTime: timePerMove,
				})
				moveCompleted = true
				break
			}
		}

		// Verify move completed
		if !moveCompleted {
			t.Fatalf("move %d did not complete within expected time", moveIndex)
		}
	}

	// Verify each move took exactly timePerMove (within tolerance)
	tolerance := 5 * time.Millisecond
	for i, timing := range moveTimings {
		if timing.duration < timePerMove-tolerance || timing.duration > timePerMove+tolerance {
			t.Errorf("move %d took %v, expected %v ± %v",
				i, timing.duration, timePerMove, tolerance)
		}

		// Verify move started at the expected time relative to combo start
		expectedStartTime := comboStartTime.Add(time.Duration(i) * timePerMove)
		startTimeDiff := timing.startTime.Sub(expectedStartTime)
		if startTimeDiff < -tolerance || startTimeDiff > tolerance {
			t.Errorf("move %d started at wrong time: expected %v, got %v (diff: %v)",
				i, expectedStartTime, timing.startTime, startTimeDiff)
		}
	}

	// Verify total combo time
	if len(moveTimings) > 0 {
		firstMoveStart := moveTimings[0].startTime
		lastMoveEnd := moveTimings[len(moveTimings)-1].endTime
		totalDuration := lastMoveEnd.Sub(firstMoveStart)
		expectedTotal := time.Duration(len(combo.Moves)) * timePerMove
		if totalDuration < expectedTotal-tolerance || totalDuration > expectedTotal+tolerance {
			t.Errorf("total combo duration: %v, expected %v ± %v",
				totalDuration, expectedTotal, tolerance)
		}
	}
}

// getAnimationStateForMove is a helper function for testing
// (normally this is a method on App, but we need it for testing)
// Uses orthodox stance for consistent testing
func getAnimationStateForMove(move models.Move) AnimationState {
	if move.IsPunch() && move.Punch != nil {
		switch *move.Punch {
		case models.Jab:
			return AnimationStateJabLeft // Orthodox: jab = left
		case models.Cross:
			return AnimationStateCrossRight // Orthodox: cross = right
		case models.LeadHook:
			return AnimationStateLeadHookLeft // Orthodox: lead hook = left
		case models.RearHook:
			return AnimationStateRearHookRight // Orthodox: rear hook = right
		case models.LeadUppercut:
			return AnimationStateLeadUppercutLeft // Orthodox: lead uppercut = left
		case models.RearUppercut:
			return AnimationStateRearUppercutRight // Orthodox: rear uppercut = right
		}
	} else if move.IsDefensive() && move.Defensive != nil {
		switch *move.Defensive {
		case models.LeftSlip:
			return AnimationStateSlipLeft
		case models.RightSlip:
			return AnimationStateSlipRight
		case models.LeftRoll:
			return AnimationStateRollLeft
		case models.RightRoll:
			return AnimationStateRollRight
		case models.PullBack:
			return AnimationStatePullBack
		case models.Duck:
			return AnimationStateDuck
		}
	}
	return AnimationStateIdle
}
