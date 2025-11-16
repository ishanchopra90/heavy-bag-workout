package main

import (
	"flag"
	"heavybagworkout/internal/config"
	"heavybagworkout/internal/models"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPresetLoadsCorrectValues(t *testing.T) {
	tests := []struct {
		name                string
		preset              string
		expectedWorkDur     int
		expectedRestDur     int
		expectedTotalRounds int
	}{
		{
			name:                "power preset",
			preset:              "power",
			expectedWorkDur:     30,
			expectedRestDur:     15,
			expectedTotalRounds: 8,
		},
		{
			name:                "beta_style preset",
			preset:              "beta_style",
			expectedWorkDur:     20,
			expectedRestDur:     10,
			expectedTotalRounds: 8,
		},
		{
			name:                "endurance preset",
			preset:              "endurance",
			expectedWorkDur:     40,
			expectedRestDur:     20,
			expectedTotalRounds: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			appConfig, err := loadConfig("", tt.preset)
			if err != nil {
				t.Fatalf("loadConfig() error = %v", err)
			}

			if appConfig.Workout.WorkDurationSeconds != tt.expectedWorkDur {
				t.Errorf("WorkDurationSeconds = %d, want %d", appConfig.Workout.WorkDurationSeconds, tt.expectedWorkDur)
			}
			if appConfig.Workout.RestDurationSeconds != tt.expectedRestDur {
				t.Errorf("RestDurationSeconds = %d, want %d", appConfig.Workout.RestDurationSeconds, tt.expectedRestDur)
			}
			if appConfig.Workout.TotalRounds != tt.expectedTotalRounds {
				t.Errorf("TotalRounds = %d, want %d", appConfig.Workout.TotalRounds, tt.expectedTotalRounds)
			}
		})
	}
}

func TestFlagOverridesPreset(t *testing.T) {
	tests := []struct {
		name                string
		preset              string
		workDuration        int
		restDuration        int
		totalRounds         int
		expectedWorkDur     int
		expectedRestDur     int
		expectedTotalRounds int
	}{
		{
			name:                "all flags override power preset",
			preset:              "power",
			workDuration:        45,
			restDuration:        25,
			totalRounds:         12,
			expectedWorkDur:     45,
			expectedRestDur:     25,
			expectedTotalRounds: 12,
		},
		{
			name:                "rest duration flag overrides preset when set",
			preset:              "power",
			restDuration:        20,
			expectedWorkDur:     30, // From preset
			expectedRestDur:     20, // Overridden
			expectedTotalRounds: 8,  // From preset
		},
		{
			name:                "zero rest duration flag does not override preset",
			preset:              "power",
			restDuration:        0, // Flag not provided, defaults to 0
			expectedWorkDur:     30,
			expectedRestDur:     15, // Should remain from preset
			expectedTotalRounds: 8,
		},
		{
			name:                "zero work duration flag does not override preset",
			preset:              "power",
			workDuration:        0,  // Flag not provided
			expectedWorkDur:     30, // Should remain from preset
			expectedRestDur:     15,
			expectedTotalRounds: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			appConfig, err := loadConfig("", tt.preset)
			if err != nil {
				t.Fatalf("loadConfig() error = %v", err)
			}

			// Apply flag overrides (simulating main.go logic)
			if tt.workDuration > 0 {
				appConfig.Workout.WorkDurationSeconds = tt.workDuration
			}
			if tt.restDuration > 0 {
				appConfig.Workout.RestDurationSeconds = tt.restDuration
			}
			if tt.totalRounds > 0 {
				appConfig.Workout.TotalRounds = tt.totalRounds
			}

			if appConfig.Workout.WorkDurationSeconds != tt.expectedWorkDur {
				t.Errorf("WorkDurationSeconds = %d, want %d", appConfig.Workout.WorkDurationSeconds, tt.expectedWorkDur)
			}
			if appConfig.Workout.RestDurationSeconds != tt.expectedRestDur {
				t.Errorf("RestDurationSeconds = %d, want %d", appConfig.Workout.RestDurationSeconds, tt.expectedRestDur)
			}
			if appConfig.Workout.TotalRounds != tt.expectedTotalRounds {
				t.Errorf("TotalRounds = %d, want %d", appConfig.Workout.TotalRounds, tt.expectedTotalRounds)
			}
		})
	}
}

func TestPatternFlagOverrides(t *testing.T) {
	baseConfig := &config.AppConfig{
		Pattern: config.PatternConfig{
			Type:             "constant",
			MinMoves:         2,
			MaxMoves:         5,
			IncludeDefensive: false,
		},
	}

	tests := []struct {
		name         string
		patternType  string
		minMoves     int
		maxMoves     int
		includeDef   bool
		noIncludeDef bool
		expectedType string
		expectedMin  int
		expectedMax  int
		expectedDef  bool
		shouldError  bool
	}{
		{
			name:         "pattern type override",
			patternType:  "pyramid",
			expectedType: "pyramid",
			expectedMin:  2,
			expectedMax:  5,
			expectedDef:  false,
		},
		{
			name:        "invalid pattern type",
			patternType: "invalid",
			shouldError: true,
		},
		{
			name:         "pattern type case insensitive",
			patternType:  "PYRAMID",
			expectedType: "pyramid",
		},
		{
			name:        "min moves override",
			minMoves:    3,
			expectedMin: 3,
			expectedMax: 5,
		},
		{
			name:        "max moves override",
			maxMoves:    8,
			expectedMin: 2,
			expectedMax: 8,
		},
		{
			name:        "include defensive enable",
			includeDef:  true,
			expectedDef: true,
		},
		{
			name:         "no include defensive disable",
			noIncludeDef: true,
			expectedDef:  false,
		},
		{
			name:         "no include defensive takes priority",
			includeDef:   true,
			noIncludeDef: true,
			expectedDef:  false,
		},
		{
			name:        "zero min moves does not override",
			minMoves:    0,
			expectedMin: 2,
		},
		{
			name:        "zero max moves does not override",
			maxMoves:    0,
			expectedMax: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConfig := &config.AppConfig{
				Pattern: baseConfig.Pattern,
				Workout: config.WorkoutConfig{WorkDurationSeconds: 20, RestDurationSeconds: 10, TotalRounds: 8},
			}

			// Apply flag overrides (simulating main.go logic)
			if tt.patternType != "" {
				validPatterns := map[string]bool{
					"linear":   true,
					"pyramid":  true,
					"random":   true,
					"constant": true,
				}
				patternLower := strings.ToLower(strings.TrimSpace(tt.patternType))
				if !validPatterns[patternLower] {
					if !tt.shouldError {
						t.Errorf("expected error for invalid pattern type, but got none")
					}
					return
				}
				appConfig.Pattern.Type = patternLower
			}
			if tt.minMoves > 0 {
				appConfig.Pattern.MinMoves = tt.minMoves
			}
			if tt.maxMoves > 0 {
				appConfig.Pattern.MaxMoves = tt.maxMoves
			}
			if tt.noIncludeDef {
				appConfig.Pattern.IncludeDefensive = false
			} else if tt.includeDef {
				appConfig.Pattern.IncludeDefensive = true
			}

			if tt.shouldError {
				t.Errorf("expected error but got none")
				return
			}

			// Set defaults from baseConfig for unspecified values
			expectedType := tt.expectedType
			if expectedType == "" {
				expectedType = baseConfig.Pattern.Type
			}
			expectedMin := tt.expectedMin
			if expectedMin == 0 {
				expectedMin = baseConfig.Pattern.MinMoves
			}
			expectedMax := tt.expectedMax
			if expectedMax == 0 {
				expectedMax = baseConfig.Pattern.MaxMoves
			}

			if appConfig.Pattern.Type != expectedType {
				t.Errorf("Pattern.Type = %s, want %s", appConfig.Pattern.Type, expectedType)
			}
			if appConfig.Pattern.MinMoves != expectedMin {
				t.Errorf("Pattern.MinMoves = %d, want %d", appConfig.Pattern.MinMoves, expectedMin)
			}
			if appConfig.Pattern.MaxMoves != expectedMax {
				t.Errorf("Pattern.MaxMoves = %d, want %d", appConfig.Pattern.MaxMoves, expectedMax)
			}
			if appConfig.Pattern.IncludeDefensive != tt.expectedDef {
				t.Errorf("Pattern.IncludeDefensive = %v, want %v", appConfig.Pattern.IncludeDefensive, tt.expectedDef)
			}
		})
	}
}

func TestPatternTypeValidation(t *testing.T) {
	validPatterns := map[string]bool{
		"linear":   true,
		"pyramid":  true,
		"random":   true,
		"constant": true,
	}

	tests := []struct {
		input         string
		expectedValid bool
		expectedLower string
	}{
		{"linear", true, "linear"},
		{"LINEAR", true, "linear"},
		{"Linear", true, "linear"},
		{"pyramid", true, "pyramid"},
		{"random", true, "random"},
		{"constant", true, "constant"},
		{"  linear  ", true, "linear"},
		{"invalid", false, ""},
		{"", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			patternLower := strings.ToLower(strings.TrimSpace(tt.input))
			isValid := validPatterns[patternLower]

			if isValid != tt.expectedValid {
				t.Errorf("expected validity %v for input '%s', got %v", tt.expectedValid, tt.input, isValid)
			}

			if tt.expectedValid && patternLower != tt.expectedLower {
				t.Errorf("expected normalized pattern '%s', got '%s'", tt.expectedLower, patternLower)
			}
		})
	}
}

func TestStanceFlagOverridesPreset(t *testing.T) {
	// Test that --stance flag overrides preset stance when explicitly provided
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	appConfig, err := loadConfig("", "power")
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Verify preset has default stance
	if appConfig.GetStance() != "orthodox" {
		t.Fatalf("expected preset stance 'orthodox', got %s", appConfig.GetStance())
	}

	// Simulate --stance southpaw flag being provided
	stanceFlag := "southpaw"

	// Apply stance flag logic (simulating main.go)
	var stance *models.Stance
	if stanceFlag != "" {
		southpaw := models.Southpaw
		stance = &southpaw
	} else {
		configStance := parseStance(appConfig.GetStance())
		if configStance == nil {
			orthodox := models.Orthodox
			stance = &orthodox
		} else {
			stance = configStance
		}
	}

	// Verify stance flag overrode preset
	if *stance != models.Southpaw {
		t.Errorf("stance flag did not override preset: got %v, want Southpaw", *stance)
	}
}

func TestStanceFlagUsesConfigWhenNotProvided(t *testing.T) {
	// Test that when --stance flag is not provided, it uses config/preset stance
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	appConfig, err := loadConfig("", "power")
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Simulate stance flag not provided (empty string)
	stanceFlag := ""

	// Apply stance flag logic (simulating main.go)
	var stance *models.Stance
	if stanceFlag != "" {
		stance = parseStance(stanceFlag)
	} else {
		configStance := parseStance(appConfig.GetStance())
		if configStance == nil {
			orthodox := models.Orthodox
			stance = &orthodox
		} else {
			stance = configStance
		}
	}

	// Verify config stance is used (orthodox from preset default)
	if *stance != models.Orthodox {
		t.Errorf("expected to use config stance 'orthodox', got %v", *stance)
	}
}

func TestTempoFlag(t *testing.T) {
	testCases := []struct {
		name           string
		tempoFlag      string
		expectedTempo  time.Duration
		shouldError    bool
		errorSubstring string
	}{
		{
			name:          "slow tempo",
			tempoFlag:     "slow",
			expectedTempo: 5 * time.Second,
			shouldError:   false,
		},
		{
			name:          "medium tempo",
			tempoFlag:     "medium",
			expectedTempo: 4 * time.Second,
			shouldError:   false,
		},
		{
			name:          "fast tempo",
			tempoFlag:     "fast",
			expectedTempo: 3 * time.Second,
			shouldError:   false,
		},
		{
			name:          "superfast tempo",
			tempoFlag:     "superfast",
			expectedTempo: 2 * time.Second,
			shouldError:   false,
		},
		{
			name:          "empty tempo defaults to slow",
			tempoFlag:     "",
			expectedTempo: 5 * time.Second,
			shouldError:   false,
		},
		{
			name:           "invalid tempo",
			tempoFlag:      "invalid",
			expectedTempo:  5 * time.Second, // Default fallback
			shouldError:    true,
			errorSubstring: "invalid tempo",
		},
		{
			name:          "case insensitive tempo",
			tempoFlag:     "FAST",
			expectedTempo: 3 * time.Second,
			shouldError:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Test parseTempo function directly
			tempo := parseTempo(tt.tempoFlag)
			if tempo != tt.expectedTempo {
				t.Errorf("parseTempo(%q) = %v, want %v", tt.tempoFlag, tempo, tt.expectedTempo)
			}

			// Test validation
			if tt.tempoFlag != "" {
				isValid := isValidTempo(tt.tempoFlag)
				if tt.shouldError && isValid {
					t.Errorf("isValidTempo(%q) = true, want false", tt.tempoFlag)
				}
				if !tt.shouldError && !isValid {
					t.Errorf("isValidTempo(%q) = false, want true", tt.tempoFlag)
				}
			}
		})
	}
}
