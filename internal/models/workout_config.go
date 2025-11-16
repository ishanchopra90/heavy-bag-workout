package models

import "time"

// WorkoutConfig contains the configuration for a workout
type WorkoutConfig struct {
	WorkDuration time.Duration // Duration of work periods (e.g., 20 seconds)
	RestDuration time.Duration // Duration of rest periods (e.g., 10 seconds)
	TotalRounds  int           // Total number of rounds
}

// WorkoutPreset defines named preset configurations
type WorkoutPreset string

const (
	PresetBetaStyle WorkoutPreset = "beta_style" // 20s work / 10s rest / 8 rounds
	PresetEndurance WorkoutPreset = "endurance"  // 40s work / 20s rest / 10 rounds
	PresetPower     WorkoutPreset = "power"      // 30s work / 15s rest / 8 rounds
)

var presetConfigMap = map[WorkoutPreset]WorkoutConfig{
	PresetBetaStyle: {
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  8,
	},
	PresetEndurance: {
		WorkDuration: 40 * time.Second,
		RestDuration: 20 * time.Second,
		TotalRounds:  10,
	},
	PresetPower: {
		WorkDuration: 30 * time.Second,
		RestDuration: 15 * time.Second,
		TotalRounds:  8,
	},
}

// NewWorkoutConfig creates a new workout configuration with default values
func NewDefaultWorkoutConfig() WorkoutConfig {
	return WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  10,
	}
}

func NewWorkoutConfig(workDuration, restDuration time.Duration, totalRounds int) WorkoutConfig {
	return WorkoutConfig{
		WorkDuration: workDuration,
		RestDuration: restDuration,
		TotalRounds:  totalRounds,
	}
}

// PresetWorkoutConfig returns a preset configuration by name.
// Defaults to Beta Style if the preset is unknown.
func PresetWorkoutConfig(preset WorkoutPreset) WorkoutConfig {
	if config, ok := presetConfigMap[preset]; ok {
		return config
	}
	return presetConfigMap[PresetBetaStyle]
}

// AvailablePresets returns the list of preset names.
func AvailablePresets() []WorkoutPreset {
	return []WorkoutPreset{
		PresetBetaStyle,
		PresetEndurance,
		PresetPower,
	}
}

// Validate checks if the workout configuration is valid
func (wc WorkoutConfig) Validate() error {
	if wc.WorkDuration <= 0 {
		return ErrInvalidWorkDuration
	}
	if wc.RestDuration < 0 {
		return ErrInvalidRestDuration
	}
	if wc.TotalRounds <= 0 {
		return ErrInvalidTotalRounds
	}
	return nil
}
