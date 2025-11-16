package main

import (
	"flag"
	"fmt"
	"heavybagworkout/internal/cli"
	"heavybagworkout/internal/config"
	"heavybagworkout/internal/generator"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/timer"
	"os"
	"strings"
	"time"
)

func main() {
	// Command-line flags
	var (
		configFile         = flag.String("config", "", "Path to JSON configuration file")
		preset             = flag.String("preset", "", "Use a preset configuration (beta_style, endurance, power)")
		workDuration       = flag.Int("work-duration", 0, "Work period duration in seconds (overrides config)")
		restDuration       = flag.Int("rest-duration", 0, "Rest period duration in seconds (overrides config)")
		totalRounds        = flag.Int("rounds", 0, "Total number of rounds (overrides config)")
		patternType        = flag.String("pattern", "", "Workout pattern type: linear, pyramid, random, or constant (overrides config)")
		minMoves           = flag.Int("min-moves", 0, "Minimum moves per combo (overrides config)")
		maxMoves           = flag.Int("max-moves", 0, "Maximum moves per combo (overrides config)")
		includeDefensive   = flag.Bool("include-defensive", false, "Include defensive moves in combos (use with --no-include-defensive to disable)")
		noIncludeDefensive = flag.Bool("no-include-defensive", false, "Disable defensive moves in combos (overrides config)")
		useLLM             = flag.Bool("use-llm", false, "Use LLM for combo generation (overrides config)")
		openAIAPIKey       = flag.String("openai-api-key", "", "OpenAI API key (overrides config and env var)")
		stanceFlag         = flag.String("stance", "", "Boxer's stance: orthodox or southpaw (overrides config/preset)")
		tempoFlag          = flag.String("tempo", "", "Workout tempo: slow (5s), medium (4s), fast (3s), or superfast (2s) (default: slow)")
		showVersion        = flag.Bool("version", false, "Show version information")
		showHelp           = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println("Version: 0.1.0")
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	appConfig, err := loadConfig(*configFile, *preset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Apply command-line flag overrides
	// Note: We check > 0 for workDuration and totalRounds to distinguish "not set" (0) from "explicitly set"
	// For restDuration, we also check > 0 because 0 is a valid value, but flag defaults to 0 when not provided
	// To set restDuration to 0 explicitly, users should use a config file
	if *workDuration > 0 {
		appConfig.Workout.WorkDurationSeconds = *workDuration
	}
	if *restDuration > 0 {
		appConfig.Workout.RestDurationSeconds = *restDuration
	}
	if *totalRounds > 0 {
		appConfig.Workout.TotalRounds = *totalRounds
	}
	if *patternType != "" {
		validPatterns := map[string]bool{
			"linear":   true,
			"pyramid":  true,
			"random":   true,
			"constant": true,
		}
		patternLower := strings.ToLower(strings.TrimSpace(*patternType))
		if !validPatterns[patternLower] {
			fmt.Fprintf(os.Stderr, "Error: invalid pattern type '%s'. Must be one of: linear, pyramid, random, constant\n", *patternType)
			os.Exit(1)
		}
		appConfig.Pattern.Type = patternLower
	}
	if *minMoves > 0 {
		appConfig.Pattern.MinMoves = *minMoves
	}
	if *maxMoves > 0 {
		appConfig.Pattern.MaxMoves = *maxMoves
	}
	if *noIncludeDefensive {
		appConfig.Pattern.IncludeDefensive = false
	} else if *includeDefensive {
		appConfig.Pattern.IncludeDefensive = true
	}
	if *useLLM {
		appConfig.Generator.UseLLM = true
	}
	if *openAIAPIKey != "" {
		appConfig.OpenAIAPIKey = *openAIAPIKey
	}

	// Validate configuration after applying overrides
	if err := appConfig.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Generate workout
	fmt.Println("Puppy Power - Heavy Bag Workout App")
	fmt.Println("=====================================")
	fmt.Println("\nGenerating workout...")

	workoutConfig := appConfig.Workout.ToModelsWorkoutConfig()
	workoutPattern := appConfig.Pattern.ToModelsWorkoutPattern()

	// Parse stance flag (overrides config if provided)
	var stance *models.Stance
	if *stanceFlag != "" {
		stance = parseStance(*stanceFlag)
		if stance == nil {
			fmt.Fprintf(os.Stderr, "Error: invalid stance '%s'. Must be 'orthodox' or 'southpaw'\n", *stanceFlag)
			os.Exit(1)
		}
	} else {
		// Use stance from config, defaulting to orthodox
		configStance := parseStance(appConfig.GetStance())
		if configStance == nil {
			// This shouldn't happen if validation passed, but handle it
			orthodox := models.Orthodox
			stance = &orthodox
		} else {
			stance = configStance
		}
	}

	// Parse and validate tempo flag
	if *tempoFlag != "" {
		if !isValidTempo(*tempoFlag) {
			fmt.Fprintf(os.Stderr, "Error: invalid tempo '%s'. Must be one of: slow, medium, fast, superfast\n", *tempoFlag)
			os.Exit(1)
		}
	}

	var workout models.Workout

	if appConfig.Generator.UseLLM {
		apiKey := appConfig.GetOpenAIAPIKey()
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "\nError: OpenAI API key required for LLM generation.\n")
			fmt.Fprintf(os.Stderr, "Set OPENAI_API_KEY environment variable or use --openai-api-key flag.\n")
			os.Exit(1)
		}

		fmt.Println("  Using LLM generator...")
		llmGenerator := generator.NewLLMWorkoutGenerator(apiKey)
		var err error
		workout, err = llmGenerator.GenerateWorkoutWithStance(workoutConfig, workoutPattern, *stance)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError generating workout with LLM: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("  Using in-house generator...")
		inHouseGenerator := generator.NewWorkoutGenerator()
		var err error
		workout, err = inHouseGenerator.GenerateWorkout(workoutConfig, workoutPattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError generating workout: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("  Workout generated successfully!")
	fmt.Printf("  Total rounds: %d\n", len(workout.Rounds))
	fmt.Println()

	// Parse tempo flag
	tempo := parseTempo(*tempoFlag)

	// Create audio handler (enabled by default)
	audioHandler := timer.NewDefaultAudioCueHandler(true)

	// Create and run CLI interface with stance and tempo
	workoutInterface := cli.NewWorkoutInterfaceWithStanceAndTempo(workout, audioHandler, *stance, tempo)
	if err := workoutInterface.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running workout: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(configFile, preset string) (*config.AppConfig, error) {
	// Priority: config file > preset > default
	if configFile != "" {
		return config.LoadFromFile(configFile)
	}
	if preset != "" {
		return config.LoadFromPreset(preset)
	}
	return config.LoadDefault(), nil
}

func printHelp() {
	fmt.Println("Puppy Power - Heavy Bag Workout App")
	fmt.Println("=====================================")
	fmt.Println("\nUsage:")
	fmt.Println("  heavybagworkout [flags]")
	fmt.Println("Flags:")
	fmt.Println("  --config string           Path to JSON configuration file")
	fmt.Println("  --preset string           Use a preset configuration (beta_style, endurance, power)")
	fmt.Println("  --work-duration int       Work period duration in seconds (overrides config)")
	fmt.Println("  --rest-duration int       Rest period duration in seconds (overrides config)")
	fmt.Println("  --rounds int              Total number of rounds (overrides config)")
	fmt.Println("  --pattern string          Workout pattern type: linear, pyramid, random, or constant (overrides config)")
	fmt.Println("  --min-moves int           Minimum moves per combo (overrides config)")
	fmt.Println("  --max-moves int           Maximum moves per combo (overrides config)")
	fmt.Println("  --include-defensive       Include defensive moves in combos")
	fmt.Println("  --no-include-defensive    Disable defensive moves in combos")
	fmt.Println("  --use-llm                 Use LLM for combo generation (overrides config)")
	fmt.Println("  --openai-api-key string   OpenAI API key (overrides config and env var)")
	fmt.Println("  --stance string           Boxer's stance: orthodox or southpaw (overrides config/preset, defaults to orthodox if not set)")
	fmt.Println("  --tempo string            Workout tempo: slow (5s), medium (4s), fast (3s), or superfast (2s) (default: slow)")
	fmt.Println("  --version                 Show version information")
	fmt.Println("  --help                    Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  heavybagworkout --preset beta_style")
	fmt.Println("  heavybagworkout --config configs/custom.json")
	fmt.Println("  heavybagworkout --work-duration 30 --rounds 10 --use-llm")
	fmt.Println("  heavybagworkout --preset beta_style --stance southpaw")
	fmt.Println("  heavybagworkout --pattern pyramid --min-moves 2 --max-moves 6 --include-defensive")
	fmt.Println("  heavybagworkout --preset power --tempo fast")
	fmt.Println("\nConfiguration Priority:")
	fmt.Println("  1. Command-line flags (highest priority)")
	fmt.Println("  2. Config file (--config)")
	fmt.Println("  3. Preset (--preset)")
	fmt.Println("  4. Default configuration (lowest priority)")
}

// parseStance parses a stance string and returns the corresponding Stance value
func parseStance(s string) *models.Stance {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "orthodox":
		stance := models.Orthodox
		return &stance
	case "southpaw":
		stance := models.Southpaw
		return &stance
	default:
		return nil
	}
}

// isValidTempo checks if a tempo string is valid
func isValidTempo(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	validTempos := map[string]bool{
		"slow":      true,
		"medium":    true,
		"fast":      true,
		"superfast": true,
	}
	return validTempos[s]
}

// parseTempo parses a tempo string and returns the corresponding duration
// Valid values: "slow" or "" (default, 5s), "medium" (4s), "fast" (3s), "superfast" (2s)
func parseTempo(s string) time.Duration {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "medium":
		return 4 * time.Second
	case "fast":
		return 3 * time.Second
	case "superfast":
		return 2 * time.Second
	case "slow", "":
		// Default to slow (5 seconds)
		return 5 * time.Second
	default:
		// Invalid tempo, default to slow
		return 5 * time.Second
	}
}
