# Animation System Architecture

This document describes the animation system used in the GUI to display character sprites performing boxing moves.

## Overview

The animation system provides visual feedback by displaying animated character sprites (Scrappy Doo) performing the combos in real-time. The system is designed for deterministic, frame-accurate timing synchronized with audio beeps.

### Animation System Architecture

```mermaid
graph TB
    App[App<br/>GUI Application]
    AnimationSequence[Animation Sequence<br/>Timer Orchestrator]
    CharacterSprite[CharacterSprite<br/>Animation Controller]
    AssetLoader[Asset Loader<br/>Image Loading]
    WorkoutTimer[WorkoutTimer<br/>Period Management]
    AudioHandler[Audio Handler<br/>Beep Synchronization]
    
    App -->|manages| AnimationSequence
    App -->|uses| CharacterSprite
    App -->|receives callbacks| WorkoutTimer
    App -->|triggers| AudioHandler
    
    AnimationSequence -->|controls| CharacterSprite
    AnimationSequence -->|uses timers| Timer[time.AfterFunc<br/>Move/Idle Timers]
    
    CharacterSprite -->|loads from| AssetLoader
    AssetLoader -->|reads| Assets[assets/orthodox/<br/>assets/southpaw/]
    
    WorkoutTimer -->|calls| App
    AudioHandler -->|synchronizes| AnimationSequence
    
    style App fill:#4fc3f7
    style AnimationSequence fill:#81c784
    style CharacterSprite fill:#ffb74d
    style AssetLoader fill:#ba68c8
    style WorkoutTimer fill:#64b5f6
    style AudioHandler fill:#ef5350
```

## Core Components

### 1. CharacterSprite (`internal/gui/animation.go`)

The main animation controller that manages character sprite states and frame rendering.

**Key Features:**
- Animation state machine (idle, punches, defensive moves)
- Frame-based animation system
- Stance-aware animations (orthodox vs southpaw)
- Tempo-based timing support
- Asset loading system

**Animation States:**
- `AnimationStateIdle`: Ready/idle pose
- `AnimationStateJabLeft/Right`: Jab animations
- `AnimationStateCrossLeft/Right`: Cross animations
- `AnimationStateLeadHookLeft/Right`: Lead hook animations
- `AnimationStateRearHookLeft/Right`: Rear hook animations
- `AnimationStateLeadUppercutLeft/Right`: Lead uppercut animations
- `AnimationStateRearUppercutLeft/Right`: Rear uppercut animations
- `AnimationStateSlipLeft/Right`: Slip defensive moves
- `AnimationStateRollLeft/Right`: Roll defensive moves
- `AnimationStatePullBack`: Pull back defensive move
- `AnimationStateDuck`: Duck defensive move

### 2. Animation Sequence System (`internal/gui/animation_sequence.go`)

A timer-based system that orchestrates the sequence of animations for a combo.

**Key Features:**
- Deterministic timing using `time.AfterFunc()`
- Single goroutine execution (no race conditions)
- Precise move timing (400ms per move)
- Idle animation between combos
- Work period duration tracking

**Flow:**
1. Beep plays → `startAnimationSequence()` called
2. Check if work period has elapsed
3. Start first move animation with timer
4. Timer fires → move to next move
5. After all moves → start idle animation
6. Idle timer fires → start next combo sequence

### 3. Asset Loader (`internal/gui/assets.go`)

Loads sprite images from the file system.

**Features:**
- Stance-specific asset directories (`assets/orthodox/`, `assets/southpaw/`)
- Format support: JPG (preferred), PNG (fallback)
- Error handling with graceful fallbacks
- Idle animation fallback to jab pose

**Asset Structure:**

```mermaid
graph LR
    Assets[assets/] --> Orthodox[orthodox/]
    Assets --> Southpaw[southpaw/]
    
    Orthodox --> O1[idle.jpg]
    Orthodox --> O2[jab.jpg]
    Orthodox --> O3[cross.jpg]
    Orthodox --> O4[left_hook.jpg]
    Orthodox --> O5[right_hook.jpg]
    Orthodox --> O6[left_uppercut.jpg]
    Orthodox --> O7[right_uppercut.jpg]
    Orthodox --> O8[left_slip.jpg]
    Orthodox --> O9[right_slip.jpg]
    Orthodox --> O10[left_roll.jpg]
    Orthodox --> O11[right_roll.jpg]
    Orthodox --> O12[pull_back.jpg]
    Orthodox --> O13[duck.jpg]
    
    Southpaw --> S1[same structure]
    
    style Assets fill:#4fc3f7
    style Orthodox fill:#81c784
    style Southpaw fill:#ffb74d
```

## Animation Timing

### Move Timing

Each move in a combo gets exactly **400ms** (`timePerMove`), regardless of tempo:
- This ensures consistent animation speed
- All moves get equal time
- Tempo affects beep intervals, not move speed

### Tempo Intervals

Tempo determines the interval between beeps (combo reminders):
- **Slow**: 5 seconds
- **Medium**: 4 seconds
- **Fast**: 3 seconds
- **Superfast**: 1 second

### Idle Animation

After a combo completes, idle animation plays for the remainder until the next beep:
- `idleDuration = tempoInterval - (numMoves * timePerMove)`
- If combo takes longer than tempo interval, idle duration is 0
- Idle animation is non-looping during work periods

## Animation Flow

### Work Period Flow

```mermaid
flowchart TD
    Start[Work Period Starts] --> OnStart[OnPeriodStart called]
    OnStart --> Store[Store workoutStartTime<br/>and workoutPeriodDuration]
    Store --> IdleInit[Set idle animation]
    IdleInit --> Beep[Beep plays]
    Beep --> StartSeq[startAnimationSequence]
    StartSeq --> CheckPeriod{Has work<br/>period elapsed?}
    CheckPeriod -->|Yes| Rest[Transition to rest period]
    CheckPeriod -->|No| MoveLoop[For each move in combo]
    MoveLoop --> SetAnim[Set move animation]
    SetAnim --> ShowGo[Show 'go!' indicator 500ms]
    ShowGo --> StartTimer[Start timer 400ms]
    StartTimer --> TimerFire[Timer fires]
    TimerFire --> NextMove{More moves?}
    NextMove -->|Yes| MoveLoop
    NextMove -->|No| AllComplete[All moves complete]
    AllComplete --> IdleAnim[Start idle animation]
    IdleAnim --> IdleTimer[Idle timer fires]
    IdleTimer --> CheckPeriod
    
    style Start fill:#e1f5ff
    style Beep fill:#fff9c4
    style MoveLoop fill:#b3e5fc
    style Rest fill:#ffccbc
```

### Rest Period Flow

```mermaid
flowchart TD
    Start[Rest Period Starts] --> OnStart[OnPeriodStart called]
    OnStart --> Stop[Stop work period animations]
    Stop --> StartRest[startRestPeriodAnimation]
    StartRest --> SetIdle[Set idle animation<br/>non-looping]
    SetIdle --> StartTimer[Start rest period timer]
    StartTimer --> Wait[Wait for timer]
    Wait --> TimerFires[Timer fires when<br/>rest period completes]
    TimerFires --> Next[WorkoutTimer transitions<br/>to next work period]
    
    style Start fill:#ffccbc
    style SetIdle fill:#fff9c4
    style Next fill:#c8e6c9
```

## Synchronization

### Audio Synchronization

Animations are synchronized with audio beeps:
- Beep plays at the start of each combo sequence
- Animation sequence starts immediately when beep plays
- This ensures visual and audio cues are aligned

### Timer Synchronization

The animation system uses multiple timer layers:
1. **WorkoutTimer**: Manages overall workout flow (rounds, periods)
2. **Animation Timers**: Manage individual move and idle durations
3. **Period Check**: At start of each combo, check if work period has elapsed

### Deterministic Execution

The system ensures deterministic behavior:
- Single goroutine for animation sequence (no race conditions)
- Precise timer-based timing (no frame skipping)
- State checks at sequence boundaries
- Clean stop/start transitions

## Stance Support

Animations adapt to the selected stance:

**Orthodox (Right-handed):**
- Jab → Left hand animation
- Cross → Right hand animation
- Lead Hook → Left hook
- Rear Hook → Right hook
- Lead Uppercut → Left uppercut
- Rear Uppercut → Right uppercut

**Southpaw (Left-handed):**
- Jab → Right hand animation
- Cross → Left hand animation
- Lead Hook → Right hook
- Rear Hook → Left hook
- Lead Uppercut → Right uppercut
- Rear Uppercut → Left uppercut

Defensive moves are the same for both stances (no stance-specific variations).

## Error Handling

The animation system includes robust error handling:

1. **Asset Loading Failures**: Falls back to placeholder frames or jab pose
2. **Timer Cancellation**: Clean shutdown when workout stops/pauses
3. **State Validation**: Checks period state before starting animations
4. **Bounds Checking**: Validates combo moves before accessing

## Performance Considerations

1. **Frame Rendering**: Animations use single-frame sprites (not multi-frame sequences)
2. **Timer Efficiency**: Uses `time.AfterFunc()` for efficient timer management
3. **Asset Caching**: Assets are loaded once and reused
4. **Goroutine Management**: Single goroutine per animation sequence prevents overhead

## Future Enhancements

Potential improvements to the animation system:

1. **Multi-Frame Animations**: Support for animated sprite sequences
2. **Transition Effects**: Smooth transitions between moves
3. **Impact Effects**: Visual effects when punches "land"
4. **Performance Optimization**: Frame rate optimization for smoother animations
5. **Animation Variants**: Multiple animation styles for the same move

