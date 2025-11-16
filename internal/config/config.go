package config

import (
	"encoding/json"
	"fmt"
	"heavybagworkout/internal/models"
	"os"
	"time"
)

// AppConfig represents the application configuration
type AppConfig struct {
	Workout      WorkoutConfig   `json:"workout"`
	Pattern      PatternConfig   `json:"pattern"`
	Generator    GeneratorConfig `json:"generator"`
	Stance       string          `json:"stance,omitempty"`         // "orthodox" or "southpaw", defaults to "orthodox"
	OpenAIAPIKey string          `json:"openai_api_key,omitempty"` // Optional, can be set via env var
}

// WorkoutConfig represents workout timing configuration
type WorkoutConfig struct {
	WorkDurationSeconds int `json:"work_duration_seconds"`
	RestDurationSeconds int `json:"rest_duration_seconds"`
	TotalRounds         int `json:"total_rounds"`
}

// PatternConfig represents combo pattern configuration
type PatternConfig struct {
	Type             string `json:"type"` // "linear", "pyramid", "random", "constant"
	MinMoves         int    `json:"min_moves"`
	MaxMoves         int    `json:"max_moves"`
	IncludeDefensive bool   `json:"include_defensive"`
}

// GeneratorConfig represents combo generation method
type GeneratorConfig struct {
	UseLLM   bool   `json:"use_llm"`             // true = LLM, false = in-house
	LLMModel string `json:"llm_model,omitempty"` // Optional, defaults to gpt-4o-mini
}

// Validate validates the configuration
func (c *AppConfig) Validate() error {
	if err := c.Workout.Validate(); err != nil {
		return fmt.Errorf("workout config: %w", err)
	}
	if err := c.Pattern.Validate(); err != nil {
		return fmt.Errorf("pattern config: %w", err)
	}
	if err := c.Generator.Validate(); err != nil {
		return fmt.Errorf("generator config: %w", err)
	}
	// Set stance to "orthodox" by default if not specified
	if c.Stance == "" {
		c.Stance = "orthodox"
	}

	return nil
}

// Validate validates workout configuration
func (wc *WorkoutConfig) Validate() error {
	if wc.WorkDurationSeconds <= 0 {
		return fmt.Errorf("work_duration_seconds must be greater than 0, got %d", wc.WorkDurationSeconds)
	}
	if wc.RestDurationSeconds < 0 {
		return fmt.Errorf("rest_duration_seconds must be non-negative, got %d", wc.RestDurationSeconds)
	}
	if wc.TotalRounds <= 0 {
		return fmt.Errorf("total_rounds must be greater than 0, got %d", wc.TotalRounds)
	}
	return nil
}

// Validate validates pattern configuration
func (pc *PatternConfig) Validate() error {
	validTypes := map[string]bool{
		"linear":   true,
		"pyramid":  true,
		"random":   true,
		"constant": true,
	}
	if !validTypes[pc.Type] {
		return fmt.Errorf("pattern type must be one of: linear, pyramid, random, constant, got %s", pc.Type)
	}
	if pc.MinMoves <= 0 {
		return fmt.Errorf("min_moves must be greater than 0, got %d", pc.MinMoves)
	}
	if pc.MaxMoves < pc.MinMoves {
		return fmt.Errorf("max_moves (%d) must be >= min_moves (%d)", pc.MaxMoves, pc.MinMoves)
	}
	return nil
}

// Validate validates generator configuration
func (gc *GeneratorConfig) Validate() error {
	if gc.UseLLM && gc.LLMModel == "" {
		// Allow empty model, will use default in LLM generator
	}
	return nil
}

// ToModelsWorkoutConfig converts config to models.WorkoutConfig
func (wc *WorkoutConfig) ToModelsWorkoutConfig() models.WorkoutConfig {
	return models.NewWorkoutConfig(
		time.Duration(wc.WorkDurationSeconds)*time.Second,
		time.Duration(wc.RestDurationSeconds)*time.Second,
		wc.TotalRounds,
	)
}

// ToModelsWorkoutPattern converts config to models.WorkoutPattern
func (pc *PatternConfig) ToModelsWorkoutPattern() models.WorkoutPattern {
	var patternType models.WorkoutPatternType
	switch pc.Type {
	case "linear":
		patternType = models.PatternLinear
	case "pyramid":
		patternType = models.PatternPyramid
	case "random":
		patternType = models.PatternRandom
	case "constant":
		patternType = models.PatternConstant
	default:
		patternType = models.PatternConstant
	}
	return models.NewWorkoutPattern(patternType, pc.MinMoves, pc.MaxMoves, pc.IncludeDefensive)
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(filename string) (*AppConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// SaveToFile saves configuration to a JSON file
func (c *AppConfig) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadDefault returns the default configuration
func LoadDefault() *AppConfig {
	return &AppConfig{
		Workout: WorkoutConfig{
			WorkDurationSeconds: 20,
			RestDurationSeconds: 10,
			TotalRounds:         8,
		},
		Pattern: PatternConfig{
			Type:             "constant",
			MinMoves:         2,
			MaxMoves:         5,
			IncludeDefensive: false,
		},
		Generator: GeneratorConfig{
			UseLLM:   false,
			LLMModel: "gpt-4o-mini",
		},
		Stance: "orthodox", // Default stance
	}
}

// LoadFromPreset loads configuration from a named preset
func LoadFromPreset(presetName string) (*AppConfig, error) {
	var preset models.WorkoutPreset
	switch presetName {
	case "beta_style", "beta-style":
		preset = models.PresetBetaStyle
	case "endurance":
		preset = models.PresetEndurance
	case "power":
		preset = models.PresetPower
	default:
		return nil, fmt.Errorf("unknown preset: %s (valid presets: beta_style, endurance, power)", presetName)
	}

	workoutConfig := models.PresetWorkoutConfig(preset)

	return &AppConfig{
		Workout: WorkoutConfig{
			WorkDurationSeconds: int(workoutConfig.WorkDuration.Seconds()),
			RestDurationSeconds: int(workoutConfig.RestDuration.Seconds()),
			TotalRounds:         workoutConfig.TotalRounds,
		},
		Pattern: PatternConfig{
			Type:             "constant",
			MinMoves:         2,
			MaxMoves:         5,
			IncludeDefensive: false,
		},
		Generator: GeneratorConfig{
			UseLLM:   false,
			LLMModel: "gpt-4o-mini",
		},
		Stance: "orthodox", // Default stance for presets
	}, nil
}

// GetOpenAIAPIKey returns the OpenAI API key from config or environment variable
func (c *AppConfig) GetOpenAIAPIKey() string {
	if c.OpenAIAPIKey != "" {
		return c.OpenAIAPIKey
	}
	// Fall back to environment variable
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return apiKey
	}
	return ""
}

// GetStance returns the stance from config, defaulting to orthodox if not set
func (c *AppConfig) GetStance() string {
	if c.Stance == "" {
		return "orthodox"
	}
	return c.Stance
}
