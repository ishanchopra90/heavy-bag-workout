package generator

import (
	"errors"
	"heavybagworkout/internal/mocks"
	"heavybagworkout/internal/models"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestLLMWorkoutGeneratorGenerateWorkout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	var capturedReq *http.Request
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			capturedReq = req
			return newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[1,2,7]}}]}"}}]}`), nil
		})

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, true)

	workout, err := gen.GenerateWorkout(config, pattern)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if workout.RoundCount() != 1 {
		t.Fatalf("expected 1 round, got %d", workout.RoundCount())
	}

	round := workout.Rounds[0]
	expected := "1, 2, Left Slip"
	if comboStr := round.Combo.String(); comboStr != expected {
		t.Fatalf("expected combo %s, got %s", expected, comboStr)
	}

	if capturedReq == nil {
		t.Fatalf("expected HTTP request to be recorded")
	}
}

func TestLLMWorkoutGeneratorGenerateWorkoutWithStance_Orthodox(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	var capturedPrompt string
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			// Read the request body to check prompt contains stance info
			body, _ := io.ReadAll(req.Body)
			capturedPrompt = string(body)
			return newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[1,2]}}]}"}}]}`), nil
		})

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, false)

	workout, err := gen.GenerateWorkoutWithStance(config, pattern, models.Orthodox)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if workout.RoundCount() != 1 {
		t.Fatalf("expected 1 round, got %d", workout.RoundCount())
	}

	// Verify prompt contains orthodox stance information
	if !strings.Contains(capturedPrompt, "orthodox") {
		t.Errorf("expected prompt to contain 'orthodox', got: %s", capturedPrompt)
	}
	if !strings.Contains(capturedPrompt, "left hook") || !strings.Contains(capturedPrompt, "right hook") {
		t.Errorf("expected prompt to contain stance-specific punch names for orthodox")
	}
}

func TestLLMWorkoutGeneratorGenerateWorkoutWithStance_Southpaw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	var capturedPrompt string
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			capturedPrompt = string(body)
			return newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[1,2]}}]}"}}]}`), nil
		})

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, true)

	workout, err := gen.GenerateWorkoutWithStance(config, pattern, models.Southpaw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if workout.RoundCount() != 1 {
		t.Fatalf("expected 1 round, got %d", workout.RoundCount())
	}

	// Verify prompt contains southpaw stance information
	if !strings.Contains(capturedPrompt, "southpaw") {
		t.Errorf("expected prompt to contain 'southpaw', got: %s", capturedPrompt)
	}
	if !strings.Contains(capturedPrompt, "right hook") || !strings.Contains(capturedPrompt, "left hook") {
		t.Errorf("expected prompt to contain stance-specific punch names for southpaw")
	}
}

func TestLLMWorkoutGeneratorGenerateWorkoutWithStance_DefensiveMovePairing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	var capturedPrompt string
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			capturedPrompt = string(body)
			return newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[1,2,7]}}]}"}}]}`), nil
		})

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, true)

	_, err := gen.GenerateWorkoutWithStance(config, pattern, models.Orthodox)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify prompt contains defensive move pairing instructions
	if !strings.Contains(capturedPrompt, "Left Slip") || !strings.Contains(capturedPrompt, "Right Slip") {
		t.Errorf("expected prompt to contain defensive move pairing instructions")
	}
	if !strings.Contains(capturedPrompt, "is followed by") {
		t.Errorf("expected prompt to explain how defensive moves pair with punches (using 'is followed by')")
	}
}

func TestLLMWorkoutGeneratorGenerateWorkout_InvalidMoveNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[99]}}]}"}}]}`), nil)

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, true)

	_, err := gen.GenerateWorkout(config, pattern)
	if err == nil || err.Error() != "invalid move number: 99" {
		t.Fatalf("expected invalid move number error, got %v", err)
	}
}

func TestLLMWorkoutGeneratorGenerateWorkout_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"invalid_json\":}}"}]}`), nil)

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, true)

	_, err := gen.GenerateWorkout(config, pattern)
	if err == nil {
		t.Fatalf("expected JSON error, got nil")
	}
}

func TestLLMWorkoutGeneratorGenerateWorkout_ClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(nil, errors.New("api down"))

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  1,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, true)

	_, err := gen.GenerateWorkout(config, pattern)
	if err == nil || err.Error() != "failed to generate workout: failed to send request: api down" {
		t.Fatalf("expected api down error, got %v", err)
	}
}

func TestLLMWorkoutGenerator_OneComboPerRound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	// Mock response with 3 rounds, each with exactly 1 combo
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[1,2]}},{\"round_number\":2,\"combo\":{\"moves\":[1,2,3]}},{\"round_number\":3,\"combo\":{\"moves\":[2,3,4]}}]}"}}]}`), nil)

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 20 * time.Second,
		RestDuration: 10 * time.Second,
		TotalRounds:  3,
	}
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 4, false)

	workout, err := gen.GenerateWorkout(config, pattern)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if workout.RoundCount() != 3 {
		t.Fatalf("expected 3 rounds, got %d", workout.RoundCount())
	}

	// Verify each round has a combo (struct enforces single combo)
	for i, round := range workout.Rounds {
		if len(round.Combo.Moves) == 0 {
			t.Errorf("round %d: expected combo with moves, got empty combo", i+1)
		}
	}
}

func TestLLMWorkoutGenerator_ExcludesDefensiveMovesWhenDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	var capturedPrompt string
	mockHTTP.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			capturedPrompt = string(body)
			// Mock response with punches only (no defensive moves)
			return newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[{\"round_number\":1,\"combo\":{\"moves\":[1,2,3,4]}},{\"round_number\":2,\"combo\":{\"moves\":[1,2,3,4]}}]}"}}]}`), nil
		})

	openAIClient := NewOpenAIClientWithHTTPClient("test-key", mockHTTP)
	gen := NewLLMWorkoutGeneratorWithOpenAIClient(openAIClient)
	config := models.WorkoutConfig{
		WorkDuration: 5 * time.Second,
		RestDuration: 5 * time.Second,
		TotalRounds:  2,
	}
	// Pattern with IncludeDefensive = false
	pattern := models.NewWorkoutPattern(models.PatternLinear, 4, 4, false)

	workout, err := gen.GenerateWorkoutWithStance(config, pattern, models.Southpaw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify prompt explicitly says not to use defensive moves
	if !strings.Contains(capturedPrompt, "DO NOT use defensive moves") {
		t.Errorf("expected prompt to explicitly say 'DO NOT use defensive moves'")
	}
	if !strings.Contains(capturedPrompt, "Use ONLY numbers 1-6 for punches") {
		t.Errorf("expected prompt to say 'Use ONLY numbers 1-6 for punches'")
	}
	if !strings.Contains(capturedPrompt, "All combos must consist of punches only") {
		t.Errorf("expected prompt to say 'All combos must consist of punches only'")
	}

	// Verify prompt does NOT include defensive move pairing instructions
	if strings.Contains(capturedPrompt, "Defensive moves should be paired appropriately") {
		t.Errorf("expected prompt to NOT include defensive move pairing instructions when IncludeDefensive is false")
	}

	// Verify JSON example doesn't include defensive moves (should not have number 7)
	if strings.Contains(capturedPrompt, `"moves": [1, 2, 7`) {
		t.Errorf("expected JSON example to NOT include defensive moves (number 7)")
	}

	// Verify workout has no defensive moves
	for i, round := range workout.Rounds {
		for _, move := range round.Combo.Moves {
			if move.IsDefensive() {
				t.Errorf("round %d: expected no defensive moves, but found %v", i+1, move)
			}
		}
	}
}
