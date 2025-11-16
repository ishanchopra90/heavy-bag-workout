package cli

import (
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/timer"
	"heavybagworkout/internal/types"
	"sync"
	"testing"
	"time"
)

func TestNewWorkoutInterface(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 20*time.Second, 10*time.Second),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	if workoutInterface.workout.RoundCount() != 1 {
		t.Errorf("expected RoundCount 1, got %d", workoutInterface.workout.RoundCount())
	}
	if workoutInterface.display.stance != models.Orthodox {
		t.Errorf("expected stance Orthodox, got %v", workoutInterface.display.stance)
	}
	if workoutInterface.audioHandler == nil {
		t.Error("expected audioHandler to be set")
	}
}

func TestNewWorkoutInterfaceWithStance(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterfaceWithStance(workout, audioHandler, models.Southpaw)

	if workoutInterface.display.stance != models.Southpaw {
		t.Errorf("expected stance Southpaw, got %v", workoutInterface.display.stance)
	}
}

func TestWorkoutInterface_SetStance(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	workoutInterface.SetStance(models.Southpaw)

	if workoutInterface.display.stance != models.Southpaw {
		t.Errorf("expected stance Southpaw after SetStance, got %v", workoutInterface.display.stance)
	}
}

func TestWorkoutInterface_Pause(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(2*time.Second, 1*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 2*time.Second, 1*time.Second),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	workoutInterface.Pause()

	if !workoutInterface.display.isPaused {
		t.Error("expected display to be paused")
	}

	// Verify timer can be resumed (indicating it was paused)
	time.Sleep(100 * time.Millisecond)
	remainingBeforeResume := workoutInterface.workoutTimer.RemainingTime()

	if err := workoutInterface.Resume(); err != nil {
		t.Fatalf("failed to resume: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	remainingAfterResume := workoutInterface.workoutTimer.RemainingTime()

	// Time should have decreased after resume, indicating it was paused
	if remainingAfterResume >= remainingBeforeResume {
		t.Error("expected time to decrease after resume, indicating timer was paused")
	}
}

func TestWorkoutInterface_Resume(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(2*time.Second, 1*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 2*time.Second, 1*time.Second),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	workoutInterface.Pause()
	time.Sleep(50 * time.Millisecond)

	if err := workoutInterface.Resume(); err != nil {
		t.Fatalf("failed to resume workout: %v", err)
	}

	if workoutInterface.display.isPaused {
		t.Error("expected display to not be paused after resume")
	}

	// Verify timer is running by checking that time decreases
	remaining1 := workoutInterface.workoutTimer.RemainingTime()
	time.Sleep(100 * time.Millisecond)
	remaining2 := workoutInterface.workoutTimer.RemainingTime()

	if remaining2 >= remaining1 {
		t.Error("expected time to decrease, indicating timer is running")
	}
}

func TestWorkoutInterface_Stop(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(2*time.Second, 1*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 2*time.Second, 1*time.Second),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	workoutInterface.Stop()

	if workoutInterface.workoutTimer.CurrentRound() != 0 {
		t.Errorf("expected currentRound 0 after stop, got %d", workoutInterface.workoutTimer.CurrentRound())
	}

	// Verify quitChan receives signal
	select {
	case <-workoutInterface.quitChan:
		// Good, quit signal received
	case <-time.After(100 * time.Millisecond):
		t.Error("expected quit signal after stop")
	}
}

func TestWorkoutInterface_TogglePause(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(2*time.Second, 1*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 2*time.Second, 1*time.Second),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Toggle to pause
	if err := workoutInterface.TogglePause(); err != nil {
		t.Fatalf("failed to toggle pause: %v", err)
	}

	if !workoutInterface.display.isPaused {
		t.Error("expected display to be paused after toggle")
	}

	time.Sleep(50 * time.Millisecond)

	// Toggle to resume
	if err := workoutInterface.TogglePause(); err != nil {
		t.Fatalf("failed to toggle resume: %v", err)
	}

	if workoutInterface.display.isPaused {
		t.Error("expected display to not be paused after second toggle")
	}
}

func TestWorkoutInterface_WorkoutTimerIntegration(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(1*time.Second, 500*time.Millisecond, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
			models.NewWorkoutRound(2, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	// Set up display handler
	workoutInterface.workoutTimer.SetDisplayHandler(workoutInterface.display)

	var workoutCompleted bool
	var mu sync.Mutex

	workoutInterface.workoutTimer.OnWorkoutComplete(func() {
		mu.Lock()
		workoutCompleted = true
		mu.Unlock()
		workoutInterface.quitChan <- true
	})

	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	// Verify initial state
	if workoutInterface.workoutTimer.CurrentRound() != 1 {
		t.Errorf("expected currentRound 1, got %d", workoutInterface.workoutTimer.CurrentRound())
	}
	if workoutInterface.workoutTimer.CurrentPeriod() != types.PeriodWork {
		t.Errorf("expected PeriodWork, got %v", workoutInterface.workoutTimer.CurrentPeriod())
	}

	// Wait for workout to complete
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		if workoutCompleted {
			mu.Unlock()
			break
		}
		mu.Unlock()
	}

	mu.Lock()
	if !workoutCompleted {
		mu.Unlock()
		t.Error("expected workout to complete")
		return
	}
	mu.Unlock()
}

func TestWorkoutInterface_DisplayHandlerIntegration(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(1*time.Second, 500*time.Millisecond, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{
				models.NewPunchMove(models.Jab),
			}), 1*time.Second, 500*time.Millisecond),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	// Verify display handler is set
	workoutInterface.workoutTimer.SetDisplayHandler(workoutInterface.display)

	if workoutInterface.workoutTimer == nil {
		t.Fatal("workoutTimer is nil")
	}

	// Start workout to trigger display callbacks
	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Verify display state is updated
	if workoutInterface.display.currentRound == 0 {
		t.Error("expected currentRound to be set after workout start")
	}

	workoutInterface.Stop()
}

func TestWorkoutInterface_AudioHandlerIntegration(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(1*time.Second, 500*time.Millisecond, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	// Verify audio handler is set
	workoutInterface.workoutTimer.SetAudioHandler(audioHandler)

	if workoutInterface.audioHandler == nil {
		t.Error("expected audioHandler to be set")
	}

	// Audio handler should be a NoOpAudioCueHandler, so calls should not panic
	audioHandler.PlayBeep()
	audioHandler.PlayPeriodTransition(types.PeriodWork)
	audioHandler.PlayWorkoutStart()
	audioHandler.PlayWorkoutComplete()
}

func TestWorkoutInterface_CurrentRoundAndPeriod(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(1*time.Second, 500*time.Millisecond, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
			models.NewWorkoutRound(2, models.NewCombo([]models.Move{}), 1*time.Second, 500*time.Millisecond),
		},
	)

	audioHandler := timer.NewNoOpAudioCueHandler()
	workoutInterface := NewWorkoutInterface(workout, audioHandler)

	// Before start
	if workoutInterface.workoutTimer.CurrentRound() != 0 {
		t.Errorf("expected currentRound 0 before start, got %d", workoutInterface.workoutTimer.CurrentRound())
	}

	if err := workoutInterface.workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}

	// After start
	if workoutInterface.workoutTimer.CurrentRound() != 1 {
		t.Errorf("expected currentRound 1 after start, got %d", workoutInterface.workoutTimer.CurrentRound())
	}

	if workoutInterface.workoutTimer.CurrentPeriod() != types.PeriodWork {
		t.Errorf("expected PeriodWork after start, got %v", workoutInterface.workoutTimer.CurrentPeriod())
	}

	workoutInterface.Stop()
}
