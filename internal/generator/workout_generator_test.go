package generator

import (
	"heavybagworkout/internal/mocks"
	"heavybagworkout/internal/models"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestWorkoutGeneratorGenerateWorkout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGen := mocks.NewMockcombosForWorkPeriodGenerator(ctrl)
	combo1 := models.NewCombo([]models.Move{models.NewPunchMove(models.Jab)})
	combo2 := models.NewCombo([]models.Move{models.NewPunchMove(models.Cross)})

	config := models.NewWorkoutConfig(20*time.Second, 10*time.Second, 2)
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, false)

	// Expect calls for each round
	// First round has no previous move count (nil)
	mockGen.EXPECT().
		GenerateCombosForWorkPeriod(1, 2, config.WorkDuration, pattern, nil).
		Return(combo1)
	// Second round has previous move count from first round
	prevMoveCount := combo1.Length()
	mockGen.EXPECT().
		GenerateCombosForWorkPeriod(2, 2, config.WorkDuration, pattern, gomock.Any()).
		DoAndReturn(func(roundNumber, totalRounds int, workDuration time.Duration, pattern models.WorkoutPattern, prevCount *int) models.Combo {
			if prevCount == nil || *prevCount != prevMoveCount {
				t.Errorf("expected previous move count %d, got %v", prevMoveCount, prevCount)
			}
			return combo2
		})

	wg := NewWorkoutGeneratorWithFactory(func(includeDefensive bool) combosForWorkPeriodGenerator {
		return mockGen
	})

	workout, err := wg.GenerateWorkout(config, pattern)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if workout.RoundCount() != 2 {
		t.Fatalf("expected 2 rounds, got %d", workout.RoundCount())
	}

	if duration := wg.CalculateTotalWorkoutDuration(config); duration != 60*time.Second {
		t.Fatalf("expected duration 60s, got %s", duration)
	}
}

func TestWorkoutGeneratorDistributeCombosAcrossWorkPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGen := mocks.NewMockcombosForWorkPeriodGenerator(ctrl)
	expectedCombo := models.NewCombo([]models.Move{models.NewPunchMove(models.Jab)})

	config := models.NewWorkoutConfig(20*time.Second, 10*time.Second, 1)
	pattern := models.NewWorkoutPattern(models.PatternPyramid, 1, 4, true)

	mockGen.EXPECT().
		GenerateCombosForWorkPeriod(1, 1, config.WorkDuration, pattern, nil).
		Return(expectedCombo)

	wg := NewWorkoutGeneratorWithFactory(func(includeDefensive bool) combosForWorkPeriodGenerator {
		return mockGen
	})

	roundCombo := wg.distributeCombosAcrossWorkPeriod(mockGen, 1, config, pattern, nil)

	// Should return exactly 1 combo
	if roundCombo.Length() != expectedCombo.Length() {
		t.Fatalf("expected combo length %d, got %d", expectedCombo.Length(), roundCombo.Length())
	}
}

func TestWorkoutGeneratorInvalidConfig(t *testing.T) {
	wg := NewWorkoutGenerator()
	config := models.NewWorkoutConfig(0, 10*time.Second, 2)
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, false)

	if _, err := wg.GenerateWorkout(config, pattern); err == nil {
		t.Fatalf("expected validation error for invalid config")
	}
}

func TestWorkoutPresetConfigurations(t *testing.T) {
	// Test Beta Style preset
	betaConfig := models.PresetWorkoutConfig(models.PresetBetaStyle)
	if betaConfig.WorkDuration != 20*time.Second {
		t.Fatalf("expected 20s work duration for beta style, got %v", betaConfig.WorkDuration)
	}
	if betaConfig.RestDuration != 10*time.Second {
		t.Fatalf("expected 10s rest duration for beta style, got %v", betaConfig.RestDuration)
	}
	if betaConfig.TotalRounds != 8 {
		t.Fatalf("expected 8 rounds for beta style, got %d", betaConfig.TotalRounds)
	}

	// Test Endurance preset
	enduranceConfig := models.PresetWorkoutConfig(models.PresetEndurance)
	if enduranceConfig.WorkDuration != 40*time.Second {
		t.Fatalf("expected 40s work duration for endurance, got %v", enduranceConfig.WorkDuration)
	}
	if enduranceConfig.RestDuration != 20*time.Second {
		t.Fatalf("expected 20s rest duration for endurance, got %v", enduranceConfig.RestDuration)
	}
	if enduranceConfig.TotalRounds != 10 {
		t.Fatalf("expected 10 rounds for endurance, got %d", enduranceConfig.TotalRounds)
	}

	// Test Power preset
	powerConfig := models.PresetWorkoutConfig(models.PresetPower)
	if powerConfig.WorkDuration != 30*time.Second {
		t.Fatalf("expected 30s work duration for power, got %v", powerConfig.WorkDuration)
	}
	if powerConfig.RestDuration != 15*time.Second {
		t.Fatalf("expected 15s rest duration for power, got %v", powerConfig.RestDuration)
	}
	if powerConfig.TotalRounds != 8 {
		t.Fatalf("expected 8 rounds for power, got %d", powerConfig.TotalRounds)
	}

	// Test that preset configs can generate workouts
	wg := NewWorkoutGenerator()
	pattern := models.NewWorkoutPattern(models.PatternLinear, 1, 3, false)

	workout, err := wg.GenerateWorkout(betaConfig, pattern)
	if err != nil {
		t.Fatalf("unexpected error generating workout with beta preset: %v", err)
	}
	if workout.RoundCount() != 8 {
		t.Fatalf("expected 8 rounds, got %d", workout.RoundCount())
	}
}
