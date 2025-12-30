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

	// First attempt
	workout, err := lg.generateWorkoutAttempt(prompt, config, pattern, stance, "")
	if err == nil {
		return workout, nil
	}

	// If first attempt failed, retry once with the error message
	retryPrompt := lg.buildWorkoutPromptWithError(config, pattern, stance, err.Error())
	workout, retryErr := lg.generateWorkoutAttempt(retryPrompt, config, pattern, stance, err.Error())
	if retryErr != nil {
		return models.Workout{}, fmt.Errorf("failed to generate workout after retry: first attempt error: %v, retry error: %w", err, retryErr)
	}

	return workout, nil
}

// generateWorkoutAttempt attempts to generate a workout from a prompt
func (lg *LLMWorkoutGenerator) generateWorkoutAttempt(prompt string, config models.WorkoutConfig, pattern models.WorkoutPattern, stance models.Stance, previousError string) (models.Workout, error) {
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

	// Validate number of rounds matches configuration
	if len(workoutResp.Rounds) != config.TotalRounds {
		return models.Workout{}, fmt.Errorf("workout has %d rounds, but configuration specifies %d rounds. The LLM must generate exactly %d rounds.", len(workoutResp.Rounds), config.TotalRounds, config.TotalRounds)
	}

	// Convert JSON response to Workout model
	// Sort rounds by round number to ensure correct order (LLM might return them out of order)
	rounds := make([]models.WorkoutRound, 0, config.TotalRounds)

	// First, validate all round numbers are valid and collect them
	roundMap := make(map[int]RoundResponseJSON)
	for _, roundResp := range workoutResp.Rounds {
		// Validate round number is within expected range
		if roundResp.RoundNumber < 1 || roundResp.RoundNumber > config.TotalRounds {
			return models.Workout{}, fmt.Errorf("invalid round number %d (expected 1-%d)", roundResp.RoundNumber, config.TotalRounds)
		}
		// Check for duplicate round numbers
		if _, exists := roundMap[roundResp.RoundNumber]; exists {
			return models.Workout{}, fmt.Errorf("duplicate round number %d found in LLM response", roundResp.RoundNumber)
		}
		roundMap[roundResp.RoundNumber] = roundResp
	}

	// Process rounds in order (1, 2, 3, ..., TotalRounds)
	for roundNumber := 1; roundNumber <= config.TotalRounds; roundNumber++ {
		roundResp, exists := roundMap[roundNumber]
		if !exists {
			return models.Workout{}, fmt.Errorf("missing round number %d in LLM response (expected rounds 1-%d)", roundNumber, config.TotalRounds)
		}
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

			// Calculate acceptable range with ±1 tolerance
			minAllowed := expectedMoves - 1
			if minAllowed < pattern.MinMoves {
				minAllowed = pattern.MinMoves
			}
			maxAllowed := expectedMoves + 1
			if maxAllowed > pattern.MaxMoves {
				maxAllowed = pattern.MaxMoves
			}

			// Adjust for non-decreasing requirement: must be >= previous round
			if roundResp.RoundNumber > 1 {
				prevRoundIdx := roundResp.RoundNumber - 2 // Convert to 0-indexed
				if prevRoundIdx < len(rounds) {
					prevTotalMoves := len(rounds[prevRoundIdx].Combo.Moves)
					// Must be at least as many moves as previous round
					if prevTotalMoves > minAllowed {
						minAllowed = prevTotalMoves
					}
					// If non-decreasing forces a higher minimum, allow up to maxMoves
					// (not just expectedMoves + 1) to give the LLM flexibility
					if minAllowed > expectedMoves {
						maxAllowed = pattern.MaxMoves
					}
				}
			}

			// Validate against the adjusted range
			if totalMoves < minAllowed || totalMoves > maxAllowed {
				return models.Workout{}, fmt.Errorf("round %d: combo has %d moves, but linear pattern requires %d-%d moves (target: %d with ±1 tolerance, must be ≥ previous round)", roundResp.RoundNumber, totalMoves, minAllowed, maxAllowed, expectedMoves)
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

	// Final validation: ensure we have exactly the expected number of rounds
	// (This should never fail since we validate above, but adding as a safety check)
	if len(rounds) != config.TotalRounds {
		return models.Workout{}, fmt.Errorf("generated %d rounds, but configuration requires %d rounds", len(rounds), config.TotalRounds)
	}

	return models.NewWorkout(config, rounds), nil
}

// buildWorkoutPrompt constructs the prompt for OpenAI with stance information
func (lg *LLMWorkoutGenerator) buildWorkoutPrompt(config models.WorkoutConfig, pattern models.WorkoutPattern, stance models.Stance) string {
	return lg.buildWorkoutPromptWithError(config, pattern, stance, "")
}

// buildWorkoutPromptWithError constructs the prompt for OpenAI with stance information and optional error message
func (lg *LLMWorkoutGenerator) buildWorkoutPromptWithError(config models.WorkoutConfig, pattern models.WorkoutPattern, stance models.Stance, previousError string) string {
	var sb strings.Builder

	// If this is a retry, add error information at the top
	if previousError != "" {
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("PREVIOUS ATTEMPT FAILED - PLEASE FIX THE FOLLOWING ERROR:\n")
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString(previousError)
		sb.WriteString("\n")
		sb.WriteString("\n")
		sb.WriteString("The workout you generated violated the constraints above. Please regenerate the workout, paying special attention to the error message.\n")
		sb.WriteString("Make sure to:\n")
		if strings.Contains(previousError, "rounds, but configuration specifies") || strings.Contains(previousError, "missing round number") {
			sb.WriteString(fmt.Sprintf("  - Generate EXACTLY %d rounds, no more and no less\n", config.TotalRounds))
			sb.WriteString(fmt.Sprintf("  - Include round numbers 1 through %d in your JSON response\n", config.TotalRounds))
			sb.WriteString("  - Do not skip any round numbers\n")
			sb.WriteString("  - Do not include extra rounds beyond what was requested\n")
		}
		if strings.Contains(previousError, "non-decreasing") || strings.Contains(previousError, "≥ previous round") {
			sb.WriteString("  - Check each round against the previous round's move count\n")
			sb.WriteString("  - Ensure each round has equal or MORE moves than the previous round\n")
			sb.WriteString("  - If a previous round had 4 moves, the current round MUST also have 4 moves\n")
		}
		if strings.Contains(previousError, "maximum is") {
			sb.WriteString(fmt.Sprintf("  - Never exceed %d moves in any combo\n", pattern.MaxMoves))
		}
		if strings.Contains(previousError, "minimum is") {
			sb.WriteString(fmt.Sprintf("  - Never go below %d moves in any combo\n", pattern.MinMoves))
		}
		sb.WriteString("\n")
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("\n")
	}

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
	sb.WriteString(fmt.Sprintf("CRITICAL: You MUST generate EXACTLY %d rounds. No more, no less. ", config.TotalRounds))
	sb.WriteString(fmt.Sprintf("If you generate %d rounds, the workout will fail validation.\n", config.TotalRounds+1))
	sb.WriteString(fmt.Sprintf("If you generate %d rounds, the workout will fail validation.\n", config.TotalRounds-1))
	sb.WriteString(fmt.Sprintf("The JSON response must contain exactly %d round objects in the 'rounds' array.\n\n", config.TotalRounds))

	// Pattern description
	sb.WriteString(fmt.Sprintf("Combo Pattern: %s\n", pattern.Type))
	sb.WriteString(fmt.Sprintf("- Minimum moves per combo: %d (total moves including punches and defensive moves)\n", pattern.MinMoves))
	sb.WriteString(fmt.Sprintf("- Maximum moves per combo: %d (total moves including punches and defensive moves)\n", pattern.MaxMoves))
	sb.WriteString(fmt.Sprintf("- Include defensive moves: %v\n", pattern.IncludeDefensive))
	sb.WriteString("\n")
	sb.WriteString("CRITICAL: The min/max moves limits refer to the TOTAL number of moves in each combo (punches + defensive moves combined). ")
	sb.WriteString(fmt.Sprintf("Each combo must have between %d and %d total moves. ", pattern.MinMoves, pattern.MaxMoves))
	sb.WriteString(fmt.Sprintf("ABSOLUTE HARD LIMIT: NO combo can have more than %d moves, regardless of pattern, tolerance, or any other factor. ", pattern.MaxMoves))
	sb.WriteString(fmt.Sprintf("If you generate a combo with more than %d moves, the workout will fail validation.\n", pattern.MaxMoves))
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
	sb.WriteString("MANDATORY: Each round MUST have the target number of moves specified below:\n")
	sb.WriteString(fmt.Sprintf("ABSOLUTE LIMIT: Every combo must have between %d and %d moves (inclusive). Never exceed %d moves in any combo.\n\n", pattern.MinMoves, pattern.MaxMoves, pattern.MaxMoves))
	if pattern.Type == models.PatternLinear {
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("CRITICAL RULE FOR LINEAR PATTERN (READ THIS FIRST):\n")
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("SIMPLE FORMULA: Round N moves = MAX(target_range_min, previous_round_moves)\n")
		sb.WriteString("\n")
		sb.WriteString("This means:\n")
		sb.WriteString("  - Round 1: Use target range\n")
		sb.WriteString("  - Round 2+: Use the HIGHER of (target range minimum) OR (previous round's moves)\n")
		sb.WriteString("  - You can go up to the target range maximum OR pattern maximum, whichever applies\n")
		sb.WriteString("\n")
		sb.WriteString("EXAMPLES:\n")
		sb.WriteString("  - Round 1 target is 1 (range 1-2) → Use 1 or 2 moves\n")
		sb.WriteString("  - Round 2 target is 1 (range 1-2), Round 1 had 2 → Round 2 MUST have ≥2, so use 2-4 moves\n")
		sb.WriteString(fmt.Sprintf("  - Round 5 target is 1 (range 1-2), Round 4 had %d → Round 5 MUST have %d moves\n", pattern.MaxMoves, pattern.MaxMoves))
		sb.WriteString(fmt.Sprintf("  - Round 18 target is 2 (range 1-3), Round 17 had 4 → Round 18 MUST have 4 moves (not 3!)\n"))
		sb.WriteString("\n")
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("\n")
		sb.WriteString("For LINEAR pattern, you have ±1 tolerance from the target (e.g., if target is 1, you can use 1-2 moves; if target is 2, you can use 1-3 moves).\n")
		sb.WriteString(fmt.Sprintf("BUT REMEMBER: Even with tolerance, you can NEVER exceed %d moves total in any combo.\n", pattern.MaxMoves))
		sb.WriteString("CRITICAL NON-DECREASING RULE: Each round MUST have equal or MORE moves than the previous round. This is MANDATORY.\n")
		sb.WriteString("\n")
		sb.WriteString("HOW TO APPLY THIS RULE (READ CAREFULLY):\n")
		sb.WriteString("1. Generate rounds SEQUENTIALLY, checking each one against the previous\n")
		sb.WriteString("2. For Round 1: Use the target range shown (e.g., if target is 1, use 1-2 moves)\n")
		sb.WriteString("3. For Round 2 and beyond:\n")
		sb.WriteString("   STEP A: Check how many moves the PREVIOUS round has in your JSON\n")
		sb.WriteString("   STEP B: Check the target range for the current round\n")
		sb.WriteString("   STEP C: The current round MUST have moves ≥ previous round's moves\n")
		sb.WriteString("   STEP D: If previous round had more than target_max, current round MUST equal previous round\n")
		sb.WriteString("\n")
		sb.WriteString("CRITICAL EXAMPLE - Read this carefully:\n")
		sb.WriteString(fmt.Sprintf("If Round 8 has 4 moves (the maximum %d), then:\n", pattern.MaxMoves))
		sb.WriteString("  - Round 9 target might be 1 (range 1-2), BUT Round 9 MUST have 4 moves (≥ Round 8)\n")
		sb.WriteString("  - Round 10 target might be 1 (range 1-2), BUT Round 10 MUST have 4 moves (≥ Round 9)\n")
		sb.WriteString("  - Round 11, 12, 13... ALL must have 4 moves (cannot decrease from 4)\n")
		sb.WriteString("\n")
		sb.WriteString("WORKFLOW:\n")
		sb.WriteString("1. Generate Round 1 with moves from its target range\n")
		sb.WriteString("2. For Round 2: Check Round 1's moves, ensure Round 2 ≥ Round 1\n")
		sb.WriteString("3. For Round 3: Check Round 2's moves, ensure Round 3 ≥ Round 2\n")
		sb.WriteString("4. Continue this pattern for ALL rounds\n")
		sb.WriteString(fmt.Sprintf("5. If ANY round reaches %d moves, ALL following rounds must be %d moves\n", pattern.MaxMoves, pattern.MaxMoves))
		sb.WriteString("\n")
	}
	for i := 1; i <= config.TotalRounds; i++ {
		moves := pattern.GetMovesPerRound(i, config.TotalRounds)
		if pattern.Type == models.PatternLinear {
			minAllowed := moves - 1
			if minAllowed < pattern.MinMoves {
				minAllowed = pattern.MinMoves
			}
			maxAllowed := moves + 1
			if maxAllowed > pattern.MaxMoves {
				maxAllowed = pattern.MaxMoves
			}
			if i == 1 {
				sb.WriteString(fmt.Sprintf("  Round %d: Target %d moves (use %d-%d moves)\n", i, moves, minAllowed, maxAllowed))
			} else {
				sb.WriteString(fmt.Sprintf("  Round %d: Target %d moves (base range: %d-%d, but ACTUAL moves = MAX(%d, Round %d's moves))\n", i, moves, minAllowed, maxAllowed, minAllowed, i-1))
				sb.WriteString(fmt.Sprintf("           → If Round %d had 4 moves, Round %d MUST have 4 moves (cannot use 3 even if target allows it!)\n", i-1, i))
			}
		} else {
			sb.WriteString(fmt.Sprintf("  Round %d: Target %d moves (total moves including punches and defensive moves)\n", i, moves))
		}
	}
	sb.WriteString("\n")

	// JSON format specification
	sb.WriteString("\n")
	if pattern.Type == models.PatternLinear {
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("BEFORE GENERATING JSON - READ THIS ONE MORE TIME:\n")
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("For EACH round after Round 1, you MUST check:\n")
		sb.WriteString("  1. How many moves did the PREVIOUS round have?\n")
		sb.WriteString("  2. What is the target range for the CURRENT round?\n")
		sb.WriteString("  3. Current round moves = MAX(target_min, previous_round_moves)\n")
		sb.WriteString("\n")
		sb.WriteString("CRITICAL: If previous round had 4 moves, current round MUST have 4 moves.\n")
		sb.WriteString("          Even if the target says 1-2 or 1-3, you MUST use 4 moves.\n")
		sb.WriteString("\n")
		sb.WriteString("VALIDATION CHECKLIST (verify each round):\n")
		sb.WriteString("□ Round 1: Use target range\n")
		sb.WriteString("□ Round 2: moves ≥ Round 1's moves?\n")
		sb.WriteString("□ Round 3: moves ≥ Round 2's moves?\n")
		sb.WriteString("□ Round 4: moves ≥ Round 3's moves?\n")
		sb.WriteString("□ ... continue for ALL rounds ...\n")
		sb.WriteString(fmt.Sprintf("□ If Round N has %d moves, Round N+1 MUST have %d moves\n", pattern.MaxMoves, pattern.MaxMoves))
		sb.WriteString("═══════════════════════════════════════════════════════════════\n")
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("Return the workout in the following JSON format with EXACTLY %d rounds:\n", config.TotalRounds))
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
		if pattern.Type == models.PatternLinear {
			sb.WriteString(`{
  "rounds": [
    {
      "round_number": 1,
      "combo": {"moves": [1]}
    },
    {
      "round_number": 2,
      "combo": {"moves": [1, 2]}
    },
    {
      "round_number": 3,
      "combo": {"moves": [1, 2, 3]}
    },
    {
      "round_number": 4,
      "combo": {"moves": [1, 2, 3, 4]}
    },
    {
      "round_number": 5,
      "combo": {"moves": [1, 2, 3, 4]}
    },
    {
      "round_number": 12,
      "combo": {"moves": [1, 2, 3, 4]}
    },
    {
      "round_number": 13,
      "combo": {"moves": [1, 2, 3, 4]}
    }
  ]
}`)
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("IMPORTANT: In this example:\n"))
			sb.WriteString(fmt.Sprintf("  - Round 4 reached the maximum (%d moves)\n", pattern.MaxMoves))
			sb.WriteString(fmt.Sprintf("  - Round 5's target might be 1 (range 1-2), but it MUST have %d moves (≥ Round 4)\n", pattern.MaxMoves))
			sb.WriteString(fmt.Sprintf("  - Round 12 has %d moves (even though its target might be lower)\n", pattern.MaxMoves))
			sb.WriteString(fmt.Sprintf("  - Round 13's target might be 1 (range 1-2), but it MUST have %d moves (≥ Round 12)\n", pattern.MaxMoves))
			sb.WriteString(fmt.Sprintf("  - Round 13 CANNOT have 3 moves even if target allows 1-3, because Round 12 had %d moves!\n", pattern.MaxMoves))
		} else {
			sb.WriteString(`{
  "rounds": [
    {
      "round_number": 1,
      "combo": {"moves": [1, 2, 3]}
    },
    {
      "round_number": 2,
      "combo": {"moves": [1, 2, 5, 3]}
    },
    {
      "round_number": 3,
      "combo": {"moves": [1, 6, 2, 4]}
    }
  ]
}`)
		}
	}
	sb.WriteString("\n\n")
	sb.WriteString("You are an experienced boxing trainer designing a workout. Use your expertise to create effective combinations.\n\n")
	sb.WriteString("Guidelines:\n")
	sb.WriteString(fmt.Sprintf("- CRITICAL: Every combo must have between %d and %d moves (inclusive). Never exceed %d moves in any combo.\n", pattern.MinMoves, pattern.MaxMoves, pattern.MaxMoves))
	sb.WriteString("- Each round should have exactly 1 combo (one combo per round)\n")
	sb.WriteString("- Design realistic boxing combinations appropriate for ")
	if stance == models.Southpaw {
		sb.WriteString("a southpaw (left-handed) boxer")
	} else {
		sb.WriteString("an orthodox (right-handed) boxer")
	}
	sb.WriteString("\n")
	sb.WriteString("- As a trainer, you know that good workouts include variety - use your judgment to incorporate different punch types (1-6) throughout the workout\n")
	sb.WriteString("- Consider including uppercuts (numbers 5 and 6) where they make sense in combinations, especially as combos get longer\n")
	sb.WriteString("- Think about what combinations would be most effective for training - mix up straight punches, hooks, and uppercuts naturally\n")
	if pattern.IncludeDefensive {
		sb.WriteString("- Use numbers 1-6 for punches, 7-12 for defensive moves\n")
		sb.WriteString(fmt.Sprintf("- Each combo should have between %d and %d TOTAL moves (punches + defensive moves combined)\n", pattern.MinMoves, pattern.MaxMoves))
		sb.WriteString("- As a trainer, you know defensive moves work best when paired appropriately with punches for the stance:\n")
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
		sb.WriteString("- Use defensive moves strategically - not every combo needs them, but they add realism when used appropriately\n")
	} else {
		sb.WriteString("- Use ONLY numbers 1-6 for punches. Do not use defensive moves (numbers 7-12)\n")
		sb.WriteString("- All combos should consist of punches only, no defensive moves\n")
		sb.WriteString("- As a trainer designing punch-only combos, consider when uppercuts (numbers 5 and 6) would enhance the combination\n")
	}
	sb.WriteString("- Return ONLY valid JSON, no additional text or explanation\n")

	return sb.String()
}
