package config

import (
	"os"
	"testing"
	"time"
)

func TestWorkoutConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  WorkoutConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: WorkoutConfig{
				WorkDurationSeconds: 20,
				RestDurationSeconds: 10,
				TotalRounds:         8,
			},
			wantErr: false,
		},
		{
			name: "zero work duration",
			config: WorkoutConfig{
				WorkDurationSeconds: 0,
				RestDurationSeconds: 10,
				TotalRounds:         8,
			},
			wantErr: true,
		},
		{
			name: "negative work duration",
			config: WorkoutConfig{
				WorkDurationSeconds: -5,
				RestDurationSeconds: 10,
				TotalRounds:         8,
			},
			wantErr: true,
		},
		{
			name: "negative rest duration",
			config: WorkoutConfig{
				WorkDurationSeconds: 20,
				RestDurationSeconds: -5,
				TotalRounds:         8,
			},
			wantErr: true,
		},
		{
			name: "zero rounds",
			config: WorkoutConfig{
				WorkDurationSeconds: 20,
				RestDurationSeconds: 10,
				TotalRounds:         0,
			},
			wantErr: true,
		},
		{
			name: "zero rest duration (valid)",
			config: WorkoutConfig{
				WorkDurationSeconds: 20,
				RestDurationSeconds: 0,
				TotalRounds:         8,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPatternConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  PatternConfig
		wantErr bool
	}{
		{
			name: "valid config - linear",
			config: PatternConfig{
				Type:     "linear",
				MinMoves: 2,
				MaxMoves: 5,
			},
			wantErr: false,
		},
		{
			name: "valid config - pyramid",
			config: PatternConfig{
				Type:     "pyramid",
				MinMoves: 2,
				MaxMoves: 5,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			config: PatternConfig{
				Type:     "invalid",
				MinMoves: 2,
				MaxMoves: 5,
			},
			wantErr: true,
		},
		{
			name: "zero min moves",
			config: PatternConfig{
				Type:     "linear",
				MinMoves: 0,
				MaxMoves: 5,
			},
			wantErr: true,
		},
		{
			name: "max moves less than min moves",
			config: PatternConfig{
				Type:     "linear",
				MinMoves: 5,
				MaxMoves: 2,
			},
			wantErr: true,
		},
		{
			name: "equal min and max moves (valid)",
			config: PatternConfig{
				Type:     "constant",
				MinMoves: 3,
				MaxMoves: 3,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PatternConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppConfig_Validate(t *testing.T) {
	baseConfig := AppConfig{
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
			UseLLM: false,
		},
	}

	tests := []struct {
		name      string
		stance    string
		wantError bool
	}{
		{
			name:      "valid orthodox stance",
			stance:    "orthodox",
			wantError: false,
		},
		{
			name:      "valid southpaw stance",
			stance:    "southpaw",
			wantError: false,
		},
		{
			name:      "empty stance (defaults to orthodox)",
			stance:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := baseConfig
			config.Stance = tt.stance

			err := config.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("expected validation error for stance %s, got nil", tt.stance)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error for stance %s: %v", tt.stance, err)
				}
				// After validation, empty stance should be set to orthodox
				if tt.stance == "" && config.Stance != "orthodox" {
					t.Errorf("expected empty stance to default to 'orthodox', got %s", config.Stance)
				}
			}
		})
	}
}

func TestWorkoutConfig_ToModelsWorkoutConfig(t *testing.T) {
	wc := WorkoutConfig{
		WorkDurationSeconds: 30,
		RestDurationSeconds: 15,
		TotalRounds:         10,
	}

	modelsConfig := wc.ToModelsWorkoutConfig()

	if modelsConfig.WorkDuration != 30*time.Second {
		t.Errorf("WorkDuration = %v, want %v", modelsConfig.WorkDuration, 30*time.Second)
	}
	if modelsConfig.RestDuration != 15*time.Second {
		t.Errorf("RestDuration = %v, want %v", modelsConfig.RestDuration, 15*time.Second)
	}
	if modelsConfig.TotalRounds != 10 {
		t.Errorf("TotalRounds = %d, want %d", modelsConfig.TotalRounds, 10)
	}
}

func TestPatternConfig_ToModelsWorkoutPattern(t *testing.T) {
	tests := []struct {
		name   string
		config PatternConfig
	}{
		{
			name: "linear type",
			config: PatternConfig{
				Type:             "linear",
				MinMoves:         2,
				MaxMoves:         5,
				IncludeDefensive: true,
			},
		},
		{
			name: "pyramid type",
			config: PatternConfig{
				Type:             "pyramid",
				MinMoves:         3,
				MaxMoves:         6,
				IncludeDefensive: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := tt.config.ToModelsWorkoutPattern()
			if pattern.MinMoves != tt.config.MinMoves {
				t.Errorf("MinMoves = %d, want %d", pattern.MinMoves, tt.config.MinMoves)
			}
			if pattern.MaxMoves != tt.config.MaxMoves {
				t.Errorf("MaxMoves = %d, want %d", pattern.MaxMoves, tt.config.MaxMoves)
			}
			if pattern.IncludeDefensive != tt.config.IncludeDefensive {
				t.Errorf("IncludeDefensive = %v, want %v", pattern.IncludeDefensive, tt.config.IncludeDefensive)
			}
		})
	}
}

func TestAppConfig_GetStance(t *testing.T) {
	tests := []struct {
		name           string
		configStance   string
		expectedStance string
	}{
		{
			name:           "explicit orthodox",
			configStance:   "orthodox",
			expectedStance: "orthodox",
		},
		{
			name:           "explicit southpaw",
			configStance:   "southpaw",
			expectedStance: "southpaw",
		},
		{
			name:           "empty stance defaults to orthodox",
			configStance:   "",
			expectedStance: "orthodox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AppConfig{Stance: tt.configStance}
			got := config.GetStance()
			if got != tt.expectedStance {
				t.Errorf("GetStance() = %s, want %s", got, tt.expectedStance)
			}
		})
	}
}

func TestLoadDefault(t *testing.T) {
	config := LoadDefault()

	if config.Workout.WorkDurationSeconds != 20 {
		t.Errorf("WorkDurationSeconds = %d, want %d", config.Workout.WorkDurationSeconds, 20)
	}
	if config.Workout.RestDurationSeconds != 10 {
		t.Errorf("RestDurationSeconds = %d, want %d", config.Workout.RestDurationSeconds, 10)
	}
	if config.Workout.TotalRounds != 8 {
		t.Errorf("TotalRounds = %d, want %d", config.Workout.TotalRounds, 8)
	}
	if config.Stance != "orthodox" {
		t.Errorf("expected default stance 'orthodox', got %s", config.Stance)
	}
	if config.GetStance() != "orthodox" {
		t.Errorf("expected GetStance() to return 'orthodox', got %s", config.GetStance())
	}
	if err := config.Validate(); err != nil {
		t.Errorf("Default config validation failed: %v", err)
	}
}

func TestLoadFromPreset(t *testing.T) {
	tests := []struct {
		name          string
		presetName    string
		expectWorkDur int
		expectRestDur int
		expectRounds  int
		wantErr       bool
	}{
		{
			name:          "beta_style preset",
			presetName:    "beta_style",
			expectWorkDur: 20,
			expectRestDur: 10,
			expectRounds:  8,
			wantErr:       false,
		},
		{
			name:          "endurance preset",
			presetName:    "endurance",
			expectWorkDur: 40,
			expectRestDur: 20,
			expectRounds:  10,
			wantErr:       false,
		},
		{
			name:          "power preset",
			presetName:    "power",
			expectWorkDur: 30,
			expectRestDur: 15,
			expectRounds:  8,
			wantErr:       false,
		},
		{
			name:       "invalid preset",
			presetName: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadFromPreset(tt.presetName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromPreset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if config.Workout.WorkDurationSeconds != tt.expectWorkDur {
					t.Errorf("WorkDurationSeconds = %d, want %d", config.Workout.WorkDurationSeconds, tt.expectWorkDur)
				}
				if config.Workout.RestDurationSeconds != tt.expectRestDur {
					t.Errorf("RestDurationSeconds = %d, want %d", config.Workout.RestDurationSeconds, tt.expectRestDur)
				}
				if config.Workout.TotalRounds != tt.expectRounds {
					t.Errorf("TotalRounds = %d, want %d", config.Workout.TotalRounds, tt.expectRounds)
				}
				if config.Stance != "orthodox" {
					t.Errorf("expected preset stance 'orthodox', got %s", config.Stance)
				}
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name           string
		jsonConfig     string
		expectedStance string
		wantError      bool
		checkFunc      func(*testing.T, *AppConfig)
	}{
		{
			name: "complete config with stance",
			jsonConfig: `{
				"workout": {
					"work_duration_seconds": 25,
					"rest_duration_seconds": 12,
					"total_rounds": 6
				},
				"pattern": {
					"type": "random",
					"min_moves": 3,
					"max_moves": 6,
					"include_defensive": true
				},
				"generator": {
					"use_llm": true,
					"llm_model": "gpt-4o-mini"
				},
				"stance": "southpaw"
			}`,
			expectedStance: "southpaw",
			wantError:      false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				if config.Workout.WorkDurationSeconds != 25 {
					t.Errorf("WorkDurationSeconds = %d, want %d", config.Workout.WorkDurationSeconds, 25)
				}
				if config.Pattern.Type != "random" {
					t.Errorf("Pattern.Type = %s, want random", config.Pattern.Type)
				}
				if !config.Generator.UseLLM {
					t.Error("Generator.UseLLM = false, want true")
				}
			},
		},
		{
			name: "config with orthodox stance",
			jsonConfig: `{
				"workout": {"work_duration_seconds": 20, "rest_duration_seconds": 10, "total_rounds": 8},
				"pattern": {"type": "constant", "min_moves": 2, "max_moves": 5, "include_defensive": false},
				"generator": {"use_llm": false},
				"stance": "orthodox"
			}`,
			expectedStance: "orthodox",
			wantError:      false,
		},
		{
			name: "config without stance (defaults to orthodox)",
			jsonConfig: `{
				"workout": {"work_duration_seconds": 20, "rest_duration_seconds": 10, "total_rounds": 8},
				"pattern": {"type": "constant", "min_moves": 2, "max_moves": 5, "include_defensive": false},
				"generator": {"use_llm": false}
			}`,
			expectedStance: "", // Empty string before validation, but GetStance() returns "orthodox"
			wantError:      false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				// After validation, stance should be set to orthodox
				if config.Stance != "orthodox" {
					t.Errorf("expected stance to be set to 'orthodox' after validation, got %s", config.Stance)
				}
				if config.GetStance() != "orthodox" {
					t.Errorf("GetStance() = %s, want orthodox", config.GetStance())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test_config*.json")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.jsonConfig); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			config, err := LoadFromFile(tmpFile.Name())

			if tt.wantError {
				if err == nil {
					t.Errorf("expected validation error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadFromFile() error = %v", err)
			}

			// Check stance if explicitly set
			if tt.expectedStance != "" && config.Stance != tt.expectedStance {
				t.Errorf("expected stance %s, got %s", tt.expectedStance, config.Stance)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}
}

func TestAppConfig_GetOpenAIAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		config        *AppConfig
		envKey        string
		expectedKey   string
		expectedEmpty bool
	}{
		{
			name: "api key in config",
			config: &AppConfig{
				OpenAIAPIKey: "config-key-123",
			},
			expectedKey:   "config-key-123",
			expectedEmpty: false,
		},
		{
			name: "api key in env var",
			config: &AppConfig{
				OpenAIAPIKey: "",
			},
			envKey:        "env-key-456",
			expectedKey:   "env-key-456",
			expectedEmpty: false,
		},
		{
			name: "no api key",
			config: &AppConfig{
				OpenAIAPIKey: "",
			},
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.envKey)
				defer os.Unsetenv("OPENAI_API_KEY")
			} else {
				os.Unsetenv("OPENAI_API_KEY")
			}

			key := tt.config.GetOpenAIAPIKey()
			if tt.expectedEmpty {
				if key != "" {
					t.Errorf("GetOpenAIAPIKey() = %s, want empty string", key)
				}
			} else {
				if key != tt.expectedKey {
					t.Errorf("GetOpenAIAPIKey() = %s, want %s", key, tt.expectedKey)
				}
			}
		})
	}
}

func TestAppConfig_SaveToFile(t *testing.T) {
	config := &AppConfig{
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
			UseLLM: false,
		},
		Stance: "southpaw",
	}

	tmpFile, err := os.CreateTemp("", "test_save_config*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := config.SaveToFile(tmpFile.Name()); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Load it back and verify all fields including stance
	loadedConfig, err := LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadFromFile() after SaveToFile() error = %v", err)
	}

	if loadedConfig.Workout.WorkDurationSeconds != config.Workout.WorkDurationSeconds {
		t.Errorf("loaded WorkDurationSeconds = %d, want %d", loadedConfig.Workout.WorkDurationSeconds, config.Workout.WorkDurationSeconds)
	}
	if loadedConfig.Workout.RestDurationSeconds != config.Workout.RestDurationSeconds {
		t.Errorf("loaded RestDurationSeconds = %d, want %d", loadedConfig.Workout.RestDurationSeconds, config.Workout.RestDurationSeconds)
	}
	if loadedConfig.Stance != "southpaw" {
		t.Errorf("loaded Stance = %s, want southpaw", loadedConfig.Stance)
	}
}
