package generator

import (
	"encoding/json"
	"fmt"
	"heavybagworkout/internal/models"
	"strings"
)

// LLMWorkoutGenerator generates full workouts using OpenAI API
type LLMWorkoutGenerator struct {
	openAIClient *OpenAIClient
	moveMapping  models.MoveMapping
}

// NewLLMWorkoutGenerator creates a new LLM-based workout generator
func NewLLMWorkoutGenerator(apiKey string) *LLMWorkoutGenerator {
	return &LLMWorkoutGenerator{
		openAIClient: NewOpenAIClient(apiKey),
		moveMapping:  models.NewMoveMapping(),
	}
}

// NewLLMWorkoutGeneratorWithOpenAIClient allows injecting a custom OpenAI client (primarily for testing)
func NewLLMWorkoutGeneratorWithOpenAIClient(client *OpenAIClient) *LLMWorkoutGenerator {
	if client == nil {
		client = NewOpenAIClient("")
	}
	return &LLMWorkoutGenerator{
		openAIClient: client,
		moveMapping:  models.NewMoveMapping(),
	}
}

// GenerateWorkout generates an entire workout using a single OpenAI API call
func (lg *LLMWorkoutGenerator) GenerateWorkout(config models.WorkoutConfig, pattern models.WorkoutPattern) (models.Workout, error) {
	return lg.GenerateWorkoutWithStance(config, pattern, models.Orthodox)
}

// GenerateWorkoutWithStance generates an entire workout using a single OpenAI API call with stance information
func (lg *LLMWorkoutGenerator) GenerateWorkoutWithStance(config models.WorkoutConfig, pattern models.WorkoutPattern, stance models.Stance) (models.Workout, error) {
	prompt := lg.buildWorkoutPrompt(config, pattern, stance)

	response, err := lg.openAIClient.GenerateWorkoutRequest(prompt)
	if err != nil {
		return models.Workout{}, fmt.Errorf("failed to generate workout: %w", err)
	}

	// Parse the JSON response
	var workoutResp WorkoutResponseJSON
	// Clean the response in case there's markdown code blocks
	cleanedResponse := strings.TrimSpace(response)
	if strings.HasPrefix(cleanedResponse, "```json") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSpace(cleanedResponse)
	} else if strings.HasPrefix(cleanedResponse, "```") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSpace(cleanedResponse)
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &workoutResp); err != nil {
		return models.Workout{}, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Convert JSON response to Workout model
	rounds := make([]models.WorkoutRound, 0, len(workoutResp.Rounds))
	for _, roundResp := range workoutResp.Rounds {
		if len(roundResp.Combo.Moves) == 0 {
			return models.Workout{}, fmt.Errorf("no combo in round %d", roundResp.RoundNumber)
		}
		comboJSON := roundResp.Combo
		moves := make([]models.Move, 0, len(comboJSON.Moves))
		for _, moveNum := range comboJSON.Moves {
			move, ok := lg.moveMapping.GetMoveFromNumber(moveNum)
			if !ok {
				return models.Workout{}, fmt.Errorf("invalid move number: %d", moveNum)
			}
			moves = append(moves, move)
		}

		// Validate move count against pattern constraints
		totalMoves := len(moves)

		// Always validate min/max bounds first
		if totalMoves < pattern.MinMoves {
			return models.Workout{}, fmt.Errorf("round %d: combo has %d moves, but minimum is %d", roundResp.RoundNumber, totalMoves, pattern.MinMoves)
		}
		if totalMoves > pattern.MaxMoves {
			return models.Workout{}, fmt.Errorf("round %d: combo has %d moves, but maximum is %d", roundResp.RoundNumber, totalMoves, pattern.MaxMoves)
		}

		// For linear pattern with multiple rounds, enforce non-decreasing progression
		if pattern.Type == models.PatternLinear && config.TotalRounds > 1 {
			expectedMoves := pattern.GetMovesPerRound(roundResp.RoundNumber, config.TotalRounds)
			// Allow ±1 tolerance for rounding in linear progression
			if totalMoves < expectedMoves-1 || totalMoves > expectedMoves+1 {
				return models.Workout{}, fmt.Errorf("round %d: combo has %d moves, but linear pattern requires %d moves (with ±1 tolerance) for this round", roundResp.RoundNumber, totalMoves, expectedMoves)
			}
			// For strict linear progression, ensure it's not decreasing compared to previous round
			if roundResp.RoundNumber > 1 {
				// Find previous round's move count
				prevRoundIdx := roundResp.RoundNumber - 2 // Convert to 0-indexed
				if prevRoundIdx < len(rounds) {
					prevTotalMoves := len(rounds[prevRoundIdx].Combo.Moves)
					if totalMoves < prevTotalMoves {
						return models.Workout{}, fmt.Errorf("round %d: combo has %d moves, but linear pattern requires non-decreasing move count (previous round had %d moves)", roundResp.RoundNumber, totalMoves, prevTotalMoves)
					}
				}
			}
		}

		combo := models.NewCombo(moves)

		round := models.NewWorkoutRound(
			roundResp.RoundNumber,
			combo,
			config.WorkDuration,
			config.RestDuration,
		)
		rounds = append(rounds, round)
	}

	return models.NewWorkout(config, rounds), nil
}

// buildWorkoutPrompt constructs the prompt for OpenAI with stance information
func (lg *LLMWorkoutGenerator) buildWorkoutPrompt(config models.WorkoutConfig, pattern models.WorkoutPattern, stance models.Stance) string {
	var sb strings.Builder

	sb.WriteString("Generate a boxing workout with the following specifications:\n\n")

	// Move mappings with stance-specific punch names
	sb.WriteString(lg.moveMapping.GetMappingDescriptionWithStance(stance))
	sb.WriteString("\n")

	// Workout configuration
	sb.WriteString("Workout Configuration:")
	sb.WriteString(fmt.Sprintf("- Total Rounds: %d\n", config.TotalRounds))
	sb.WriteString(fmt.Sprintf("- Work Duration: %.0f seconds\n", config.WorkDuration.Seconds()))
	sb.WriteString(fmt.Sprintf("- Rest Duration: %.0f seconds\n", config.RestDuration.Seconds()))
	sb.WriteString("\n")

	// Pattern description
	sb.WriteString(fmt.Sprintf("Combo Pattern: %s\n", pattern.Type))
	sb.WriteString(fmt.Sprintf("- Minimum moves per combo: %d (total moves including punches and defensive moves)\n", pattern.MinMoves))
	sb.WriteString(fmt.Sprintf("- Maximum moves per combo: %d (total moves including punches and defensive moves)\n", pattern.MaxMoves))
	sb.WriteString(fmt.Sprintf("- Include defensive moves: %v\n", pattern.IncludeDefensive))
	sb.WriteString("\n")
	sb.WriteString("CRITICAL: The min/max moves limits refer to the TOTAL number of moves in each combo (punches + defensive moves combined). ")
	sb.WriteString(fmt.Sprintf("Each combo must have between %d and %d total moves. ", pattern.MinMoves, pattern.MaxMoves))
	if pattern.IncludeDefensive {
		sb.WriteString("If defensive moves are included, they count toward the total move count.\n\n")
	} else {
		sb.WriteString("Since defensive moves are disabled, all moves must be punches.\n\n")
	}

	// Pattern-specific instructions
	switch pattern.Type {
	case models.PatternLinear:
		sb.WriteString("CRITICAL: For LINEAR pattern, the number of moves MUST increase strictly from round 1 to the final round.\n")
		sb.WriteString("Round 1 must have the minimum number of moves, and the final round must have the maximum number of moves.\n")
		sb.WriteString("Each subsequent round should have equal or more moves than the previous round.\n\n")
	case models.PatternPyramid:
		sb.WriteString("CRITICAL: For PYRAMID pattern, the number of moves should increase to a peak in the middle rounds, then decrease.\n")
		sb.WriteString("The middle round(s) should have the maximum number of moves.\n\n")
	case models.PatternRandom:
		sb.WriteString("The combo complexity should vary randomly across rounds, but stay within the min-max range. Make it interesting and unpredictable.\n\n")
	case models.PatternConstant:
		sb.WriteString("CRITICAL: For CONSTANT pattern, all rounds should have approximately the same number of moves.\n\n")
	}

	// Calculate exact moves per round and make it mandatory
	sb.WriteString("MANDATORY: Each round MUST have the EXACT number of moves specified below:\n")
	for i := 1; i <= config.TotalRounds; i++ {
		moves := pattern.GetMovesPerRound(i, config.TotalRounds)
		sb.WriteString(fmt.Sprintf("  Round %d: EXACTLY %d moves (total moves including punches and defensive moves)\n", i, moves))
	}
	sb.WriteString("\n")

	// JSON format specification
	sb.WriteString("Return the workout in the following JSON format:\n")
	if pattern.IncludeDefensive {
		sb.WriteString(`{
  "rounds": [
    {
      "round_number": 1,
      "combo": {"moves": [1, 2, 3]}
    },
    {
      "round_number": 2,
      "combo": {"moves": [1, 2, 7, 3, 4]}
    }
  ]
}`)
	} else {
		sb.WriteString(`{
  "rounds": [
    {
      "round_number": 1,
      "combo": {"moves": [1, 2, 3]}
    },
    {
      "round_number": 2,
      "combo": {"moves": [1, 2, 3, 4]}
    }
  ]
}`)
	}
	sb.WriteString("\n\n")
	sb.WriteString("Important:\n")
	sb.WriteString("- Each round should have exactly 1 combo (one combo per round)\n")
	sb.WriteString("- Combos should be realistic boxing combinations for ")
	if stance == models.Southpaw {
		sb.WriteString("a southpaw (left-handed) boxer")
	} else {
		sb.WriteString("an orthodox (right-handed) boxer")
	}
	sb.WriteString("\n")
	if pattern.IncludeDefensive {
		sb.WriteString("- Use numbers 1-6 for punches, 7-12 for defensive moves\n")
		sb.WriteString(fmt.Sprintf("- IMPORTANT: Each combo must have between %d and %d TOTAL moves (punches + defensive moves combined)\n", pattern.MinMoves, pattern.MaxMoves))
		sb.WriteString("- Defensive moves should be paired appropriately with punches for the stance:\n")
		if stance == models.Southpaw {
			sb.WriteString("  * Left Slip (7) is followed by left-hand punches (Cross, Left Hook, Left Uppercut)\n")
			sb.WriteString("  * Right Slip (8) is followed by right-hand punches (Jab, right Hook, right Uppercut)\n")
			sb.WriteString("  * Left Roll (9) is followed by left-hand punches\n")
			sb.WriteString("  * Right Roll (10) is followed by right-hand punches\n")
			sb.WriteString("  * Pull Back (11) and Duck (12) can be used with any punch sequence\n")
		} else {
			sb.WriteString("  * Left Slip (7) is followed by left-hand punches (Jab, Left Hook, Left Uppercut)\n")
			sb.WriteString("  * Right Slip (8) is followed by right-hand punches (Cross, Right Hook, Right Uppercut)\n")
			sb.WriteString("  * Left Roll (9) is followed by left-hand punches\n")
			sb.WriteString("  * Right Roll (10) is followed by right-hand punches\n")
			sb.WriteString("  * Pull Back (11) and Duck (12) can be used with any punch sequence\n")
		}
		sb.WriteString("- Defensive moves should be used strategically (not every combo needs them)\n")
	} else {
		sb.WriteString("- Use ONLY numbers 1-6 for punches. DO NOT use defensive moves (numbers 7-12)\n")
		sb.WriteString("- All combos must consist of punches only, no defensive moves\n")
	}
	sb.WriteString("- Return ONLY valid JSON, no additional text or explanation\n")

	return sb.String()
}
