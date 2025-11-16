package timer

import (
	"heavybagworkout/internal/mocks"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestWorkoutTimer_StartAndComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workout := models.NewWorkout(
		models.NewWorkoutConfig(1*time.Second, 500*time.Millisecond, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
			models.NewWorkoutRound(2, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
		},
	)

	display := mocks.NewMockTimerDisplayHandler(ctrl)
	timer := NewWorkoutTimer(workout)
	timer.SetDisplayHandler(display)

	// Set up expectations
	display.EXPECT().OnWorkoutStart(2).Times(1)
	display.EXPECT().OnPeriodStart(types.PeriodWork, 1, 1*time.Second).Times(1)
	display.EXPECT().OnTimerUpdate(gomock.Any(), types.PeriodWork, 1).AnyTimes()
	display.EXPECT().OnPeriodEnd(types.PeriodWork, 1).Times(1)
	display.EXPECT().OnPeriodStart(types.PeriodRest, 1, 500*time.Millisecond).Times(1)
	display.EXPECT().OnTimerUpdate(gomock.Any(), types.PeriodRest, 1).AnyTimes()
	display.EXPECT().OnPeriodEnd(types.PeriodRest, 1).Times(1)
	display.EXPECT().OnPeriodStart(types.PeriodWork, 2, 1*time.Second).Times(1)
	display.EXPECT().OnTimerUpdate(gomock.Any(), types.PeriodWork, 2).AnyTimes()
	display.EXPECT().OnPeriodEnd(types.PeriodWork, 2).Times(1)
	display.EXPECT().OnPeriodStart(types.PeriodRest, 2, 500*time.Millisecond).Times(1)
	display.EXPECT().OnTimerUpdate(gomock.Any(), types.PeriodRest, 2).AnyTimes()
	display.EXPECT().OnPeriodEnd(types.PeriodRest, 2).Times(1)
	display.EXPECT().OnWorkoutComplete().Times(1)

	completed := false
	var completedMu sync.Mutex
	timer.OnWorkoutComplete(func() {
		completedMu.Lock()
		completed = true
		completedMu.Unlock()
	})

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting workout timer: %v", err)
	}

	// Wait for workout to complete (2 rounds * (1s work + 0.5s rest) = 3s total)
	// Add extra buffer for async operations
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		completedMu.Lock()
		if completed {
			completedMu.Unlock()
			break
		}
		completedMu.Unlock()
	}

	completedMu.Lock()
	if !completed {
		completedMu.Unlock()
		t.Fatalf("workout should have completed")
	}
	completedMu.Unlock()
}

func TestWorkoutTimer_PauseAndResume(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(2*time.Second, 1*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 2*time.Second, 1*time.Second),
		},
	)

	timer := NewWorkoutTimer(workout)

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting workout timer: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	timer.Pause()

	if timer.RemainingTime() == 0 {
		t.Fatalf("timer should have remaining time when paused")
	}

	time.Sleep(500 * time.Millisecond)

	// Time should not have changed while paused
	remainingBeforeResume := timer.RemainingTime()

	if err := timer.Resume(); err != nil {
		t.Fatalf("unexpected error resuming workout timer: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	remainingAfterResume := timer.RemainingTime()

	if remainingAfterResume >= remainingBeforeResume {
		t.Fatalf("timer should continue counting down after resume")
	}
}

func TestWorkoutTimer_CurrentRoundAndPeriod(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(1*time.Second, 500*time.Millisecond, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
			models.NewWorkoutRound(2, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
		},
	)

	timer := NewWorkoutTimer(workout)

	if timer.CurrentRound() != 0 {
		t.Fatalf("expected round 0 before start, got %d", timer.CurrentRound())
	}

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting workout timer: %v", err)
	}

	if timer.CurrentRound() != 1 {
		t.Fatalf("expected round 1, got %d", timer.CurrentRound())
	}

	if timer.CurrentPeriod() != types.PeriodWork {
		t.Fatalf("expected work period, got %v", timer.CurrentPeriod())
	}

	// Wait for work period to complete (1s) plus buffer
	time.Sleep(1200 * time.Millisecond)

	if timer.CurrentPeriod() != types.PeriodRest {
		t.Fatalf("expected rest period, got %v", timer.CurrentPeriod())
	}

	// Wait for rest period to complete (500ms) and round transition with buffer for async callbacks
	// Round increments to 2 after rest completes, then work period for round 2 starts
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		round := timer.CurrentRound()
		period := timer.CurrentPeriod()
		if round == 2 && period == types.PeriodWork {
			break
		}
	}

	// After rest period completes, round increments and work period starts for round 2
	if timer.CurrentRound() != 2 {
		t.Fatalf("expected round 2, got %d", timer.CurrentRound())
	}

	if timer.CurrentPeriod() != types.PeriodWork {
		t.Fatalf("expected work period for round 2, got %v", timer.CurrentPeriod())
	}
}

func TestWorkoutTimer_Stop(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(5*time.Second, 2*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 5*time.Second, 2*time.Second),
		},
	)

	timer := NewWorkoutTimer(workout)

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting workout timer: %v", err)
	}

	time.Sleep(1 * time.Second)

	timer.Stop()

	if timer.CurrentRound() != 0 {
		t.Fatalf("expected round 0 after stop, got %d", timer.CurrentRound())
	}
}
