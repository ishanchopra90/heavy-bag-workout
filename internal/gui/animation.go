package gui

import (
	"heavybagworkout/internal/models"
	"image"
	"time"
)

// AnimationState represents the current animation state
type AnimationState int

const (
	AnimationStateIdle AnimationState = iota
	AnimationStateJabLeft
	AnimationStateJabRight
	AnimationStateCrossLeft
	AnimationStateCrossRight
	AnimationStateLeadHookLeft
	AnimationStateLeadHookRight
	AnimationStateRearHookLeft
	AnimationStateRearHookRight
	AnimationStateLeadUppercutLeft
	AnimationStateLeadUppercutRight
	AnimationStateRearUppercutLeft
	AnimationStateRearUppercutRight
	AnimationStateSlipLeft
	AnimationStateSlipRight
	AnimationStateRollLeft
	AnimationStateRollRight
	AnimationStatePullBack
	AnimationStateDuck
)

// AnimationFrame represents a single frame in an animation
type AnimationFrame struct {
	Duration time.Duration // How long this frame should be displayed
	Image    image.Image   // The image/sprite for this frame (placeholder for now)
}

// Animation represents a complete animation sequence
type Animation struct {
	Frames []AnimationFrame
	Name   string
	Loop   bool // Whether the animation should loop
}

// CharacterSprite represents the Scrappy Doo character sprite system
type CharacterSprite struct {
	// Animation sets for different moves
	animations     map[AnimationState]*Animation
	currentState   AnimationState
	currentFrame   int
	frameStartTime time.Time
	moveStartTime  time.Time // When the current move animation started (for equal timing)
	stance         models.Stance
	assetLoader    *AssetLoader  // Asset loader for sprite images
	tempo          models.Tempo  // Task 54: Sync animations with combo timing (based on tempo setting)
	timePerMove    time.Duration // Time allocated per move in combo (for equal distribution)
}

// NewCharacterSprite creates a new Scrappy Doo character sprite
func NewCharacterSprite(stance models.Stance) *CharacterSprite {
	cs := &CharacterSprite{
		animations:    make(map[AnimationState]*Animation),
		currentState:  AnimationStateIdle,
		currentFrame:  0,
		stance:        stance,
		assetLoader:   NewAssetLoader("assets"), // Default asset directory
		tempo:         models.TempoSlow,         // Default tempo
		timePerMove:   0,                        // Will be set when combo starts
		moveStartTime: time.Now(),
	}

	// Initialize animations
	cs.initAnimations()

	return cs
}

// SetAssetLoader sets the asset loader for loading sprite images
func (cs *CharacterSprite) SetAssetLoader(loader *AssetLoader) {
	cs.assetLoader = loader
	// Reload animations with the new asset loader
	cs.initAnimations()
}

// initAnimations initializes all animations for Scrappy Doo
func (cs *CharacterSprite) initAnimations() {
	// Task 33: Create idle/ready pose animation
	cs.animations[AnimationStateIdle] = cs.createIdleAnimation()

	// Task 34-35: Jab animations
	cs.animations[AnimationStateJabLeft] = cs.createPunchAnimation("jab", AnimationStateJabLeft)
	cs.animations[AnimationStateJabRight] = cs.createPunchAnimation("jab", AnimationStateJabRight)

	// Task 36-37: Cross animations
	cs.animations[AnimationStateCrossLeft] = cs.createPunchAnimation("cross", AnimationStateCrossLeft)
	cs.animations[AnimationStateCrossRight] = cs.createPunchAnimation("cross", AnimationStateCrossRight)

	// Task 38-41: Hook animations
	cs.animations[AnimationStateLeadHookLeft] = cs.createPunchAnimation("left_hook", AnimationStateLeadHookLeft)
	cs.animations[AnimationStateLeadHookRight] = cs.createPunchAnimation("right_hook", AnimationStateLeadHookRight)
	cs.animations[AnimationStateRearHookLeft] = cs.createPunchAnimation("left_hook", AnimationStateRearHookLeft)
	cs.animations[AnimationStateRearHookRight] = cs.createPunchAnimation("right_hook", AnimationStateRearHookRight)

	// Task 42-45: Uppercut animations
	cs.animations[AnimationStateLeadUppercutLeft] = cs.createPunchAnimation("left_uppercut", AnimationStateLeadUppercutLeft)
	cs.animations[AnimationStateLeadUppercutRight] = cs.createPunchAnimation("right_uppercut", AnimationStateLeadUppercutRight)
	cs.animations[AnimationStateRearUppercutLeft] = cs.createPunchAnimation("left_uppercut", AnimationStateRearUppercutLeft)
	cs.animations[AnimationStateRearUppercutRight] = cs.createPunchAnimation("right_uppercut", AnimationStateRearUppercutRight)

	// Task 46-51: Defensive move animations
	cs.animations[AnimationStateSlipLeft] = cs.createDefensiveAnimation("left_slip", AnimationStateSlipLeft)
	cs.animations[AnimationStateSlipRight] = cs.createDefensiveAnimation("right_slip", AnimationStateSlipRight)
	cs.animations[AnimationStateRollLeft] = cs.createDefensiveAnimation("left_roll", AnimationStateRollLeft)
	cs.animations[AnimationStateRollRight] = cs.createDefensiveAnimation("right_roll", AnimationStateRollRight)
	cs.animations[AnimationStatePullBack] = cs.createDefensiveAnimation("pull_back", AnimationStatePullBack)
	cs.animations[AnimationStateDuck] = cs.createDefensiveAnimation("duck", AnimationStateDuck)
}

// createIdleAnimation creates the idle/ready pose animation (Task 33)
func (cs *CharacterSprite) createIdleAnimation() *Animation {
	// Load idle animation frames from assets
	var frames []AnimationFrame

	// If timePerMove is set (from combo sequence), use it for idle duration; otherwise use default
	var frameDuration time.Duration
	if cs.timePerMove > 0 {
		// Use the fixed duration per tempo (same as moves)
		frameDuration = cs.timePerMove
	} else {
		// Default duration when not in a combo sequence
		frameDuration = 500 * time.Millisecond
	}

	if cs.assetLoader != nil {
		images, err := cs.assetLoader.LoadIdleAnimation(cs.stance)
		if err == nil && len(images) > 0 {
			// Use loaded images with tempo-based duration
			for _, img := range images {
				frames = append(frames, AnimationFrame{
					Duration: frameDuration,
					Image:    img,
				})
			}
		}
	}

	// If no images loaded, create placeholder frames with tempo-based duration
	if len(frames) == 0 {
		frames = []AnimationFrame{
			{
				Duration: frameDuration,
				Image:    nil, // Placeholder - will be replaced with actual sprite
			},
			{
				Duration: frameDuration,
				Image:    nil, // Placeholder - will be replaced with actual sprite
			},
		}
	}

	return &Animation{
		Name:   "idle",
		Loop:   true,
		Frames: frames,
	}
}

// createPunchAnimation creates a punch animation for the given move name
// This is a generic function that works for all punch types (jab, cross, hooks, uppercuts)
// Task 54: Sync animations with combo timing (based on tempo setting)
func (cs *CharacterSprite) createPunchAnimation(moveName string, state AnimationState) *Animation {
	var frames []AnimationFrame

	// Calculate frame duration based on tempo or timePerMove.
	var frameDuration time.Duration
	if cs.timePerMove > 0 {
		// Use the allocated time per move (ensures equal distribution in combos)
		frameDuration = cs.timePerMove
	} else {
		// Fallback: use tempo-based calculation
		tempoDuration := cs.tempo.Duration()
		if cs.tempo == models.TempoSuperfast {
			tempoDuration = 1 * time.Second // Override to 1 second for superfast
		}
		frameDuration = tempoDuration / 2 // Complete punch in half the tempo interval
	}

	// Try to load sprite image based on stance
	if cs.assetLoader != nil {
		img, err := cs.assetLoader.LoadSpriteImage(cs.stance, moveName)
		if err == nil && img != nil {
			// Only one frame per punch for now
			frames = []AnimationFrame{
				{
					Duration: frameDuration,
					Image:    img,
				},
			}
		}
	}

	// If no image loaded, use placeholder frame with correct duration
	if len(frames) == 0 {
		frames = []AnimationFrame{
			{
				Duration: frameDuration,
				Image:    nil,
			},
		}
	}

	return &Animation{
		Name:   moveName,
		Loop:   false, // Punches are one-time actions
		Frames: frames,
	}
}

// createDefensiveAnimation creates a defensive move animation
// Task 54: Sync animations with combo timing (based on tempo setting)
func (cs *CharacterSprite) createDefensiveAnimation(moveName string, state AnimationState) *Animation {
	var frames []AnimationFrame

	// Calculate frame duration based on tempo or timePerMove.
	var frameDuration time.Duration
	if cs.timePerMove > 0 {
		// Use the allocated time per move (ensures equal distribution in combos)
		frameDuration = cs.timePerMove
	} else {
		// Fallback: use tempo-based calculation
		tempoDuration := cs.tempo.Duration()
		if cs.tempo == models.TempoSuperfast {
			tempoDuration = 1 * time.Second
		}
		frameDuration = tempoDuration / 2 // Complete defensive move in half the tempo interval
	}

	// Try to load sprite image based on stance
	if cs.assetLoader != nil {
		img, err := cs.assetLoader.LoadSpriteImage(cs.stance, moveName)
		if err == nil && img != nil {
			// Single frame for defensive move
			frames = []AnimationFrame{
				{
					Duration: frameDuration,
					Image:    img,
				},
			}
		}
	}

	// If no image loaded, use placeholder frame with correct duration
	if len(frames) == 0 {
		frames = []AnimationFrame{
			{
				Duration: frameDuration,
				Image:    nil,
			},
		}
	}

	return &Animation{
		Name:   moveName,
		Loop:   false, // Defensive moves are one-time actions
		Frames: frames,
	}
}

// SetStance updates the character's stance
func (cs *CharacterSprite) SetStance(stance models.Stance) {
	cs.stance = stance
}

// GetStance returns the current stance
func (cs *CharacterSprite) GetStance() models.Stance {
	return cs.stance
}

// SetAnimation sets the current animation state
// Task 55: Handle animation transitions between moves
func (cs *CharacterSprite) SetAnimation(state AnimationState) {
	if cs.currentState != state {
		// Task 55: Smooth transition - reset to first frame when changing animations
		cs.currentState = state
		cs.currentFrame = 0
		now := time.Now()
		cs.frameStartTime = now
		// Only set moveStartTime if it hasn't been explicitly set (zero value)
		// This allows external code to set moveStartTime before calling SetAnimation
		// for precise timing control in combo sequences
		if cs.moveStartTime.IsZero() {
			cs.moveStartTime = now // Track when this move started
		}
	}
}

// SetMoveStartTime sets the move start time explicitly (for precise timing control)
func (cs *CharacterSprite) SetMoveStartTime(startTime time.Time) {
	cs.moveStartTime = startTime
}

// SetFrameStartTime sets the frame start time explicitly (for precise timing control)
func (cs *CharacterSprite) SetFrameStartTime(startTime time.Time) {
	cs.frameStartTime = startTime
}

// SetTempo sets the tempo for animation timing
// Task 54: Sync animations with combo timing (based on tempo setting)
func (cs *CharacterSprite) SetTempo(tempo models.Tempo) {
	cs.tempo = tempo
	// Recreate animations with tempo-adjusted durations
	cs.initAnimations()
}

// SetTimePerMove sets the time allocated per move in a combo
// This ensures equal time distribution among moves
func (cs *CharacterSprite) SetTimePerMove(duration time.Duration) {
	cs.timePerMove = duration
}

// GetCurrentState returns the current animation state
func (cs *CharacterSprite) GetCurrentState() AnimationState {
	return cs.currentState
}

// Update updates the animation frame based on elapsed time
func (cs *CharacterSprite) Update(now time.Time) {
	anim, exists := cs.animations[cs.currentState]
	if !exists || len(anim.Frames) == 0 {
		return
	}

	// Check if we're already on the last frame of a non-looping animation
	// In this case, we should not advance frames - just stay on the last frame
	// The combo animation system in app.go will handle the transition to idle
	// based on elapsedSinceMoveStart, not frame duration
	if !anim.Loop && cs.currentFrame == len(anim.Frames)-1 {
		// Already on last frame - don't update frameStartTime or try to advance
		// This allows the frame to stay visible while the combo system checks
		// elapsedSinceMoveStart to determine when to transition to idle
		return
	}

	// Check if current frame duration has elapsed
	elapsed := now.Sub(cs.frameStartTime)
	currentFrame := anim.Frames[cs.currentFrame]

	if elapsed >= currentFrame.Duration {
		// Move to next frame
		cs.currentFrame++

		// Check if animation is complete
		if cs.currentFrame >= len(anim.Frames) {
			if anim.Loop {
				// Loop back to first frame
				cs.currentFrame = 0
				cs.frameStartTime = now
			} else {
				// Animation complete - stay on last frame
				cs.currentFrame = len(anim.Frames) - 1
				// Set frameStartTime to now when we first reach the last frame
				// This ensures accurate timing - the frame duration is timePerMove,
				// and we want to track when we entered the last frame
				cs.frameStartTime = now
				return
			}
		} else {
			// Normal frame advance - update frameStartTime
			cs.frameStartTime = now
		}
	}
}

// GetCurrentFrame returns the current animation frame
func (cs *CharacterSprite) GetCurrentFrame() *AnimationFrame {
	anim, exists := cs.animations[cs.currentState]
	if !exists || len(anim.Frames) == 0 {
		return nil
	}

	if cs.currentFrame >= len(anim.Frames) {
		cs.currentFrame = len(anim.Frames) - 1
	}

	return &anim.Frames[cs.currentFrame]
}

// GetCurrentFrameIndex returns the current frame index
func (cs *CharacterSprite) GetCurrentFrameIndex() int {
	return cs.currentFrame
}

// GetAnimation returns the animation for a given state
func (cs *CharacterSprite) GetAnimation(state AnimationState) *Animation {
	return cs.animations[state]
}

// Reset resets the animation to the beginning
func (cs *CharacterSprite) Reset() {
	cs.currentFrame = 0
	cs.frameStartTime = time.Now()
}

// GetFrameStartTime returns the time when the current frame started
func (cs *CharacterSprite) GetFrameStartTime() time.Time {
	return cs.frameStartTime
}

// GetMoveStartTime returns the time when the current move animation started
func (cs *CharacterSprite) GetMoveStartTime() time.Time {
	return cs.moveStartTime
}

// GetTimePerMove returns the time allocated per move in combo
func (cs *CharacterSprite) GetTimePerMove() time.Duration {
	return cs.timePerMove
}

// RecreateAnimations recreates all animations with current settings (tempo, timePerMove)
// This is needed when timePerMove changes to scale animation durations
func (cs *CharacterSprite) RecreateAnimations() {
	cs.initAnimations()
}
