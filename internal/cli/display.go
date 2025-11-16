package cli

import (
	"fmt"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/timer"
	"heavybagworkout/internal/types"
	"strings"
	"time"
)

// WorkoutDisplay handles the CLI display of workout information
type WorkoutDisplay struct {
	workout           models.Workout
	stance            models.Stance
	currentComboIdx   int
	totalRounds       int
	currentRound      int
	currentPeriod     types.PeriodType
	remainingTime     time.Duration
	isPaused          bool
	comboUpdateTicker *time.Ticker
	comboUpdateDone   chan bool
	audioHandler      timer.AudioCueHandler // Audio handler for beeps
	tempo             time.Duration         // Interval between beeps during work period
}

// NewWorkoutDisplay creates a new workout display with orthodox stance (default) and slow tempo (5 seconds)
func NewWorkoutDisplay(workout models.Workout) *WorkoutDisplay {
	return NewWorkoutDisplayWithStance(workout, models.Orthodox)
}

// NewWorkoutDisplayWithStance creates a new workout display with the specified stance and slow tempo (5 seconds)
func NewWorkoutDisplayWithStance(workout models.Workout, stance models.Stance) *WorkoutDisplay {
	return NewWorkoutDisplayWithStanceAndTempo(workout, stance, 5*time.Second)
}

// NewWorkoutDisplayWithStanceAndTempo creates a new workout display with the specified stance and tempo
func NewWorkoutDisplayWithStanceAndTempo(workout models.Workout, stance models.Stance, tempo time.Duration) *WorkoutDisplay {
	return &WorkoutDisplay{
		workout:      workout,
		stance:       stance,
		totalRounds:  len(workout.Rounds),
		currentRound: 0,
		tempo:        tempo,
	}
}

// SetAudioHandler sets the audio handler for 2-second beeps during work periods
func (wd *WorkoutDisplay) SetAudioHandler(handler timer.AudioCueHandler) {
	wd.audioHandler = handler
}

// SetStance sets the boxer's stance
func (wd *WorkoutDisplay) SetStance(stance models.Stance) {
	wd.stance = stance
}

// OnWorkoutStart is called when the workout starts
func (wd *WorkoutDisplay) OnWorkoutStart(totalRounds int) {
	wd.totalRounds = totalRounds
	wd.currentRound = 1
	wd.currentComboIdx = 0
	wd.isPaused = false
	wd.clearScreen()
	wd.printHeader()
	wd.printInstructions()
}

// OnPeriodStart is called when a period (work or rest) starts
func (wd *WorkoutDisplay) OnPeriodStart(periodType types.PeriodType, roundNumber int, duration time.Duration) {
	wd.currentPeriod = periodType
	wd.currentRound = roundNumber
	wd.remainingTime = duration
	wd.currentComboIdx = 0

	if periodType == types.PeriodWork {
		wd.startComboUpdates()
		wd.printWorkPeriodStart()
	} else {
		wd.stopComboUpdates()
		wd.printRestPeriodStart()
	}
}

// OnTimerUpdate is called on each timer tick
func (wd *WorkoutDisplay) OnTimerUpdate(remaining time.Duration, periodType types.PeriodType, roundNumber int) {
	wd.remainingTime = remaining
	wd.currentPeriod = periodType
	wd.currentRound = roundNumber
	wd.updateDisplay()
}

// OnPeriodEnd is called when a period ends
func (wd *WorkoutDisplay) OnPeriodEnd(periodType types.PeriodType, roundNumber int) {
	if periodType == types.PeriodWork {
		wd.stopComboUpdates()
	}
}

// OnWorkoutComplete is called when the workout completes
func (wd *WorkoutDisplay) OnWorkoutComplete() {
	wd.stopComboUpdates()
	wd.clearScreen()
	wd.printWorkoutComplete()
}

// SetPaused sets the paused state
func (wd *WorkoutDisplay) SetPaused(paused bool) {
	wd.isPaused = paused
}

// clearScreen clears the terminal screen
func (wd *WorkoutDisplay) clearScreen() {
	fmt.Print("\033[2J\033[H") // ANSI escape codes to clear screen and move cursor to top
}

// printHeader prints the workout header
func (wd *WorkoutDisplay) printHeader() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         🐶 Puppy Power - Heavy Bag Workout 🐶             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// printInstructions prints keyboard instructions
func (wd *WorkoutDisplay) printInstructions() {
	fmt.Println("Use Ctrl+C to cancel workout")
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println()
}

// printWorkPeriodStart prints the start of a work period
func (wd *WorkoutDisplay) printWorkPeriodStart() {
	fmt.Printf("🔥 ROUND %d - WORK PERIOD 🔥\n", wd.currentRound)
	fmt.Println()
}

// printRestPeriodStart prints the start of a rest period
func (wd *WorkoutDisplay) printRestPeriodStart() {
	fmt.Printf("💤 ROUND %d - REST PERIOD 💤\n", wd.currentRound)
	fmt.Println()
}

// updateDisplay updates the main display area
func (wd *WorkoutDisplay) updateDisplay() {
	// Clear and redraw the main content area
	// We'll use ANSI escape codes to position cursor
	// Save cursor position, move to start of content area, clear, then restore
	fmt.Print("\033[s") // Save cursor position

	// Move to line after instructions (approximately line 8)
	fmt.Print("\033[8;1H") // Move to line 8, column 1

	// Clear from cursor to end of screen
	fmt.Print("\033[J")

	// Print round number prominently
	wd.printRoundNumber()

	// Print progress bar
	wd.printProgress()

	// Print timer
	wd.printTimer()

	// Print combo (only during work period)
	if wd.currentPeriod == types.PeriodWork {
		wd.printCurrentCombo()
	} else {
		fmt.Println()
		fmt.Println("  Rest and recover...")
		fmt.Println()
	}

	// Print status (paused indicator)
	wd.printStatus()

	// Restore cursor position
	fmt.Print("\033[u")
}

// printRoundNumber prints the current round number prominently
func (wd *WorkoutDisplay) printRoundNumber() {
	fmt.Println()
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("                      ROUND %d of %d\n", wd.currentRound, wd.totalRounds)
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Println()
}

// printProgress prints the workout progress with rounds completed/total
func (wd *WorkoutDisplay) printProgress() {
	// Calculate rounds completed (current round - 1, since we're currently in a round)
	roundsCompleted := wd.currentRound - 1
	if roundsCompleted < 0 {
		roundsCompleted = 0
	}
	if wd.currentRound > wd.totalRounds {
		roundsCompleted = wd.totalRounds
	}

	progress := float64(roundsCompleted) / float64(wd.totalRounds) * 100
	if wd.totalRounds == 0 {
		progress = 0
	}

	barWidth := 50
	filled := int(progress / 100 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("Progress: [%s] %.0f%% (%d/%d rounds completed)\n", bar, progress, roundsCompleted, wd.totalRounds)
}

// printTimer prints the countdown timer prominently
func (wd *WorkoutDisplay) printTimer() {
	minutes := int(wd.remainingTime.Minutes())
	seconds := int(wd.remainingTime.Seconds()) % 60
	totalSeconds := int(wd.remainingTime.Seconds())

	var periodLabel string
	var periodEmoji string
	if wd.currentPeriod == types.PeriodWork {
		periodLabel = "WORK"
		periodEmoji = "🔥"
	} else {
		periodLabel = "REST"
		periodEmoji = "💤"
	}

	var pauseLabel string
	if wd.isPaused {
		pauseLabel = " ⏸ PAUSED"
	}

	// Display timer prominently with visual emphasis
	// Format: "🔥 WORK" or "💤 REST" on one line, then large timer
	fmt.Printf("%s %s%s\n", periodEmoji, periodLabel, pauseLabel)

	// Display countdown in large format (MM:SS)
	// Add visual bars around timer for emphasis
	fmt.Printf("┌─────────────────┐\n")
	fmt.Printf("│  ⏱  %02d:%02d  │\n", minutes, seconds)
	fmt.Printf("└─────────────────┘\n")

	// Show seconds remaining as a secondary indicator for quick reference
	if totalSeconds > 0 {
		fmt.Printf("   (%d seconds remaining)\n", totalSeconds)
	}

	fmt.Println()
}

// printCurrentCombo prints the current combo to perform
func (wd *WorkoutDisplay) printCurrentCombo() {
	if wd.currentRound < 1 || wd.currentRound > len(wd.workout.Rounds) || len(wd.workout.Rounds) == 0 {
		return
	}

	round := wd.workout.Rounds[wd.currentRound-1]
	combo := round.Combo

	fmt.Println("  🥊 COMBO:")

	// Display combo in a formatted way, highlighting defensive moves
	if len(combo.Moves) == 0 {
		fmt.Println("     (empty combo)")
		fmt.Println()
		return
	}

	// Build formatted combo string with visual distinction for defensive moves
	formattedMoves := make([]string, 0, len(combo.Moves))
	for _, move := range combo.Moves {
		if move.IsPunch() && move.Punch != nil {
			// Punches shown as names based on stance
			punchName := move.Punch.NameForStance(wd.stance)
			formattedMoves = append(formattedMoves, punchName)
		} else if move.IsDefensive() && move.Defensive != nil {
			// Defensive moves shown with shield emoji and name
			formattedMoves = append(formattedMoves, fmt.Sprintf("🛡 %s", move.String()))
		} else {
			formattedMoves = append(formattedMoves, move.String())
		}
	}

	// Display as a sequence with separators
	comboStr := strings.Join(formattedMoves, " → ")
	fmt.Printf("     %s\n", comboStr)

	// Show move count breakdown if there are defensive moves
	hasDefensive := false
	punchCount := 0
	defensiveCount := 0
	for _, move := range combo.Moves {
		if move.IsPunch() {
			punchCount++
		} else if move.IsDefensive() {
			hasDefensive = true
			defensiveCount++
		}
	}

	if hasDefensive && punchCount > 0 {
		fmt.Printf("     (%d punches, %d defensive moves)\n", punchCount, defensiveCount)
	}

	fmt.Println()
}

// printStatus prints the current status
func (wd *WorkoutDisplay) printStatus() {
	// Status display (currently no status to show)
}

// startComboUpdates starts a ticker that plays a beep at the configured tempo interval during work periods
// The combo remains the same throughout the round - this is just a reminder cue
func (wd *WorkoutDisplay) startComboUpdates() {
	wd.stopComboUpdates() // Stop any existing ticker

	// Use tempo if set, otherwise default to 5 seconds
	tempoInterval := wd.tempo
	if tempoInterval == 0 {
		tempoInterval = 5 * time.Second
	}

	wd.comboUpdateTicker = time.NewTicker(tempoInterval)
	wd.comboUpdateDone = make(chan bool, 1) // Buffered channel to allow non-blocking send

	// Capture channel references before starting goroutine to avoid nil pointer issues
	tickerChan := wd.comboUpdateTicker.C
	doneChan := wd.comboUpdateDone

	go func() {
		for {
			select {
			case <-tickerChan:
				if wd.currentPeriod == types.PeriodWork && !wd.isPaused {
					// Play beep as a reminder to execute the combo again
					// The combo stays the same throughout the round
					if wd.audioHandler != nil {
						wd.audioHandler.PlayBeep()
					} else {
						fmt.Print("\a") // Fallback to system bell if no audio handler
					}
					// Refresh display (combo remains the same)
					wd.updateDisplay()
				}
			case <-doneChan:
				return
			}
		}
	}()
}

// stopComboUpdates stops the combo update ticker
func (wd *WorkoutDisplay) stopComboUpdates() {
	if wd.comboUpdateTicker != nil {
		wd.comboUpdateTicker.Stop()

		// Drain any pending values from the ticker channel to prevent
		// the goroutine from processing stale ticks after we signal shutdown
		for {
			select {
			case <-wd.comboUpdateTicker.C:
				// Drain one pending tick value
			default:
				// No more pending values, done draining
				goto tickerDrained
			}
		}
	tickerDrained:
		// Safe to set to nil now since goroutine uses captured tickerChan reference
		wd.comboUpdateTicker = nil
	}
	if wd.comboUpdateDone != nil {
		// Send shutdown signal to goroutine (non-blocking with buffered channel)
		select {
		case wd.comboUpdateDone <- true:
			// Shutdown signal sent successfully
		default:
			// Channel already has a value (already signaled), skip
		}
		// Safe to set to nil now since goroutine uses captured doneChan reference
		wd.comboUpdateDone = nil
	}
}

// printWorkoutComplete prints the workout completion message
func (wd *WorkoutDisplay) printWorkoutComplete() {
	wd.printHeader()
	fmt.Println()
	fmt.Println("  🎉 WORKOUT COMPLETE! 🎉")
	fmt.Println()
	fmt.Printf("  Completed %d rounds\n", wd.totalRounds)
	fmt.Println()
	fmt.Println("  Great job! You did it! 💪")
	fmt.Println()
}

// PrintInitialDisplay prints the initial display before workout starts
func (wd *WorkoutDisplay) PrintInitialDisplay() {
	wd.clearScreen()
	wd.printHeader()
	fmt.Println("Workout Configuration:")
	fmt.Printf("  Total Rounds: %d\n", wd.totalRounds)
	if len(wd.workout.Rounds) > 0 {
		fmt.Printf("  Work Duration: %.0f seconds\n", wd.workout.Rounds[0].WorkDuration.Seconds())
		fmt.Printf("  Rest Duration: %.0f seconds\n", wd.workout.Rounds[0].RestDuration.Seconds())
	}
	fmt.Println()
	fmt.Println("Press [Enter] to view workout preview, or [Q] to quit...")
}

// PrintWorkoutPreview displays all combos for all rounds before the workout starts
func (wd *WorkoutDisplay) PrintWorkoutPreview() {
	wd.clearScreen()
	wd.printHeader()
	fmt.Println("WORKOUT PREVIEW")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Total Rounds: %d\n", wd.totalRounds)
	if len(wd.workout.Rounds) > 0 {
		fmt.Printf("Work Duration: %.0f seconds per round\n", wd.workout.Rounds[0].WorkDuration.Seconds())
		fmt.Printf("Rest Duration: %.0f seconds per round\n", wd.workout.Rounds[0].RestDuration.Seconds())
	}
	fmt.Println()
	fmt.Println("Round-by-Round Combos:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println()

	if len(wd.workout.Rounds) == 0 {
		fmt.Println("  No rounds in this workout.")
		fmt.Println()
	} else {
		for i, round := range wd.workout.Rounds {
			fmt.Printf("Round %d:\n", round.RoundNumber)

			combo := round.Combo

			if len(combo.Moves) == 0 {
				fmt.Println("  (empty combo)")
				fmt.Println()
				continue
			}

			// Build formatted combo string
			formattedMoves := make([]string, 0, len(combo.Moves))
			for _, move := range combo.Moves {
				if move.IsPunch() && move.Punch != nil {
					punchName := move.Punch.NameForStance(wd.stance)
					formattedMoves = append(formattedMoves, punchName)
				} else if move.IsDefensive() && move.Defensive != nil {
					formattedMoves = append(formattedMoves, fmt.Sprintf("🛡 %s", move.String()))
				} else {
					formattedMoves = append(formattedMoves, move.String())
				}
			}

			comboStr := strings.Join(formattedMoves, " → ")
			fmt.Printf("  %s\n", comboStr)

			// Show move count breakdown if there are defensive moves
			hasDefensive := false
			punchCount := 0
			defensiveCount := 0
			for _, move := range combo.Moves {
				if move.IsPunch() {
					punchCount++
				} else if move.IsDefensive() {
					hasDefensive = true
					defensiveCount++
				}
			}

			if hasDefensive && punchCount > 0 {
				fmt.Printf("    (%d punches, %d defensive moves)\n", punchCount, defensiveCount)
			}

			fmt.Println()

			// Add separator between rounds (except for last round)
			if i < len(wd.workout.Rounds)-1 {
				fmt.Println("  ──────────────────────────────────────────────────────────────")
				fmt.Println()
			}
		}
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Press [Enter] to start the workout, or [Q] to quit...")
}
