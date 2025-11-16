package timer

import (
	"context"
	"time"
)

// TimerState represents the current state of the timer
type TimerState int

const (
	StateIdle TimerState = iota
	StateRunning
	StatePaused
	StateCompleted
)

// TimerCallback is a function type for timer callbacks
// It receives the remaining time duration
type TimerCallback func(remaining time.Duration)

// CountdownTimer represents a countdown timer that can be started, paused, and stopped
type CountdownTimer struct {
	duration      time.Duration
	state         TimerState
	ctx           context.Context
	cancel        context.CancelFunc
	onTick        TimerCallback
	onComplete    func()
	ticker        *time.Ticker
	startTime     time.Time
	pausedAt      time.Duration
	initialRemain time.Duration // Tracks the remaining time when timer started/resumed
}

// NewCountdownTimer creates a new countdown timer with the specified duration
func NewCountdownTimer(duration time.Duration) *CountdownTimer {
	return &CountdownTimer{
		duration: duration,
		state:    StateIdle,
	}
}

// OnTick sets a callback function that will be called on each tick (every second)
func (t *CountdownTimer) OnTick(callback TimerCallback) *CountdownTimer {
	t.onTick = callback
	return t
}

// OnComplete sets a callback function that will be called when the timer completes
func (t *CountdownTimer) OnComplete(callback func()) *CountdownTimer {
	t.onComplete = callback
	return t
}

// Start starts the countdown timer
func (t *CountdownTimer) Start() error {
	if t.state == StateRunning {
		return nil // Already running
	}

	t.ctx, t.cancel = context.WithCancel(context.Background())
	if t.state == StatePaused {
		// Resume from paused state
		t.initialRemain = t.pausedAt
		t.startTime = time.Now()
		t.state = StateRunning
		go t.run(t.pausedAt)
		return nil
	}

	// Start fresh
	t.initialRemain = t.duration
	t.startTime = time.Now()
	t.pausedAt = 0
	t.state = StateRunning
	go t.run(t.duration)
	return nil
}

// Pause pauses the countdown timer
func (t *CountdownTimer) Pause() {
	if t.state != StateRunning {
		return
	}

	t.cancel()
	t.pausedAt = t.Remaining() // Use Remaining() to get accurate value
	t.state = StatePaused
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

// Stop stops the countdown timer and resets it
func (t *CountdownTimer) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.state = StateIdle
	t.pausedAt = 0
}

// Reset resets the timer to its initial duration
func (t *CountdownTimer) Reset() {
	t.Stop()
}

// Remaining returns the remaining time duration
func (t *CountdownTimer) Remaining() time.Duration {
	switch t.state {
	case StateRunning:
		elapsed := time.Since(t.startTime)
		remaining := t.initialRemain - elapsed
		if remaining < 0 {
			remaining = 0
		}
		return remaining
	case StatePaused:
		return t.pausedAt
	case StateCompleted:
		return 0
	default: // StateIdle
		return t.duration
	}
}

// State returns the current state of the timer
func (t *CountdownTimer) State() TimerState {
	return t.state
}

// Duration returns the total duration of the timer
func (t *CountdownTimer) Duration() time.Duration {
	return t.duration
}

// run is the internal goroutine that runs the countdown
func (t *CountdownTimer) run(startFrom time.Duration) {
	t.ticker = time.NewTicker(1 * time.Second)
	defer t.ticker.Stop()

	// Initial callback
	if t.onTick != nil {
		t.onTick(startFrom)
	}

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-t.ticker.C:
			remaining := t.Remaining()
			if remaining <= 0 {
				t.state = StateCompleted
				if t.onTick != nil {
					t.onTick(0)
				}
				if t.onComplete != nil {
					t.onComplete()
				}
				return
			}
			if t.onTick != nil {
				t.onTick(remaining)
			}
		}
	}
}

// WorkPeriodTimer is a specialized timer for work periods
type WorkPeriodTimer struct {
	*CountdownTimer
}

// NewWorkPeriodTimer creates a new work period timer
func NewWorkPeriodTimer(duration time.Duration) *WorkPeriodTimer {
	return &WorkPeriodTimer{
		CountdownTimer: NewCountdownTimer(duration),
	}
}

// RestPeriodTimer is a specialized timer for rest periods
type RestPeriodTimer struct {
	*CountdownTimer
}

// NewRestPeriodTimer creates a new rest period timer
func NewRestPeriodTimer(duration time.Duration) *RestPeriodTimer {
	return &RestPeriodTimer{
		CountdownTimer: NewCountdownTimer(duration),
	}
}
