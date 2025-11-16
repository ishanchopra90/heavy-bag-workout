package timer

import (
	"fmt"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/types"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DefaultAudioCueHandler provides basic audio cues using system beep
type DefaultAudioCueHandler struct {
	enabled bool
}

// NewDefaultAudioCueHandler creates a new default audio cue handler
func NewDefaultAudioCueHandler(enabled bool) *DefaultAudioCueHandler {
	return &DefaultAudioCueHandler{
		enabled: enabled,
	}
}

// PlayBeep plays a simple beep sound
func (a *DefaultAudioCueHandler) PlayBeep() {
	if !a.enabled {
		return
	}

	switch runtime.GOOS {
	case "darwin": // macOS
		exec.Command("say", "-v", "Bells", "beep").Run()
	case "linux":
		// Try to use beep command or system bell
		exec.Command("beep").Run()
		fmt.Print("\a") // ASCII bell
	case "windows":
		// Windows beep
		fmt.Print("\a")
		exec.Command("powershell", "-c", "[console]::beep(800,200)").Run()
	default:
		fmt.Print("\a") // Fallback to ASCII bell
	}
}

// PlayPeriodTransition plays a sound for period transitions
func (a *DefaultAudioCueHandler) PlayPeriodTransition(periodType types.PeriodType) {
	if !a.enabled {
		return
	}

	switch periodType {
	case types.PeriodWork:
		// Say "work" when transitioning to work period
		exec.Command("say", "-v", "Alex", "work").Run()
	case types.PeriodRest:
		// Say "rest" when transitioning to rest period
		exec.Command("say", "-v", "Alex", "rest").Run()
	}
}

// PlayWorkoutStart plays a sound when workout starts
func (a *DefaultAudioCueHandler) PlayWorkoutStart() {
	if !a.enabled {
		return
	}

	// Play multiple beeps to indicate start
	for i := 0; i < 3; i++ {
		a.PlayBeep()
		if i < 2 {
			// Small delay between beeps
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// PlayWorkoutComplete plays a sound when workout completes
func (a *DefaultAudioCueHandler) PlayWorkoutComplete() {
	if !a.enabled {
		return
	}

	// Say "workout complete" when workout finishes
	exec.Command("say", "-v", "Alex", "workout complete").Run()
}

// PlayComboCallout speaks the combo moves using text-to-speech
func (a *DefaultAudioCueHandler) PlayComboCallout(combo models.Combo, stance models.Stance) {
	if !a.enabled {
		return
	}

	if len(combo.Moves) == 0 {
		return
	}

	// Convert combo to speech-friendly string
	comboText := comboToSpeechString(combo, stance)

	// Use system text-to-speech
	switch runtime.GOOS {
	case "darwin": // macOS
		// Use 'say' command with a clear voice
		exec.Command("say", "-v", "Alex", comboText).Run()
	case "linux":
		// Try espeak or festival
		exec.Command("espeak", comboText).Run()
		// Fallback to festival if espeak not available
		// exec.Command("festival", "--tts").Run() // Would need stdin
	case "windows":
		// Use PowerShell text-to-speech
		exec.Command("powershell", "-c", fmt.Sprintf("Add-Type -AssemblyName System.Speech; $synth = New-Object System.Speech.Synthesis.SpeechSynthesizer; $synth.Speak('%s')", comboText)).Run()
	default:
		// Fallback: just beep
		a.PlayBeep()
	}
}

// comboToSpeechString converts a combo to a natural speech string
func comboToSpeechString(combo models.Combo, stance models.Stance) string {
	if len(combo.Moves) == 0 {
		return ""
	}

	var parts []string
	for _, move := range combo.Moves {
		if move.IsPunch() && move.Punch != nil {
			// Use stance-aware punch name
			punchName := move.Punch.NameForStance(stance)
			parts = append(parts, punchName)
		} else if move.IsDefensive() && move.Defensive != nil {
			// Use defensive move name
			parts = append(parts, move.Defensive.String())
		}
	}

	// Join with "then" for natural speech flow
	// Example: "jab, cross, then left hook"
	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	// For multiple moves, join with commas and "then" before the last one
	if len(parts) == 2 {
		return parts[0] + ", then " + parts[1]
	}

	// For 3+ moves: "jab, cross, left hook, then right uppercut"
	allButLast := strings.Join(parts[:len(parts)-1], ", ")
	return allButLast + ", then " + parts[len(parts)-1]
}

// NoOpAudioCueHandler is a no-op implementation for when audio is disabled
type NoOpAudioCueHandler struct{}

// NewNoOpAudioCueHandler creates a no-op audio handler
func NewNoOpAudioCueHandler() *NoOpAudioCueHandler {
	return &NoOpAudioCueHandler{}
}

func (a *NoOpAudioCueHandler) PlayBeep()                                    {}
func (a *NoOpAudioCueHandler) PlayPeriodTransition(types.PeriodType)        {}
func (a *NoOpAudioCueHandler) PlayWorkoutStart()                            {}
func (a *NoOpAudioCueHandler) PlayWorkoutComplete()                         {}
func (a *NoOpAudioCueHandler) PlayComboCallout(models.Combo, models.Stance) {}

// FileAudioCueHandler plays audio from files (for future implementation)
type FileAudioCueHandler struct {
	enabled      bool
	beepFile     string
	workFile     string
	restFile     string
	startFile    string
	completeFile string
}

// NewFileAudioCueHandler creates a file-based audio handler
func NewFileAudioCueHandler(enabled bool) *FileAudioCueHandler {
	return &FileAudioCueHandler{
		enabled: enabled,
	}
}

// SetBeepFile sets the beep sound file
func (f *FileAudioCueHandler) SetBeepFile(path string) {
	f.beepFile = path
}

// SetWorkFile sets the work period sound file
func (f *FileAudioCueHandler) SetWorkFile(path string) {
	f.workFile = path
}

// SetRestFile sets the rest period sound file
func (f *FileAudioCueHandler) SetRestFile(path string) {
	f.restFile = path
}

// SetStartFile sets the workout start sound file
func (f *FileAudioCueHandler) SetStartFile(path string) {
	f.startFile = path
}

// SetCompleteFile sets the workout complete sound file
func (f *FileAudioCueHandler) SetCompleteFile(path string) {
	f.completeFile = path
}

func (f *FileAudioCueHandler) PlayBeep() {
	if !f.enabled || f.beepFile == "" {
		return
	}
	f.playFile(f.beepFile)
}

func (f *FileAudioCueHandler) PlayPeriodTransition(periodType types.PeriodType) {
	if !f.enabled {
		return
	}

	switch periodType {
	case types.PeriodWork:
		if f.workFile != "" {
			f.playFile(f.workFile)
		} else {
			f.PlayBeep()
		}
	case types.PeriodRest:
		if f.restFile != "" {
			f.playFile(f.restFile)
		} else {
			f.PlayBeep()
		}
	}
}

func (f *FileAudioCueHandler) PlayWorkoutStart() {
	if !f.enabled {
		return
	}
	if f.startFile != "" {
		f.playFile(f.startFile)
	} else {
		f.PlayBeep()
	}
}

func (f *FileAudioCueHandler) PlayWorkoutComplete() {
	if !f.enabled {
		return
	}
	if f.completeFile != "" {
		f.playFile(f.completeFile)
	} else {
		f.PlayBeep()
	}
}

func (f *FileAudioCueHandler) PlayComboCallout(combo models.Combo, stance models.Stance) {
	if !f.enabled {
		return
	}
	// For file-based handler, fall back to default text-to-speech
	// In the future, could use pre-recorded audio files
	defaultHandler := NewDefaultAudioCueHandler(true)
	defaultHandler.PlayComboCallout(combo, stance)
}

func (f *FileAudioCueHandler) playFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	switch runtime.GOOS {
	case "darwin":
		exec.Command("afplay", path).Start()
	case "linux":
		exec.Command("aplay", path).Start()
	case "windows":
		exec.Command("powershell", "-c", fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", path)).Start()
	}
}
