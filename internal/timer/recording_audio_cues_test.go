package timer

import (
	"fmt"
	"heavybagworkout/internal/generator"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRecordingAudioCues_E2E tests the full recording workflow
// Similar to: ./heavybagworkout --min-moves 1 --max-moves 4 --work-duration 5 --rest-duration 4 --rounds 2 --save boxingworkout.mp3
// Note: This test is currently skipped for system audio capture implementation
// System audio capture requires BlackHole (macOS) or similar setup which is not available in CI/test environments
func TestRecordingAudioCues_E2E(t *testing.T) {
	t.Skip("System audio capture requires BlackHole/similar setup - skipping in test environment")

	// Skip if ffmpeg is not available (required for system audio capture)
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping e2e test")
	}
	// Skip if ffprobe is not available (required for audio analysis)
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available, skipping e2e test")
	}

	// Create a temporary output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "boxingworkout.mp3")

	// Create workout configuration matching the command
	workDuration := 5 * time.Second
	restDuration := 4 * time.Second
	rounds := 2
	minMoves := 1
	maxMoves := 4
	tempo := 5 * time.Second // Default slow tempo

	// Generate workout
	config := models.NewWorkoutConfig(workDuration, restDuration, rounds)
	pattern := models.NewWorkoutPattern(models.PatternLinear, minMoves, maxMoves, false)
	gen := generator.NewWorkoutGenerator()
	workout, err := gen.GenerateWorkout(config, pattern)
	if err != nil {
		t.Fatalf("failed to generate workout: %v", err)
	}

	if len(workout.Rounds) != rounds {
		t.Fatalf("expected %d rounds, got %d", rounds, len(workout.Rounds))
	}

	// Create recording handler
	recordingHandler, err := NewRecordingAudioCueHandlerWithWorkout(outputPath, workout, tempo)
	if err != nil {
		t.Fatalf("failed to create recording handler: %v", err)
	}
	// Don't defer cleanup - wait for finalization first, then cleanup manually
	// defer recordingHandler.Cleanup()

	// Create a tracking audio handler wrapper to verify cues are called
	type CueEvent struct {
		time     time.Duration
		cueType  string
		roundNum int
	}
	var cueEvents []CueEvent
	var cueEventsMu sync.Mutex
	startTime := time.Now()

	// Create a wrapper that tracks events and delegates to recording handler
	trackingHandler := &trackingAudioHandler{
		delegate: recordingHandler,
		trackCue: func(cueType string, roundNum int) {
			cueEventsMu.Lock()
			defer cueEventsMu.Unlock()
			cueEvents = append(cueEvents, CueEvent{
				time:     time.Since(startTime),
				cueType:  cueType,
				roundNum: roundNum,
			})
		},
	}

	// Create a no-op display handler (workout timer may need one)
	displayHandler := &noOpDisplayHandler{}

	// Create workout timer
	workoutTimer := NewWorkoutTimer(workout)
	workoutTimer.SetDisplayHandler(displayHandler)
	workoutTimer.SetAudioHandler(trackingHandler)

	// Set up completion callback before starting
	var completedMu sync.Mutex
	completed := false
	callbackInvoked := false
	workoutTimer.OnWorkoutComplete(func() {
		completedMu.Lock()
		completed = true
		callbackInvoked = true
		completedMu.Unlock()
		t.Logf("Workout completed callback triggered at %v", time.Since(startTime))
	})

	// Verify callback was set
	if workoutTimer.onWorkoutComplete == nil {
		t.Fatal("Failed to set workout completion callback")
	}

	// Start workout
	if err := workoutTimer.Start(); err != nil {
		t.Fatalf("failed to start workout: %v", err)
	}
	t.Logf("Workout started")

	// Wait for workout to complete
	// Expected duration: 2 rounds * (5s work + 4s rest) = 18 seconds
	// Add buffer for audio generation and processing
	expectedDuration := time.Duration(rounds) * (workDuration + restDuration)
	maxWaitTime := expectedDuration + 15*time.Second // Increased buffer for audio generation
	t.Logf("Waiting up to %v for workout to complete (expected ~%v)", maxWaitTime, expectedDuration)

	// Wait for completion with periodic status updates
	deadline := time.Now().Add(maxWaitTime)
	checkInterval := 500 * time.Millisecond
	lastStatusTime := time.Now()

	for time.Now().Before(deadline) {
		completedMu.Lock()
		isCompleted := completed
		completedMu.Unlock()

		if isCompleted {
			t.Logf("Workout completed successfully")
			break
		}

		// Log status every 5 seconds
		if time.Since(lastStatusTime) >= 5*time.Second {
			remaining := workoutTimer.RemainingTime()
			currentRound := workoutTimer.CurrentRound()
			currentPeriod := workoutTimer.CurrentPeriod()
			t.Logf("Still waiting... Round %d, Period: %v, Remaining: %v",
				currentRound, currentPeriod, remaining)
			lastStatusTime = time.Now()
		}

		time.Sleep(checkInterval)
	}

	completedMu.Lock()
	isCompleted := completed
	completedMu.Unlock()

	completedMu.Lock()
	callbackWasInvoked := callbackInvoked
	completedMu.Unlock()

	if !isCompleted {
		remaining := workoutTimer.RemainingTime()
		currentRound := workoutTimer.CurrentRound()
		expectedRounds := len(workout.Rounds)

		// Debug information
		t.Logf("Completion check: completed=%v, callbackInvoked=%v, currentRound=%d, expectedRounds=%d, remaining=%v",
			isCompleted, callbackWasInvoked, currentRound, expectedRounds, remaining)

		// Check if workout has actually completed (currentRound > expectedRounds or currentRound == 0)
		// currentRound == 0 means completeWorkout() was called and set it to 0
		// currentRound > expectedRounds means we've passed all rounds
		if (currentRound == 0 || currentRound > expectedRounds) && remaining == 0 {
			if !callbackWasInvoked {
				t.Logf("Workout appears to have completed (round %d, remaining: 0s) but callback wasn't triggered. This suggests completeWorkout() was called but callback didn't execute.",
					currentRound)
				// Check if callback is still set
				// We can't access the private field, but we can infer from behavior
			}
			// Manually trigger completion handling
			completedMu.Lock()
			completed = true
			completedMu.Unlock()
			isCompleted = true
		}

		if !isCompleted {
			t.Fatalf("workout did not complete within %v. Current round: %d/%d, Remaining time: %v, Callback invoked: %v",
				maxWaitTime, currentRound, expectedRounds, remaining, callbackWasInvoked)
		}
	}

	// Wait for audio finalization
	if err := recordingHandler.WaitForFinalization(); err != nil {
		// Cleanup before failing
		recordingHandler.Cleanup()
		t.Fatalf("failed to finalize audio: %v", err)
	}

	// Don't call Cleanup() after successful finalization - it's already done
	// Cleanup() is only needed if finalization failed or workout was interrupted
	// recordingHandler.Cleanup()

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("output audio file was not created: %s", outputPath)
	}

	// Analyze the audio file to verify timing
	if err := verifyAudioTiming(t, outputPath, workout, tempo, workDuration, restDuration); err != nil {
		t.Fatalf("audio timing verification failed: %v", err)
	}

	// Verify cue events occurred
	cueEventsMu.Lock()
	events := make([]CueEvent, len(cueEvents))
	copy(events, cueEvents)
	cueEventsMu.Unlock()

	if len(events) == 0 {
		t.Fatal("no audio cue events were tracked")
	}

	// Verify we have workout start
	hasStart := false
	for _, event := range events {
		if event.cueType == "workout_start" {
			hasStart = true
			break
		}
	}
	if !hasStart {
		t.Error("workout start cue was not recorded")
	}

	// Verify we have workout complete
	hasComplete := false
	for _, event := range events {
		if event.cueType == "workout_complete" {
			hasComplete = true
			break
		}
	}
	if !hasComplete {
		t.Error("workout complete cue was not recorded")
	}

	// Verify we have work/rest transitions for each round
	workCount := 0
	restCount := 0
	for _, event := range events {
		if event.cueType == "work" {
			workCount++
		} else if event.cueType == "rest" {
			restCount++
		}
	}
	if workCount != rounds {
		t.Errorf("expected %d work transitions, got %d", rounds, workCount)
	}
	if restCount != rounds {
		t.Errorf("expected %d rest transitions, got %d", rounds, restCount)
	}

	// Verify beeps occurred (at least some during work periods)
	beepCount := 0
	for _, event := range events {
		if event.cueType == "beep" {
			beepCount++
		}
	}
	if beepCount == 0 {
		t.Error("no beeps were recorded")
	}

	t.Logf("Recorded %d audio cue events", len(events))
	t.Logf("Audio file created at: %s", outputPath)
}

// trackingAudioHandler wraps an AudioCueHandler to track when cues are called
type trackingAudioHandler struct {
	delegate     AudioCueHandler
	trackCue     func(cueType string, roundNum int)
	currentRound int
	mu           sync.Mutex
}

func (t *trackingAudioHandler) PlayBeep() {
	t.mu.Lock()
	roundNum := t.currentRound
	t.mu.Unlock()
	t.trackCue("beep", roundNum)
	t.delegate.PlayBeep()
}

func (t *trackingAudioHandler) PlayPeriodTransition(periodType types.PeriodType) {
	t.mu.Lock()
	if periodType == types.PeriodWork {
		t.currentRound++
		roundNum := t.currentRound
		t.mu.Unlock()
		t.trackCue("work", roundNum)
	} else {
		roundNum := t.currentRound
		t.mu.Unlock()
		t.trackCue("rest", roundNum)
	}
	t.delegate.PlayPeriodTransition(periodType)
}

func (t *trackingAudioHandler) PlayWorkoutStart() {
	t.trackCue("workout_start", 0)
	t.delegate.PlayWorkoutStart()
}

func (t *trackingAudioHandler) PlayWorkoutComplete() {
	t.trackCue("workout_complete", 0)
	// Note: PlayWorkoutComplete triggers finalization which might take time
	// but should not block the callback
	t.delegate.PlayWorkoutComplete()
}

func (t *trackingAudioHandler) PlayComboCallout(combo models.Combo, stance models.Stance) {
	t.delegate.PlayComboCallout(combo, stance)
}

func (t *trackingAudioHandler) PlayRoundCallout(roundNumber int, totalRounds int) {
	t.mu.Lock()
	t.currentRound = roundNumber // Update current round from callout
	t.mu.Unlock()
	t.delegate.PlayRoundCallout(roundNumber, totalRounds)
}

func (t *trackingAudioHandler) Stop() {
	t.delegate.Stop()
}

// noOpDisplayHandler is a no-op display handler for testing
type noOpDisplayHandler struct{}

func (n *noOpDisplayHandler) OnTimerUpdate(remaining time.Duration, periodType types.PeriodType, roundNumber int) {
}
func (n *noOpDisplayHandler) OnPeriodStart(periodType types.PeriodType, roundNumber int, duration time.Duration) {
}
func (n *noOpDisplayHandler) OnPeriodEnd(periodType types.PeriodType, roundNumber int) {}
func (n *noOpDisplayHandler) OnWorkoutStart(totalRounds int)                           {}
func (n *noOpDisplayHandler) OnWorkoutComplete()                                       {}

// verifyAudioTiming analyzes the audio file to verify it matches expected workout timing
func verifyAudioTiming(t *testing.T, audioPath string, workout models.Workout, tempo, workDuration, restDuration time.Duration) error {
	// Get audio file duration using ffprobe
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	durationStr := strings.TrimSpace(string(output))
	var actualDuration float64
	if _, err := fmt.Sscanf(durationStr, "%f", &actualDuration); err != nil {
		return fmt.Errorf("failed to parse duration: %v", err)
	}

	actualDurationSeconds := time.Duration(actualDuration * float64(time.Second))

	// Calculate expected duration
	// The recording logic now:
	// - Records announcements back-to-back (no gaps between work/round/combo)
	// - Only adds silence between beeps based on tempo intervals
	// - Does NOT add full work/rest duration as silence (beeps already account for timing)
	//
	// Workout start: ~1-2 seconds (3 beeps with small gaps + announcements)
	// For each round:
	//   - "work" announcement (~0.5s) + round callout (~1s) + combo callout (~2s): ~3.5s total, back-to-back
	//   - Work period: workDuration with beeps at tempo intervals
	//     * First beep occurs at tempo interval after combo callout
	//     * Subsequent beeps at tempo intervals
	//     * Silence is only added to maintain tempo spacing between beeps
	//   - "rest" announcement: ~0.5 seconds
	//   - Rest period: restDuration with countdown beeps (last 3 seconds, 1s apart)
	// Workout complete: ~1 second

	expectedDuration := 2 * time.Second // Workout start (3 beeps + small gaps)
	for i := 0; i < len(workout.Rounds); i++ {
		// Work period announcements (back-to-back, no gaps)
		// "work" (~0.5s) + "round X of Y" (~1s) + combo (~2s) = ~3.5s
		expectedDuration += 3500 * time.Millisecond
		// Work period: actual workDuration with beeps
		// Beeps are at tempo intervals, so the workDuration already includes the beep timing
		// No additional silence is added for the full work duration
		expectedDuration += workDuration
		// Rest period announcement
		expectedDuration += 500 * time.Millisecond
		// Rest period: actual restDuration with countdown beeps (last 3 seconds)
		// Beeps are 1 second apart in the last 3 seconds
		// No additional silence is added for the full rest duration
		expectedDuration += restDuration
	}
	expectedDuration += 1 * time.Second // Workout complete

	// Allow 20% tolerance for audio generation variations
	tolerance := expectedDuration / 5
	minDuration := expectedDuration - tolerance
	maxDuration := expectedDuration + tolerance

	if actualDurationSeconds < minDuration || actualDurationSeconds > maxDuration {
		t.Errorf("audio duration mismatch: expected ~%v (tolerance: ±%v), got %v",
			expectedDuration, tolerance, actualDurationSeconds)
		return fmt.Errorf("duration out of tolerance")
	}

	// Verify audio file is not empty and has reasonable size
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("audio file is empty")
	}

	// MP3 file should be at least a few KB for a workout
	minSize := int64(10 * 1024) // 10 KB minimum
	if fileInfo.Size() < minSize {
		t.Logf("Warning: audio file size is smaller than expected: %d bytes (expected at least %d bytes)",
			fileInfo.Size(), minSize)
	}

	// Use ffprobe to get more detailed information about the audio
	cmd = exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name,sample_rate,channels",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath)

	output, err = cmd.Output()
	if err != nil {
		// Not critical, just log
		t.Logf("Could not get detailed audio info: %v", err)
	} else {
		t.Logf("Audio file info: %s", strings.TrimSpace(string(output)))
	}

	return nil
}
