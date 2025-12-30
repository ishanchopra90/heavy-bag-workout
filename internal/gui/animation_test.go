package gui

import (
	"heavybagworkout/internal/models"
	"testing"
	"time"
)

// TestCharacterSprite_LastFrameTiming verifies that the last frame of a combo
// animation doesn't take longer than timePerMove, ensuring idle frame has time to show
func TestCharacterSprite_LastFrameTiming(t *testing.T) {
	cs := NewCharacterSprite(models.Orthodox)
	timePerMove := 100 * time.Millisecond
	cs.SetTimePerMove(timePerMove)

	// Set up a punch animation (non-looping, single frame)
	cs.SetAnimation(AnimationStateJabLeft)
	anim := cs.GetAnimation(AnimationStateJabLeft)
	if anim == nil {
		t.Fatal("expected jab animation to exist")
	}

	// Verify animation has frames
	if len(anim.Frames) == 0 {
		t.Fatal("expected animation to have frames")
	}

	// Get the move start time (set when animation starts)
	moveStartTime := cs.GetMoveStartTime()
	startTime := time.Now()

	// Simulate updates to reach the last frame
	// First, advance past the frame duration to reach the last frame
	now := startTime.Add(timePerMove)
	cs.Update(now)

	// Verify we're on the last frame
	currentFrame := cs.GetCurrentFrameIndex()
	if currentFrame != len(anim.Frames)-1 {
		t.Fatalf("expected to be on last frame after timePerMove, got frame %d", currentFrame)
	}

	// Store frameStartTime when we first reach the last frame
	frameStartTimeOnLastFrame := cs.GetFrameStartTime()

	// Now simulate multiple updates while on the last frame
	// The key is that frameStartTime should NOT be reset on subsequent updates
	// and elapsedSinceMoveStart should continue to grow accurately
	for i := 1; i <= 5; i++ {
		now = startTime.Add(timePerMove + time.Duration(i)*10*time.Millisecond)
		cs.Update(now)

		// Verify we're still on the last frame
		if cs.GetCurrentFrameIndex() != len(anim.Frames)-1 {
			t.Errorf("expected to still be on last frame, got frame %d", cs.GetCurrentFrameIndex())
		}

		// frameStartTime should NOT have changed (we're already on last frame)
		currentFrameStartTime := cs.GetFrameStartTime()
		if !currentFrameStartTime.Equal(frameStartTimeOnLastFrame) {
			t.Errorf("frameStartTime should not change when already on last frame: original=%v, current=%v",
				frameStartTimeOnLastFrame, currentFrameStartTime)
		}

		// elapsedSinceMoveStart should be accurate and growing
		elapsedSinceMoveStart := now.Sub(moveStartTime)
		expectedElapsed := timePerMove + time.Duration(i)*10*time.Millisecond
		tolerance := 5 * time.Millisecond
		if elapsedSinceMoveStart < expectedElapsed-tolerance || elapsedSinceMoveStart > expectedElapsed+tolerance {
			t.Errorf("elapsedSinceMoveStart should be ~%v, got %v", expectedElapsed, elapsedSinceMoveStart)
		}
	}

	// After timePerMove has elapsed, we should be able to transition to idle
	// The combo system checks: elapsedSinceMoveStart >= timePerMove
	finalNow := startTime.Add(timePerMove)
	cs.Update(finalNow)
	finalElapsedSinceMoveStart := finalNow.Sub(moveStartTime)

	if finalElapsedSinceMoveStart < timePerMove {
		t.Errorf("expected elapsedSinceMoveStart >= timePerMove, got %v < %v",
			finalElapsedSinceMoveStart, timePerMove)
	}
}

// TestCharacterSprite_LastFrameDoesNotResetFrameStartTime verifies that
// when already on the last frame, Update() doesn't reset frameStartTime
func TestCharacterSprite_LastFrameDoesNotResetFrameStartTime(t *testing.T) {
	cs := NewCharacterSprite(models.Orthodox)
	timePerMove := 100 * time.Millisecond
	cs.SetTimePerMove(timePerMove)

	// Set up a punch animation
	cs.SetAnimation(AnimationStateJabLeft)
	anim := cs.GetAnimation(AnimationStateJabLeft)
	if anim == nil || len(anim.Frames) == 0 {
		t.Fatal("expected animation with frames")
	}

	// Advance to the last frame
	startTime := time.Now()
	moveStartTime := cs.GetMoveStartTime()

	// First update: advance past frame duration to reach last frame
	now := startTime.Add(timePerMove)
	cs.Update(now)

	// Verify we're on the last frame
	if cs.GetCurrentFrameIndex() != len(anim.Frames)-1 {
		t.Fatalf("expected to be on last frame, got frame %d", cs.GetCurrentFrameIndex())
	}

	// Store frameStartTime after first reaching last frame
	frameStartTimeAfterFirst := cs.GetFrameStartTime()

	// Update again while already on last frame
	now = now.Add(10 * time.Millisecond)
	cs.Update(now)

	// frameStartTime should NOT have changed (we're already on last frame)
	frameStartTimeAfterSecond := cs.GetFrameStartTime()
	if !frameStartTimeAfterSecond.Equal(frameStartTimeAfterFirst) {
		t.Errorf("frameStartTime should not change when already on last frame: first=%v, second=%v",
			frameStartTimeAfterFirst, frameStartTimeAfterSecond)
	}

	// Verify elapsedSinceMoveStart is still accurate
	elapsedSinceMoveStart := now.Sub(moveStartTime)
	if elapsedSinceMoveStart < timePerMove {
		t.Errorf("elapsedSinceMoveStart should be >= timePerMove, got %v < %v",
			elapsedSinceMoveStart, timePerMove)
	}
}
