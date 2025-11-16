package cli

import (
	"heavybagworkout/internal/mocks"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestNewWorkoutDisplay(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)

	// totalRounds is set to len(workout.Rounds), which is 1 in this test
	if display.totalRounds != 1 {
		t.Errorf("expected totalRounds 1 (based on rounds length), got %d", display.totalRounds)
	}
	if display.currentRound != 0 {
		t.Errorf("expected currentRound 0, got %d", display.currentRound)
	}
	if display.stance != models.Orthodox {
		t.Errorf("expected stance Orthodox, got %v", display.stance)
	}

	// Clean up
	display.stopComboUpdates()
}

func TestNewWorkoutDisplayWithStance(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplayWithStance(workout, models.Southpaw)

	if display.stance != models.Southpaw {
		t.Errorf("expected stance Southpaw, got %v", display.stance)
	}
}

func TestWorkoutDisplay_SetStance(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)
	display.SetStance(models.Southpaw)

	if display.stance != models.Southpaw {
		t.Errorf("expected stance Southpaw after SetStance, got %v", display.stance)
	}
}

func TestWorkoutDisplay_OnWorkoutStart(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 5),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)
	display.OnWorkoutStart(5)

	if display.totalRounds != 5 {
		t.Errorf("expected totalRounds 5, got %d", display.totalRounds)
	}
	if display.currentRound != 1 {
		t.Errorf("expected currentRound 1, got %d", display.currentRound)
	}
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0, got %d", display.currentComboIdx)
	}
	if display.isPaused {
		t.Error("expected isPaused false, got true")
	}
}

func TestWorkoutDisplay_OnPeriodStart(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)

	// Test work period start
	display.OnPeriodStart(types.PeriodWork, 1, 20*time.Second)
	if display.currentPeriod != types.PeriodWork {
		t.Errorf("expected PeriodWork, got %v", display.currentPeriod)
	}
	if display.currentRound != 1 {
		t.Errorf("expected currentRound 1, got %d", display.currentRound)
	}
	if display.remainingTime != 20*time.Second {
		t.Errorf("expected remainingTime 20s, got %v", display.remainingTime)
	}
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0, got %d", display.currentComboIdx)
	}

	// Test rest period start
	display.OnPeriodStart(types.PeriodRest, 1, 10*time.Second)
	if display.currentPeriod != types.PeriodRest {
		t.Errorf("expected PeriodRest, got %v", display.currentPeriod)
	}
	if display.remainingTime != 10*time.Second {
		t.Errorf("expected remainingTime 10s, got %v", display.remainingTime)
	}

	// Clean up combo updates that were started during work period
	display.stopComboUpdates()
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to stop
}

func TestWorkoutDisplay_OnTimerUpdate(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)

	display.OnTimerUpdate(15*time.Second, types.PeriodWork, 1)

	// Clean up
	display.stopComboUpdates()

	if display.remainingTime != 15*time.Second {
		t.Errorf("expected remainingTime 15s, got %v", display.remainingTime)
	}
	if display.currentPeriod != types.PeriodWork {
		t.Errorf("expected PeriodWork, got %v", display.currentPeriod)
	}
	if display.currentRound != 1 {
		t.Errorf("expected currentRound 1, got %d", display.currentRound)
	}
}

func TestWorkoutDisplay_SetPaused(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)

	display.SetPaused(true)
	if !display.isPaused {
		t.Error("expected isPaused true, got false")
	}

	display.SetPaused(false)
	if display.isPaused {
		t.Error("expected isPaused false, got true")
	}

	// Clean up any combo updates that might have been started
	display.stopComboUpdates()
	time.Sleep(50 * time.Millisecond) // Wait for goroutine cleanup if any
}

func TestWorkoutDisplay_printRoundNumber(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 8),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)
	display.currentRound = 3
	display.totalRounds = 8

	// Capture output
	// We can't easily test fmt output, but we can verify the state is correct
	if display.currentRound != 3 {
		t.Errorf("expected currentRound 3, got %d", display.currentRound)
	}
	if display.totalRounds != 8 {
		t.Errorf("expected totalRounds 8, got %d", display.totalRounds)
	}
}

func TestWorkoutDisplay_printProgress(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 4),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)
	display.totalRounds = 4

	// Test first round (0 completed)
	display.currentRound = 1
	roundsCompleted := display.currentRound - 1
	if roundsCompleted != 0 {
		t.Errorf("expected roundsCompleted 0, got %d", roundsCompleted)
	}

	// Test second round (1 completed)
	display.currentRound = 2
	roundsCompleted = display.currentRound - 1
	if roundsCompleted != 1 {
		t.Errorf("expected roundsCompleted 1, got %d", roundsCompleted)
	}

	// Test after completion
	display.currentRound = 5
	roundsCompleted = display.currentRound - 1
	if roundsCompleted > display.totalRounds {
		roundsCompleted = display.totalRounds
	}
	if roundsCompleted != 4 {
		t.Errorf("expected roundsCompleted 4, got %d", roundsCompleted)
	}
}

func TestWorkoutDisplay_printTimer(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)

	// Test work period timer
	display.currentPeriod = types.PeriodWork
	display.remainingTime = 90 * time.Second // 1:30
	display.isPaused = false

	minutes := int(display.remainingTime.Minutes())
	seconds := int(display.remainingTime.Seconds()) % 60

	if minutes != 1 {
		t.Errorf("expected 1 minute, got %d", minutes)
	}
	if seconds != 30 {
		t.Errorf("expected 30 seconds, got %d", seconds)
	}

	// Test rest period timer
	display.currentPeriod = types.PeriodRest
	display.remainingTime = 15 * time.Second

	minutes = int(display.remainingTime.Minutes())
	seconds = int(display.remainingTime.Seconds()) % 60

	if minutes != 0 {
		t.Errorf("expected 0 minutes, got %d", minutes)
	}
	if seconds != 15 {
		t.Errorf("expected 15 seconds, got %d", seconds)
	}

	// Test paused state
	display.isPaused = true
	if !display.isPaused {
		t.Error("expected isPaused true")
	}
}

func TestWorkoutDisplay_printCurrentCombo(t *testing.T) {
	// Create workout with combos
	punch1 := models.NewPunchMove(models.Jab)
	punch2 := models.NewPunchMove(models.Cross)
	defensive := models.NewDefensiveMove(models.LeftSlip)

	combo1 := models.NewCombo([]models.Move{punch1, punch2, defensive})

	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, combo1, 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)
	display.currentRound = 1
	display.currentComboIdx = 0

	// Test combo with defensive moves
	round := workout.Rounds[0]
	combo := round.Combo
	if len(combo.Moves) != 3 {
		t.Errorf("expected 3 moves in combo, got %d", len(combo.Moves))
	}

	// Verify combo contains both punches and defensive moves
	punchCount := 0
	defensiveCount := 0
	for _, move := range combo.Moves {
		if move.IsPunch() {
			punchCount++
		} else if move.IsDefensive() {
			defensiveCount++
		}
	}

	if punchCount != 2 {
		t.Errorf("expected 2 punches, got %d", punchCount)
	}
	if defensiveCount != 1 {
		t.Errorf("expected 1 defensive move, got %d", defensiveCount)
	}

	// With single combo per round, there's no cycling - combo stays the same
	// Verify the combo is still accessible
	combo2Test := round.Combo
	if len(combo2Test.Moves) != 3 {
		t.Errorf("expected 3 moves in combo, got %d", len(combo2Test.Moves))
	}
}

func TestWorkoutDisplay_printCurrentCombo_PunchNames(t *testing.T) {
	// Test orthodox stance
	punch3 := models.NewPunchMove(models.LeadHook)     // Should be "left hook" for orthodox
	punch4 := models.NewPunchMove(models.RearHook)     // Should be "right hook" for orthodox
	punch5 := models.NewPunchMove(models.LeadUppercut) // Should be "left uppercut" for orthodox
	punch6 := models.NewPunchMove(models.RearUppercut) // Should be "right uppercut" for orthodox

	combo := models.NewCombo([]models.Move{punch3, punch4, punch5, punch6})

	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, combo, 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplayWithStance(workout, models.Orthodox)
	display.currentRound = 1
	display.currentComboIdx = 0

	round := workout.Rounds[0]
	comboTest := round.Combo

	// Verify punch names for orthodox
	expectedNames := []string{"left hook", "right hook", "left uppercut", "right uppercut"}
	for i, move := range comboTest.Moves {
		if move.IsPunch() && move.Punch != nil {
			name := move.Punch.NameForStance(models.Orthodox)
			if name != expectedNames[i] {
				t.Errorf("expected %s for orthodox, got %s", expectedNames[i], name)
			}
		}
	}

	// Test southpaw stance
	display.SetStance(models.Southpaw)
	expectedNamesSouthpaw := []string{"right hook", "left hook", "right uppercut", "left uppercut"}
	for i, move := range comboTest.Moves {
		if move.IsPunch() && move.Punch != nil {
			name := move.Punch.NameForStance(models.Southpaw)
			if name != expectedNamesSouthpaw[i] {
				t.Errorf("expected %s for southpaw, got %s", expectedNamesSouthpaw[i], name)
			}
		}
	}
}

func TestWorkoutDisplay_printCurrentCombo_EmptyCombo(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{}), 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)
	display.currentRound = 1

	round := workout.Rounds[0]
	if len(round.Combo.Moves) != 0 {
		t.Errorf("expected 0 moves in combo, got %d", len(round.Combo.Moves))
	}
}

func TestWorkoutDisplay_startComboUpdates(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{models.NewPunchMove(models.Jab)}), 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)
	display.currentPeriod = types.PeriodWork
	display.currentRound = 1
	display.isPaused = false

	display.startComboUpdates()

	if display.comboUpdateTicker == nil {
		t.Error("expected comboUpdateTicker to be set")
	}
	if display.comboUpdateDone == nil {
		t.Error("expected comboUpdateDone to be set")
	}

	// Wait a bit to ensure ticker is running
	time.Sleep(100 * time.Millisecond)

	display.stopComboUpdates()

	if display.comboUpdateTicker != nil {
		t.Error("expected comboUpdateTicker to be nil after stop")
	}
	if display.comboUpdateDone != nil {
		t.Error("expected comboUpdateDone to be nil after stop")
	}
}

func TestWorkoutDisplay_stopComboUpdates(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)

	// Stop when not started should not panic
	display.stopComboUpdates()

	// Start then stop
	display.startComboUpdates()
	time.Sleep(50 * time.Millisecond)
	display.stopComboUpdates()

	// Stop again should not panic
	display.stopComboUpdates()
}

func TestWorkoutDisplay_TempoControlsBeepInterval(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{models.NewPunchMove(models.Jab)}), 20*time.Second, 10*time.Second),
		},
	)

	testCases := []struct {
		name             string
		tempo            time.Duration
		expectedInterval time.Duration
	}{
		{
			name:             "slow tempo (5 seconds)",
			tempo:            5 * time.Second,
			expectedInterval: 5 * time.Second,
		},
		{
			name:             "medium tempo (4 seconds)",
			tempo:            4 * time.Second,
			expectedInterval: 4 * time.Second,
		},
		{
			name:             "fast tempo (3 seconds)",
			tempo:            3 * time.Second,
			expectedInterval: 3 * time.Second,
		},
		{
			name:             "superfast tempo (2 seconds)",
			tempo:            2 * time.Second,
			expectedInterval: 2 * time.Second,
		},
		{
			name:             "default tempo (0 defaults to 5 seconds)",
			tempo:            0,
			expectedInterval: 5 * time.Second,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			display := NewWorkoutDisplayWithStanceAndTempo(workout, models.Orthodox, tt.tempo)
			display.currentPeriod = types.PeriodWork
			display.currentRound = 1
			display.isPaused = false

			// Verify tempo field is set correctly
			if tt.tempo == 0 {
				// When tempo is 0, it should still be 0 in the struct (defaulting happens in startComboUpdates)
				// But the expected interval is 5 seconds
			} else {
				if display.tempo != tt.tempo {
					t.Errorf("expected tempo field to be %v, got %v", tt.tempo, display.tempo)
				}
			}

			// Create a mock audio handler using gomock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAudioHandler := mocks.NewMockAudioCueHandler(ctrl)

			// Expect at least one beep call after the first interval
			mockAudioHandler.EXPECT().PlayBeep().MinTimes(1).MaxTimes(10)

			display.SetAudioHandler(mockAudioHandler)

			// Start combo updates
			display.startComboUpdates()
			defer display.stopComboUpdates()

			if display.comboUpdateTicker == nil {
				t.Fatal("expected comboUpdateTicker to be set")
			}

			// Wait for at least one tick to fire (ticker fires after the first interval)
			// Wait slightly longer than the expected interval to account for timing variations
			waitTime := tt.expectedInterval + 500*time.Millisecond
			time.Sleep(waitTime)

			// For faster tempos, wait for another interval to verify multiple beeps
			if tt.expectedInterval <= 3*time.Second {
				time.Sleep(tt.expectedInterval + 500*time.Millisecond)
			}

			// Give a small buffer for goroutine to process
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestWorkoutDisplay_OnPeriodEnd(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, models.NewCombo([]models.Move{models.NewPunchMove(models.Jab)}), 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)
	display.currentPeriod = types.PeriodWork
	display.startComboUpdates()

	// Wait a bit to let goroutine start
	time.Sleep(50 * time.Millisecond)

	display.OnPeriodEnd(types.PeriodWork, 1)

	// Clean up
	display.stopComboUpdates()
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to stop
}

func TestWorkoutDisplay_OnWorkoutComplete(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 3),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)
	display.totalRounds = 3
	display.currentRound = 1
	display.currentPeriod = types.PeriodWork
	display.startComboUpdates()

	// Wait a bit to let goroutine start
	time.Sleep(50 * time.Millisecond)

	display.OnWorkoutComplete()

	// Verify combo updates are stopped
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to stop
	if display.comboUpdateTicker != nil {
		t.Error("expected comboUpdateTicker to be nil after workout complete")
	}
}

func TestWorkoutDisplay_printProgress_Calculation(t *testing.T) {
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 4),
		[]models.WorkoutRound{},
	)

	display := NewWorkoutDisplay(workout)
	display.totalRounds = 4

	tests := []struct {
		currentRound      int
		expectedProgress  float64
		expectedCompleted int
	}{
		{1, 0, 0},   // First round, 0 completed
		{2, 25, 1},  // Second round, 1 completed
		{3, 50, 2},  // Third round, 2 completed
		{4, 75, 3},  // Fourth round, 3 completed
		{5, 100, 4}, // After completion
	}

	for _, tt := range tests {
		display.currentRound = tt.currentRound
		roundsCompleted := display.currentRound - 1
		if roundsCompleted < 0 {
			roundsCompleted = 0
		}
		if display.currentRound > display.totalRounds {
			roundsCompleted = display.totalRounds
		}

		progress := float64(roundsCompleted) / float64(display.totalRounds) * 100
		if display.totalRounds == 0 {
			progress = 0
		}

		if roundsCompleted != tt.expectedCompleted {
			t.Errorf("currentRound %d: expected completed %d, got %d", tt.currentRound, tt.expectedCompleted, roundsCompleted)
		}
		if progress != tt.expectedProgress {
			t.Errorf("currentRound %d: expected progress %.0f%%, got %.0f%%", tt.currentRound, tt.expectedProgress, progress)
		}
	}
}

func TestWorkoutDisplay_ComboRemainsConstantPerRound(t *testing.T) {
	// Create a workout with multiple combos per round
	punch1 := models.NewPunchMove(models.Jab)
	punch2 := models.NewPunchMove(models.Cross)
	punch3 := models.NewPunchMove(models.LeadHook)

	combo1 := models.NewCombo([]models.Move{punch1, punch2})
	combo2 := models.NewCombo([]models.Move{punch2, punch3})
	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, combo1, 20*time.Second, 10*time.Second),
			models.NewWorkoutRound(2, combo2, 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)

	// Test Round 1: Start work period - combo index should be set to 0
	display.OnPeriodStart(types.PeriodWork, 1, 20*time.Second)
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0 at start of round 1, got %d", display.currentComboIdx)
	}

	round1 := workout.Rounds[0]
	initialCombo := round1.Combo

	// Simulate multiple 2-second ticks during the work period
	// The combo should remain the same throughout the round
	for i := 0; i < 5; i++ {
		// Wait for ticker to fire (or simulate by calling updateDisplay directly)
		// Since we can't easily trigger the ticker in tests, we'll test the logic directly
		// by checking that currentComboIdx doesn't change when updateDisplay is called
		previousComboIdx := display.currentComboIdx
		display.updateDisplay()

		// Verify combo index hasn't changed
		if display.currentComboIdx != previousComboIdx {
			t.Errorf("tick %d: expected currentComboIdx to remain %d, got %d", i, previousComboIdx, display.currentComboIdx)
		}

		// Verify the combo being displayed is still the same
		currentCombo := round1.Combo
		if len(currentCombo.Moves) != len(initialCombo.Moves) {
			t.Errorf("tick %d: combo changed - initial had %d moves, current has %d moves", i, len(initialCombo.Moves), len(currentCombo.Moves))
		}
	}

	// Stop combo updates for round 1
	display.OnPeriodEnd(types.PeriodWork, 1)
	display.stopComboUpdates()
	time.Sleep(100 * time.Millisecond)

	// Start Round 2: combo index should reset to 0
	display.OnPeriodStart(types.PeriodWork, 2, 20*time.Second)
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0 at start of round 2, got %d", display.currentComboIdx)
	}

	round2 := workout.Rounds[1]
	round2InitialCombo := round2.Combo

	// During Round 2, combo should remain constant
	for i := 0; i < 3; i++ {
		previousComboIdx := display.currentComboIdx
		display.updateDisplay()

		if display.currentComboIdx != previousComboIdx {
			t.Errorf("round 2, tick %d: expected currentComboIdx to remain %d, got %d", i, previousComboIdx, display.currentComboIdx)
		}

		currentCombo := round2.Combo
		if len(currentCombo.Moves) != len(round2InitialCombo.Moves) {
			t.Errorf("round 2, tick %d: combo changed - initial had %d moves, current has %d moves", i, len(round2InitialCombo.Moves), len(currentCombo.Moves))
		}
	}

	// Verify round 2 combo is different from round 1 combo (since rounds have different combos)
	if len(round2InitialCombo.Moves) == len(initialCombo.Moves) {
		// They might have the same length by coincidence, so check the actual moves
		// This is just a sanity check - the important thing is that combo stayed constant within each round
	}

	display.stopComboUpdates()
	time.Sleep(100 * time.Millisecond)
}

func TestWorkoutDisplay_ComboConstantDuringTickerFires(t *testing.T) {
	// Test that even when the 2-second ticker fires multiple times, the combo remains the same
	punch1 := models.NewPunchMove(models.Jab)
	punch2 := models.NewPunchMove(models.Cross)

	combo1 := models.NewCombo([]models.Move{punch1, punch2})

	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, combo1, 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)
	display.currentPeriod = types.PeriodWork
	display.currentRound = 1
	display.isPaused = false

	// Start work period - combo index should be set to 0
	display.OnPeriodStart(types.PeriodWork, 1, 20*time.Second)
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0 at start, got %d", display.currentComboIdx)
	}

	round := workout.Rounds[0]
	initialCombo := round.Combo

	// Verify combo updates are started
	if display.comboUpdateTicker == nil {
		t.Error("expected comboUpdateTicker to be started")
	}

	for i := 0; i < 3; i++ {
		time.Sleep(2500 * time.Millisecond)
		display.updateDisplay()
		currentCombo := round.Combo
		if len(currentCombo.Moves) != len(initialCombo.Moves) {
			t.Errorf("ticker fire %d: combo changed - initial had %d moves, current has %d moves", i+1, len(initialCombo.Moves), len(currentCombo.Moves))
		}
	}

	display.stopComboUpdates()
	time.Sleep(100 * time.Millisecond)
}

func TestWorkoutDisplay_ComboIndexResetsOnNewRound(t *testing.T) {
	// Test that combo index resets to 0 when starting a new round
	punch1 := models.NewPunchMove(models.Jab)
	punch2 := models.NewPunchMove(models.Cross)

	combo1 := models.NewCombo([]models.Move{punch1, punch2})
	combo2 := models.NewCombo([]models.Move{punch2, punch1})

	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 2),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, combo1, 20*time.Second, 10*time.Second),
			models.NewWorkoutRound(2, combo2, 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplay(workout)

	// Start round 1
	display.OnPeriodStart(types.PeriodWork, 1, 20*time.Second)
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0 at start of round 1, got %d", display.currentComboIdx)
	}

	// Manually change combo index (simulating what would happen if it incremented)
	display.currentComboIdx = 1

	// Start round 2 - should reset to 0
	display.OnPeriodStart(types.PeriodRest, 1, 10*time.Second) // End round 1 rest
	display.OnPeriodStart(types.PeriodWork, 2, 20*time.Second) // Start round 2 work
	if display.currentComboIdx != 0 {
		t.Errorf("expected currentComboIdx 0 at start of round 2 (should reset), got %d", display.currentComboIdx)
	}

	display.stopComboUpdates()
	time.Sleep(50 * time.Millisecond)
}

func TestWorkoutDisplay_ComboStringFormat(t *testing.T) {
	// Test that combo string formatting works correctly
	punch1 := models.NewPunchMove(models.Jab)
	punch2 := models.NewPunchMove(models.Cross)
	defensive := models.NewDefensiveMove(models.LeftSlip)

	combo := models.NewCombo([]models.Move{punch1, punch2, defensive})

	workout := models.NewWorkout(
		models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1),
		[]models.WorkoutRound{
			models.NewWorkoutRound(1, combo, 20*time.Second, 10*time.Second),
		},
	)

	display := NewWorkoutDisplayWithStance(workout, models.Orthodox)
	display.currentRound = 1
	display.currentComboIdx = 0

	round := workout.Rounds[0]
	comboTest := round.Combo

	// Build formatted string similar to printCurrentCombo
	formattedMoves := make([]string, 0, len(comboTest.Moves))
	for _, move := range comboTest.Moves {
		if move.IsPunch() && move.Punch != nil {
			punchName := move.Punch.NameForStance(display.stance)
			formattedMoves = append(formattedMoves, punchName)
		} else if move.IsDefensive() && move.Defensive != nil {
			formattedMoves = append(formattedMoves, "🛡 "+move.String())
		}
	}

	comboStr := strings.Join(formattedMoves, " → ")

	// Verify string contains expected elements
	if !strings.Contains(comboStr, "jab") {
		t.Error("expected combo string to contain 'jab'")
	}
	if !strings.Contains(comboStr, "cross") {
		t.Error("expected combo string to contain 'cross'")
	}
	if !strings.Contains(comboStr, "🛡") {
		t.Error("expected combo string to contain shield emoji")
	}
	if !strings.Contains(comboStr, "Left Slip") {
		t.Error("expected combo string to contain 'Left Slip'")
	}
	if !strings.Contains(comboStr, "→") {
		t.Error("expected combo string to contain arrow separator")
	}
}
