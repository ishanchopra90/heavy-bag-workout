package cli

import (
	"bufio"
	"fmt"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/timer"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WorkoutInterface manages the CLI interface for running workouts
type WorkoutInterface struct {
	workout      models.Workout
	workoutTimer *timer.WorkoutTimer
	display      *WorkoutDisplay
	audioHandler timer.AudioCueHandler
	quitChan     chan bool
}

// NewWorkoutInterface creates a new workout interface with orthodox stance (default) and slow tempo (5 seconds)
func NewWorkoutInterface(workout models.Workout, audioHandler timer.AudioCueHandler) *WorkoutInterface {
	return NewWorkoutInterfaceWithStance(workout, audioHandler, models.Orthodox)
}

// NewWorkoutInterfaceWithStance creates a new workout interface with the specified stance and slow tempo (5 seconds)
func NewWorkoutInterfaceWithStance(workout models.Workout, audioHandler timer.AudioCueHandler, stance models.Stance) *WorkoutInterface {
	return NewWorkoutInterfaceWithStanceAndTempo(workout, audioHandler, stance, 5*time.Second)
}

// NewWorkoutInterfaceWithStanceAndTempo creates a new workout interface with the specified stance and tempo
func NewWorkoutInterfaceWithStanceAndTempo(workout models.Workout, audioHandler timer.AudioCueHandler, stance models.Stance, tempo time.Duration) *WorkoutInterface {
	wt := timer.NewWorkoutTimer(workout)
	wt.SetStance(stance) // Set stance for combo callouts
	return &WorkoutInterface{
		workout:      workout,
		workoutTimer: wt,
		display:      NewWorkoutDisplayWithStanceAndTempo(workout, stance, tempo),
		audioHandler: audioHandler,
		quitChan:     make(chan bool, 1),
	}
}

// SetStance sets the boxer's stance
func (wi *WorkoutInterface) SetStance(stance models.Stance) {
	wi.display.SetStance(stance)
	wi.workoutTimer.SetStance(stance)
}

// Run starts the workout interface and handles user input
func (wi *WorkoutInterface) Run() error {
	// Set up display handler
	wi.workoutTimer.SetDisplayHandler(wi.display)

	// Set up audio handler if provided
	if wi.audioHandler != nil {
		wi.workoutTimer.SetAudioHandler(wi.audioHandler)
		// Also set audio handler in display for 2-second beeps
		wi.display.SetAudioHandler(wi.audioHandler)
	}

	// Set up callbacks
	wi.workoutTimer.OnWorkoutComplete(func() {
		wi.quitChan <- true
	})

	// Print initial display
	wi.display.PrintInitialDisplay()

	// Wait for user to press Enter to view preview
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	input, _ := reader.ReadString('\n')
	if len(input) > 0 && (input[0] == 'q' || input[0] == 'Q') {
		fmt.Println("\nWorkout cancelled.")
		return nil
	}

	// Show workout preview
	wi.display.PrintWorkoutPreview()

	// Wait for user to press Enter to start
	fmt.Print("> ")
	input, _ = reader.ReadString('\n')
	if len(input) > 0 && (input[0] == 'q' || input[0] == 'Q') {
		fmt.Println("\nWorkout cancelled.")
		return nil
	}

	// Start the workout
	if err := wi.workoutTimer.Start(); err != nil {
		return fmt.Errorf("failed to start workout: %w", err)
	}

	// Handle keyboard input in a goroutine
	go wi.handleInput(reader)

	// Handle system signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for quit signal or completion
	select {
	case <-wi.quitChan:
		// Workout completed normally
		return nil
	case <-sigChan:
		// Interrupted by user (Ctrl+C)
		fmt.Println("\n\nWorkout interrupted.")
		wi.workoutTimer.Stop()
		return nil
	}
}

// handleInput handles keyboard input during the workout
func (wi *WorkoutInterface) handleInput(reader *bufio.Reader) {
	for {
		// Set terminal to raw mode would be better, but for simplicity
		// we'll use a simpler approach with ReadString
		// Note: This is a simplified version. For better UX, consider using
		// a library like github.com/eiannone/keyboard for non-blocking input

		// For now, we'll check for input in a non-blocking way
		// This is a basic implementation - in production, you'd want
		// to use a proper terminal input library

		select {
		case <-wi.quitChan:
			return
		default:
			// Try to read input (this will block, so we need a better approach)
			// For now, we'll handle this differently
		}
	}
}

// Pause pauses the workout
func (wi *WorkoutInterface) Pause() {
	wi.workoutTimer.Pause()
	wi.display.SetPaused(true)
}

// Resume resumes the workout
func (wi *WorkoutInterface) Resume() error {
	if err := wi.workoutTimer.Resume(); err != nil {
		return err
	}
	wi.display.SetPaused(false)
	return nil
}

// Stop stops the workout
func (wi *WorkoutInterface) Stop() {
	wi.workoutTimer.Stop()
	wi.quitChan <- true
}

// TogglePause toggles between pause and resume
func (wi *WorkoutInterface) TogglePause() error {
	if wi.display.isPaused {
		return wi.Resume()
	}
	wi.Pause()
	return nil
}
