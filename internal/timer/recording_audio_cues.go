package timer

import (
	"fmt"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// debugEnabled checks if debug logging is enabled via environment variable
func debugEnabled() bool {
	return os.Getenv("HEAVYBAG_DEBUG") == "1" || os.Getenv("DEBUG") == "1"
}

// debugLog prints a debug message if debug logging is enabled
func debugLog(format string, args ...interface{}) {
	if debugEnabled() {
		fmt.Fprintf(os.Stderr, "Debug: "+format, args...)
	}
}

// RecordingAudioCueHandler records system audio output during the workout
// This is a simplified implementation that captures all audio output
// rather than generating individual audio files and concatenating them
type RecordingAudioCueHandler struct {
	baseHandler      *DefaultAudioCueHandler
	outputPath       string
	recordingCmd     *exec.Cmd
	recordingMu      sync.Mutex
	isRecording      bool
	workoutStartTime time.Time
}

// NewRecordingAudioCueHandler creates a new recording audio cue handler
func NewRecordingAudioCueHandler(outputPath string) (*RecordingAudioCueHandler, error) {
	return NewRecordingAudioCueHandlerWithWorkout(outputPath, models.Workout{}, 5*time.Second)
}

// NewRecordingAudioCueHandlerWithWorkout creates a new recording audio cue handler
// The workout and tempo parameters are kept for API compatibility but not used in this implementation
func NewRecordingAudioCueHandlerWithWorkout(outputPath string, workout models.Workout, tempo time.Duration) (*RecordingAudioCueHandler, error) {
	handler := &RecordingAudioCueHandler{
		baseHandler:      NewDefaultAudioCueHandler(true),
		outputPath:       outputPath,
		isRecording:      false,
		workoutStartTime: time.Time{},
	}

	return handler, nil
}

// findBlackHoleDeviceIndex finds the BlackHole audio device index on macOS
func findBlackHoleDeviceIndex() (int, error) {
	// List all available audio devices using ffmpeg
	cmd := exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	var stderr strings.Builder
	cmd.Stderr = &stderr

	// Run the command (it will fail, but we need the stderr output)
	_ = cmd.Run()

	output := stderr.String()

	// Debug: print the output to help diagnose issues
	debugLog("ffmpeg device list output:\n%s\n", output)

	// Look for BlackHole in the audio devices section
	// ffmpeg output format:
	//   AVFoundation audio devices:
	//   [AVFoundation input device @ ...] [<index>] <device_name>
	//   [<index>] <device_name>
	// We need to look in the audio INPUT devices section
	inAudioSection := false
	blackHolePattern := regexp.MustCompile(`\[(\d+)\]\s+.*[Bb]lack[Hh]ole`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Check if we're entering the audio devices section
		if strings.Contains(line, "AVFoundation audio devices:") {
			inAudioSection = true
			continue
		}
		// Exit audio section when we hit video section
		if strings.Contains(line, "AVFoundation video devices:") {
			inAudioSection = false
			continue
		}

		// Only look for BlackHole in the audio section
		if !inAudioSection {
			continue
		}

		matches := blackHolePattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			// Found BlackHole device
			index, err := strconv.Atoi(matches[1])
			if err == nil {
				debugLog("Found BlackHole at device index %d\n", index)
				return index, nil
			}
		}
	}

	// Try alternative pattern in case format is slightly different
	inAudioSection = false
	altPattern := regexp.MustCompile(`(\d+)\.\s+.*[Bb]lack[Hh]ole`)
	for _, line := range lines {
		if strings.Contains(line, "AVFoundation audio devices:") {
			inAudioSection = true
			continue
		}
		if strings.Contains(line, "AVFoundation video devices:") {
			inAudioSection = false
			continue
		}

		if !inAudioSection {
			continue
		}

		matches := altPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			index, err := strconv.Atoi(matches[1])
			if err == nil {
				debugLog("Found BlackHole at device index %d (alt pattern)\n", index)
				return index, nil
			}
		}
	}

	return -1, fmt.Errorf("BlackHole audio device not found. Please ensure:\n1. BlackHole is installed\n2. BlackHole is set as your system output device (or part of a Multi-Output Device)\n3. BlackHole appears in Audio MIDI Setup")
}

// startSystemAudioRecording starts recording system audio output
func (r *RecordingAudioCueHandler) startSystemAudioRecording() error {
	r.recordingMu.Lock()
	defer r.recordingMu.Unlock()

	if r.isRecording {
		return nil // Already recording
	}

	// Determine output format from file extension
	ext := strings.ToLower(filepath.Ext(r.outputPath))
	var format string
	var codec string

	switch ext {
	case ".mp3":
		format = "mp3"
		codec = "libmp3lame"
	case ".m4a", ".aac":
		format = "m4a"
		codec = "aac"
	case ".wav":
		format = "wav"
		codec = "pcm_s16le"
	default:
		// Default to MP3 for Android compatibility
		format = "mp3"
		codec = "libmp3lame"
		r.outputPath = r.outputPath + ".mp3"
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		// Use ffmpeg with avfoundation to capture system audio from BlackHole
		// This requires BlackHole to be installed and set as the output device
		// Format: ffmpeg -f avfoundation -i ":<audio_device_index>" output.mp3

		// Find BlackHole device index
		blackHoleIndex, err := findBlackHoleDeviceIndex()
		if err != nil {
			return fmt.Errorf("failed to find BlackHole device: %w. Please ensure BlackHole is installed and configured in Audio MIDI Setup", err)
		}

		// Build ffmpeg command to capture from BlackHole
		// Format: -i ":<audio_index>" means no video, audio from device at index
		audioInput := fmt.Sprintf(":%d", blackHoleIndex)
		args := []string{
			"-f", "avfoundation",
			"-i", audioInput,
			"-acodec", codec,
			"-ar", "44100", // Sample rate
			"-ac", "2", // Stereo
		}

		if format == "mp3" {
			args = append(args, "-b:a", "192k") // Bitrate for MP3
		}

		args = append(args, "-y", r.outputPath) // Overwrite output file

		cmd = exec.Command("ffmpeg", args...)

		// Capture stderr for debugging (ffmpeg outputs to stderr)
		var stderr strings.Builder
		cmd.Stderr = &stderr

		// Start recording in background
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start audio recording: %w (stderr: %s). Note: On macOS, ensure BlackHole is installed, configured in Audio MIDI Setup, and set as your system output device.", err, stderr.String())
		}

		r.recordingCmd = cmd
		r.isRecording = true

		// Log that recording started with full command for debugging
		debugLog("Started system audio recording from BlackHole (device %d) to %s\n", blackHoleIndex, r.outputPath)
		debugLog("ffmpeg command: ffmpeg %s\n", strings.Join(args, " "))
		debugLog("Make sure BlackHole is set as your system output device in System Settings > Sound\n")

	case "linux":
		// Linux: Use PulseAudio to capture system audio
		// Requires pulseaudio and ffmpeg
		args := []string{
			"-f", "pulse",
			"-i", "default",
			"-acodec", codec,
			"-ar", "44100",
			"-ac", "2",
		}

		if format == "mp3" {
			args = append(args, "-b:a", "192k")
		}

		args = append(args, "-y", r.outputPath)

		cmd = exec.Command("ffmpeg", args...)
		var stderr strings.Builder
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start audio recording: %w (stderr: %s)", err, stderr.String())
		}

		r.recordingCmd = cmd
		r.isRecording = true
		debugLog("Started system audio recording to %s\n", r.outputPath)

	case "windows":
		// Windows: Use dshow (DirectShow) or wasapi (Windows Audio Session API)
		// Try wasapi first (Windows 7+)
		args := []string{
			"-f", "wasapi",
			"-i", "default",
			"-acodec", codec,
			"-ar", "44100",
			"-ac", "2",
		}

		if format == "mp3" {
			args = append(args, "-b:a", "192k")
		}

		args = append(args, "-y", r.outputPath)

		cmd = exec.Command("ffmpeg", args...)
		var stderr strings.Builder
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start audio recording: %w (stderr: %s)", err, stderr.String())
		}

		r.recordingCmd = cmd
		r.isRecording = true
		debugLog("Started system audio recording to %s\n", r.outputPath)

	default:
		return fmt.Errorf("system audio recording not supported on %s", runtime.GOOS)
	}

	return nil
}

// stopSystemAudioRecording stops the system audio recording
func (r *RecordingAudioCueHandler) stopSystemAudioRecording() error {
	r.recordingMu.Lock()
	defer r.recordingMu.Unlock()

	if !r.isRecording || r.recordingCmd == nil {
		return nil // Not recording
	}

	// For ffmpeg, we need to send 'q' to stdin to quit gracefully and finalize the file
	// However, since we started it with Start() (not Run()), we don't have stdin access
	// So we'll use SIGINT which ffmpeg handles gracefully on macOS
	if r.recordingCmd.Process != nil {
		// Send SIGINT to ffmpeg to stop recording gracefully
		if err := r.recordingCmd.Process.Signal(os.Interrupt); err != nil {
			// If interrupt fails, try SIGTERM
			if err := r.recordingCmd.Process.Signal(os.Kill); err != nil {
				// Last resort: force kill
				if err := r.recordingCmd.Process.Kill(); err != nil {
					return fmt.Errorf("failed to stop audio recording: %w", err)
				}
			}
		}
	}

	// Wait for the process to finish (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- r.recordingCmd.Wait()
	}()

	select {
	case err := <-done:
		// ffmpeg might return an error on interrupt, which is normal
		// Check if the output file was created
		if _, statErr := os.Stat(r.outputPath); statErr != nil {
			return fmt.Errorf("audio recording stopped but output file not found: %w (ffmpeg error: %v)", statErr, err)
		}
	case <-time.After(5 * time.Second):
		// Timeout - force kill if still running
		if r.recordingCmd.Process != nil {
			r.recordingCmd.Process.Kill()
		}
		return fmt.Errorf("timeout waiting for ffmpeg to stop")
	}

	r.isRecording = false
	r.recordingCmd = nil

	debugLog("Stopped system audio recording\n")

	return nil
}

// PlayBeep plays a beep (delegates to base handler)
func (r *RecordingAudioCueHandler) PlayBeep() {
	r.baseHandler.PlayBeep()
}

// PlayPeriodTransition plays a period transition (delegates to base handler)
func (r *RecordingAudioCueHandler) PlayPeriodTransition(periodType types.PeriodType) {
	r.baseHandler.PlayPeriodTransition(periodType)
}

// PlayWorkoutStart starts the workout and begins recording
func (r *RecordingAudioCueHandler) PlayWorkoutStart() {
	r.workoutStartTime = time.Now()

	// Start system audio recording BEFORE any audio cues play
	if err := r.startSystemAudioRecording(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to start audio recording: %v\n", err)
		fmt.Fprintf(os.Stderr, "Note: On macOS, you may need to install BlackHole, configure it in Audio MIDI Setup, and set it as your system output device\n")
		// Continue anyway - the workout will still run, just without recording
	} else {
		// Give ffmpeg a small moment to initialize and start capturing
		// This ensures we don't miss the first audio cues
		time.Sleep(100 * time.Millisecond)
	}

	// Now play the workout start audio cues (which will be captured by the recording)
	r.baseHandler.PlayWorkoutStart()
}

// PlayWorkoutComplete stops recording and finalizes the audio file
func (r *RecordingAudioCueHandler) PlayWorkoutComplete() {
	r.baseHandler.PlayWorkoutComplete()

	// Stop system audio recording
	if err := r.stopSystemAudioRecording(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to stop audio recording: %v\n", err)
	} else {
		fmt.Printf("Audio recording saved to: %s\n", r.outputPath)
	}
}

// PlayComboCallout plays combo callout (delegates to base handler)
func (r *RecordingAudioCueHandler) PlayComboCallout(combo models.Combo, stance models.Stance) {
	r.baseHandler.PlayComboCallout(combo, stance)
}

// PlayRoundCallout plays round callout (delegates to base handler)
func (r *RecordingAudioCueHandler) PlayRoundCallout(roundNumber int, totalRounds int) {
	r.baseHandler.PlayRoundCallout(roundNumber, totalRounds)
}

// Stop cancels all running audio commands
func (r *RecordingAudioCueHandler) Stop() {
	r.baseHandler.Stop()
}

// Cleanup stops recording if workout is interrupted
func (r *RecordingAudioCueHandler) Cleanup() {
	r.stopSystemAudioRecording()
}

// WaitForFinalization waits for audio recording to complete
// For system audio capture, this just ensures recording has stopped
func (r *RecordingAudioCueHandler) WaitForFinalization() error {
	// Give a small delay to ensure recording has stopped
	time.Sleep(500 * time.Millisecond)

	r.recordingMu.Lock()
	stillRecording := r.isRecording
	r.recordingMu.Unlock()

	if stillRecording {
		// Force stop if still recording
		return r.stopSystemAudioRecording()
	}

	// Verify output file exists and has content
	fileInfo, err := os.Stat(r.outputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("audio output file was not created: %s", r.outputPath)
	}
	if err != nil {
		return fmt.Errorf("error checking audio output file: %w", err)
	}

	// Check if file has reasonable size (at least 1KB - very small files are likely empty/corrupted)
	if fileInfo.Size() < 1024 {
		return fmt.Errorf("audio output file is too small (%d bytes), likely empty or corrupted. Make sure BlackHole is set as your system output device", fileInfo.Size())
	}

	debugLog("Audio file created successfully: %s (%d bytes)\n", r.outputPath, fileInfo.Size())
	return nil
}
