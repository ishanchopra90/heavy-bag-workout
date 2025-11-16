package timer

import (
	"sync"
	"testing"
	"time"
)

func TestCountdownTimer_StartAndComplete(t *testing.T) {
	timer := NewCountdownTimer(2 * time.Second)
	completed := false
	var lastTick time.Duration

	timer.OnTick(func(remaining time.Duration) {
		lastTick = remaining
	}).OnComplete(func() {
		completed = true
	})

	// After the timer completes, validate that lastTick reached zero (or close, given possible race)
	// We'll do this after waiting for timer completion.

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting timer: %v", err)
	}

	// Wait for completion
	time.Sleep(2500 * time.Millisecond)

	if !completed {
		t.Fatalf("timer should have completed")
	}

	if timer.State() != StateCompleted {
		t.Fatalf("expected state Completed, got %v", timer.State())
	}

	if timer.Remaining() != 0 {
		t.Fatalf("expected remaining time 0, got %v", timer.Remaining())
	}

	if lastTick != 0 {
		t.Fatalf("expected last tick to be 0, got %v", lastTick)
	}
}

func TestCountdownTimer_PauseAndResume(t *testing.T) {
	timer := NewCountdownTimer(3 * time.Second)
	tickCount := 0
	var mu sync.Mutex

	timer.OnTick(func(remaining time.Duration) {
		mu.Lock()
		tickCount++
		mu.Unlock()
	})

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting timer: %v", err)
	}

	// Let it run for 1 second
	time.Sleep(1100 * time.Millisecond)

	timer.Pause()

	mu.Lock()
	ticksBeforePause := tickCount
	mu.Unlock()

	if timer.State() != StatePaused {
		t.Fatalf("expected state Paused, got %v", timer.State())
	}

	// Wait a bit to ensure it's paused
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	ticksAfterPause := tickCount
	mu.Unlock()

	if ticksAfterPause != ticksBeforePause {
		t.Fatalf("timer should not tick while paused, got %d ticks after pause", ticksAfterPause-ticksBeforePause)
	}

	// Resume
	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error resuming timer: %v", err)
	}

	if timer.State() != StateRunning {
		t.Fatalf("expected state Running, got %v", timer.State())
	}

	// Let it run a bit more
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	ticksAfterResume := tickCount
	mu.Unlock()

	if ticksAfterResume <= ticksAfterPause {
		t.Fatalf("timer should continue ticking after resume")
	}
}

func TestCountdownTimer_Stop(t *testing.T) {
	timer := NewCountdownTimer(5 * time.Second)
	completed := false

	timer.OnComplete(func() {
		completed = true
	})

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting timer: %v", err)
	}

	time.Sleep(1 * time.Second)

	timer.Stop()

	if timer.State() != StateIdle {
		t.Fatalf("expected state Idle, got %v", timer.State())
	}

	time.Sleep(2 * time.Second)

	if completed {
		t.Fatalf("timer should not complete after being stopped")
	}

	if timer.Remaining() != timer.Duration() {
		t.Fatalf("timer should be reset to full duration after stop")
	}
}

func TestCountdownTimer_Reset(t *testing.T) {
	timer := NewCountdownTimer(3 * time.Second)

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting timer: %v", err)
	}

	time.Sleep(1 * time.Second)

	timer.Reset()

	if timer.State() != StateIdle {
		t.Fatalf("expected state Idle after reset, got %v", timer.State())
	}

	if timer.Remaining() != timer.Duration() {
		t.Fatalf("expected remaining time to equal duration after reset, got %v", timer.Remaining())
	}
}

func TestWorkPeriodTimer(t *testing.T) {
	workTimer := NewWorkPeriodTimer(2 * time.Second)
	completed := false

	workTimer.OnComplete(func() {
		completed = true
	})

	if err := workTimer.Start(); err != nil {
		t.Fatalf("unexpected error starting work timer: %v", err)
	}

	time.Sleep(2500 * time.Millisecond)

	if !completed {
		t.Fatalf("work timer should have completed")
	}
}

func TestRestPeriodTimer(t *testing.T) {
	restTimer := NewRestPeriodTimer(1 * time.Second)
	completed := false

	restTimer.OnComplete(func() {
		completed = true
	})

	if err := restTimer.Start(); err != nil {
		t.Fatalf("unexpected error starting rest timer: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	if !completed {
		t.Fatalf("rest timer should have completed")
	}
}

func TestCountdownTimer_RemainingTime(t *testing.T) {
	timer := NewCountdownTimer(5 * time.Second)

	if err := timer.Start(); err != nil {
		t.Fatalf("unexpected error starting timer: %v", err)
	}

	time.Sleep(2 * time.Second)

	remaining := timer.Remaining()
	if remaining > 3*time.Second+500*time.Millisecond || remaining < 2*time.Second+500*time.Millisecond {
		t.Fatalf("expected remaining time around 3s, got %v", remaining)
	}
}
