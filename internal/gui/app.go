package gui

import (
	"fmt"
	"heavybagworkout/internal/config"
	"heavybagworkout/internal/generator"
	"heavybagworkout/internal/models"
	"heavybagworkout/internal/timer"
	"heavybagworkout/internal/types"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// App represents the main GUI application
type App struct {
	theme *material.Theme

	// Form fields for workout parameters
	workDurationEditor widget.Editor
	restDurationEditor widget.Editor
	totalRoundsEditor  widget.Editor

	// Pattern dropdown
	patternDropdownOpen bool
	patternButton       widget.Clickable
	patternOptions      []widget.Clickable
	selectedPattern     models.WorkoutPatternType

	// Min/Max moves fields
	minMovesEditor widget.Editor
	maxMovesEditor widget.Editor

	// Defensive moves checkbox
	includeDefensive widget.Bool

	// Stance dropdown
	stanceDropdownOpen bool
	stanceButton       widget.Clickable
	stanceOptions      []widget.Clickable
	selectedStance     models.Stance

	// Tempo dropdown
	tempoDropdownOpen bool
	tempoButton       widget.Clickable
	tempoOptions      []widget.Clickable
	selectedTempo     models.Tempo

	// LLM generation checkbox
	useLLM widget.Bool

	// OpenAI API key field
	openAIAPIKeyEditor widget.Editor

	// Preset dropdown
	presetDropdownOpen bool
	presetButton       widget.Clickable
	presetOptions      []widget.Clickable
	selectedPreset     *models.WorkoutPreset // nil means "Custom" / no preset

	// Validation errors (stored as strings for display)
	validationErrors map[string]string

	// Action buttons
	startWorkoutButton   widget.Clickable
	loadConfigButton     widget.Clickable
	saveConfigButton     widget.Clickable
	configFilePathEditor widget.Editor

	// Status message (for showing success/error messages)
	statusMessage string
	statusError   bool

	// Workout display state
	showWorkoutDisplay bool // true = show workout display, false = show form
	showCompletion     bool // true = show completion screen
	showWorkoutPreview bool // true = show workout preview/confirmation screen

	// Workout progress state (will be updated by timer callbacks)
	currentRound int // Current round number (1-indexed, 0 if not started)
	totalRounds  int // Total number of rounds in the workout

	// Timer state (will be updated by timer callbacks)
	currentPeriod types.PeriodType // Current period (Work/Rest)
	remainingTime time.Duration    // Remaining time in current period

	// Current combo state (will be updated by timer callbacks)
	currentCombo models.Combo // Current combo for the active round
	showGo       bool         // Show "go!" indicator when combo should be performed (at tempo intervals during work period)

	// Timer-based animation sequence state
	animationTimer        *time.Timer   // Current animation timer (for move or idle)
	animationTimerStop    chan struct{} // Channel to stop the animation sequence goroutine
	workoutStartTime      time.Time     // When the current workout period started (for period duration check)
	workoutPeriodDuration time.Duration // Duration of the current workout period
	currentMoveIndex      int           // Current move index in the combo (0 = first move)

	// Workout control state
	isPaused       bool             // Whether the workout is paused
	pauseResumeBtn widget.Clickable // Pause/Resume button
	stopBtn        widget.Clickable // Stop/Quit button

	// Completion screen controls
	completionDoneBtn widget.Clickable // Button to return to form from completion screen

	// Workout preview controls
	confirmWorkoutBtn widget.Clickable // Button to confirm and start workout
	backToFormBtn     widget.Clickable // Button to go back to form from preview

	// Generated workout (stored after generation, used by timer)
	workout models.Workout

	// Workout timer (Task 58)
	workoutTimer *timer.WorkoutTimer
	audioHandler timer.AudioCueHandler // Audio handler for tempo-based beeps

	// Window reference for invalidating frames (needed for timer updates)
	window interface {
		Invalidate()
	}

	// List widget for scrolling
	list widget.List

	// Animation system (Tasks 32-34)
	characterSprite   *CharacterSprite
	characterRenderer *CharacterRenderer
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	app := &App{
		theme:           material.NewTheme(),
		selectedPattern: models.PatternLinear,
		selectedStance:  models.Orthodox,
		selectedTempo:   models.TempoSlow,
		currentPeriod:   types.PeriodWork,
	}

	// Initialize editors with single-line mode
	app.workDurationEditor.SingleLine = true
	app.workDurationEditor.Submit = true
	app.restDurationEditor.SingleLine = true
	app.restDurationEditor.Submit = true
	app.totalRoundsEditor.SingleLine = true
	app.totalRoundsEditor.Submit = true
	app.minMovesEditor.SingleLine = true
	app.minMovesEditor.Submit = true
	app.maxMovesEditor.SingleLine = true
	app.maxMovesEditor.Submit = true
	app.openAIAPIKeyEditor.SingleLine = true
	app.openAIAPIKeyEditor.Submit = true
	app.openAIAPIKeyEditor.Mask = '*' // Mask the API key input
	app.configFilePathEditor.SingleLine = true
	app.configFilePathEditor.Submit = true

	// Set default values
	app.workDurationEditor.SetText("20")
	app.restDurationEditor.SetText("10")
	app.totalRoundsEditor.SetText("10")
	app.minMovesEditor.SetText("3")
	app.maxMovesEditor.SetText("5")

	// Initialize pattern options clickables
	app.patternOptions = make([]widget.Clickable, 4)

	// Initialize stance options clickables
	app.stanceOptions = make([]widget.Clickable, 2)

	// Initialize tempo options clickables
	app.tempoOptions = make([]widget.Clickable, 4)

	// Initialize preset options clickables (3 presets + 1 for "Custom")
	app.presetOptions = make([]widget.Clickable, 4)

	// Initialize validation errors map
	app.validationErrors = make(map[string]string)

	// Initialize animation system (Tasks 32-34)
	app.characterSprite = NewCharacterSprite(app.selectedStance)
	app.characterRenderer = NewCharacterRenderer(app.characterSprite)

	return app
}

// SetWindow sets the window reference for invalidating frames
func (a *App) SetWindow(window interface{ Invalidate() }) {
	a.window = window
}

// Layout handles the main application layout
func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Note: We don't need to invalidate here because:
	// 1. Layout is called on every frame event automatically
	// 2. Timer callbacks call window.Invalidate() when state changes
	// This ensures the display updates smoothly without excessive invalidations

	// Choose layout based on current state
	if a.showCompletion {
		return a.layoutCompletionScreen(gtx)
	}
	if a.showWorkoutPreview {
		return a.layoutWorkoutPreview(gtx)
	}
	if a.showWorkoutDisplay {
		return a.layoutWorkoutDisplay(gtx)
	}
	return a.layoutForm(gtx)
}

// IsWorkoutInProgress returns true if a workout is currently in progress
// Used by the window close handler to determine if we should prompt for confirmation
func (a *App) IsWorkoutInProgress() bool {
	return a.workoutTimer != nil && a.showWorkoutDisplay && !a.showCompletion && a.currentRound > 0
}

// layoutForm handles the workout parameter form layout
func (a *App) layoutForm(gtx layout.Context) layout.Dimensions {
	// Add padding around the form
	inset := layout.Inset{
		Top:    unit.Dp(20),
		Bottom: unit.Dp(20),
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Use List widget for scrollable content
		a.list.Axis = layout.Vertical
		return material.List(a.theme, &a.list).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
			// Create a vertical list layout for the form
			return layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceStart,
				Alignment: layout.Start,
			}.Layout(gtx,
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(a.theme, "Puppy Power - Heavy Bag Workout")
					title.Alignment = text.Start
					return title.Layout(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Preset dropdown
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutPresetDropdown(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Work Duration field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutFormFieldWithValidation(gtx, "Work Duration (seconds)", &a.workDurationEditor, "workDuration")
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Rest Duration field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutFormFieldWithValidation(gtx, "Rest Duration (seconds)", &a.restDurationEditor, "restDuration")
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Total Rounds field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutFormFieldWithValidation(gtx, "Total Rounds", &a.totalRoundsEditor, "totalRounds")
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Pattern dropdown
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutPatternDropdown(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Min Moves field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutFormFieldWithValidation(gtx, "Minimum Moves per Combo", &a.minMovesEditor, "minMoves")
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Max Moves field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutFormFieldWithValidation(gtx, "Maximum Moves per Combo", &a.maxMovesEditor, "maxMoves")
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Include Defensive Moves checkbox
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutCheckbox(gtx, "Include defensive moves?", &a.includeDefensive, "includeDefensive")
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Stance dropdown
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutStanceDropdown(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Tempo dropdown
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutTempoDropdown(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Use LLM generation checkbox
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutCheckbox(gtx, "Use LLM generation", &a.useLLM, "useLLM")
				}),

				// Spacing and OpenAI API Key field (shown when LLM is enabled)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !a.useLLM.Value {
						return layout.Dimensions{}
					}
					return layout.Flex{
						Axis:      layout.Vertical,
						Spacing:   layout.SpaceStart,
						Alignment: layout.Start,
					}.Layout(gtx,
						// Spacing
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
						}),
						// API Key field
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutFormFieldWithValidation(gtx, "OpenAI API Key (optional)", &a.openAIAPIKeyEditor, "openAIAPIKey")
						}),
					)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Config file path field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutFormField(gtx, "Config File Path (optional)", &a.configFilePathEditor)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
				}),

				// Action buttons row
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Spacing:   layout.SpaceBetween,
						Alignment: layout.Middle,
					}.Layout(gtx,
						// Load Config button
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.loadConfigButton.Clicked(gtx) {
								a.handleLoadConfig()
							}
							btn := material.Button(a.theme, &a.loadConfigButton, "Load Config")
							return btn.Layout(gtx)
						}),

						// Save Config button
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.saveConfigButton.Clicked(gtx) {
								a.handleSaveConfig()
							}
							btn := material.Button(a.theme, &a.saveConfigButton, "Save Config")
							return btn.Layout(gtx)
						}),
					)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Start Workout button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.startWorkoutButton.Clicked(gtx) {
						a.handleStartWorkout()
					}
					btn := material.Button(a.theme, &a.startWorkoutButton, "Start Workout")
					// Make the button more prominent
					btn.Background = color.NRGBA{R: 0, G: 150, B: 0, A: 255} // Green background
					return btn.Layout(gtx)
				}),

				// Status message
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.statusMessage != "" {
						return layout.Inset{
							Top: unit.Dp(10),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							statusColor := color.NRGBA{R: 0, G: 150, B: 0, A: 255} // Green for success
							if a.statusError {
								statusColor = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Red for error
							}
							lbl := material.Body2(a.theme, a.statusMessage)
							lbl.Color = statusColor
							return lbl.Layout(gtx)
						})
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

// layoutWorkoutDisplay handles the workout display screen layout
// This layout structure is ready to be populated with workout data in subsequent tasks
// Task 56: Add visual feedback for work vs rest periods (different background/colors)
func (a *App) layoutWorkoutDisplay(gtx layout.Context) layout.Dimensions {
	// Task 56: Set background color based on period type
	var bgColor color.NRGBA
	if a.currentPeriod == types.PeriodWork {
		// Work period: Light red/orange background to indicate activity
		bgColor = color.NRGBA{R: 255, G: 245, B: 238, A: 255} // Light orange/red tint
	} else {
		// Rest period: Light blue background to indicate rest
		bgColor = color.NRGBA{R: 235, G: 245, B: 255, A: 255} // Light blue tint
	}

	// Fill entire area with background color
	paint.FillShape(gtx.Ops, bgColor, clip.Rect{Max: gtx.Constraints.Max}.Op())

	// Add padding around the workout display
	inset := layout.Inset{
		Top:    unit.Dp(20),
		Bottom: unit.Dp(20),
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Use List widget for scrollable content if needed
		a.list.Axis = layout.Vertical
		return material.List(a.theme, &a.list).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
			// Main vertical layout for workout display
			return layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceStart,
				Alignment: layout.Middle, // Center content horizontally
			}.Layout(gtx,
				// Title section
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(a.theme, "Workout in Progress")
					title.Alignment = text.Middle
					return title.Layout(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
				}),

				// Round number section (Task 23 - will be implemented)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutRoundNumber(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Progress bar section (Task 24 - will be implemented)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutWorkoutProgress(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
				}),

				// Countdown timer section (Task 25 - will be implemented)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutCountdownTimer(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Period indicator section
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutPeriodIndicator(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
				}),

				// Character animation section (Tasks 32-34, 52 - heavy bag is part of character animation)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutCharacterAnimation(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Combo moves section (Task 27 - will be implemented)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutComboMoves(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(40)}.Layout(gtx)
				}),

				// Control buttons section (Tasks 28-29 - will be implemented)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutWorkoutControls(gtx)
				}),
			)
		})
	})
}

// layoutCompletionScreen displays the workout completion screen (Task 31)
func (a *App) layoutCompletionScreen(gtx layout.Context) layout.Dimensions {
	// Handle completion screen button click
	if a.completionDoneBtn.Clicked(gtx) {
		a.resetWorkoutState()
	}

	inset := layout.Inset{
		Top:    unit.Dp(40),
		Bottom: unit.Dp(40),
		Left:   unit.Dp(40),
		Right:  unit.Dp(40),
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Spacing:   layout.SpaceStart,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Completion title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H3(a.theme, "Workout Complete!")
				title.Alignment = text.Middle
				title.Color = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green color
				return title.Layout(gtx)
			}),

			// Spacing
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
			}),

			// Completion message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var message string
				if a.totalRounds > 0 {
					message = fmt.Sprintf("Congratulations! You completed all %d rounds.", a.totalRounds)
				} else {
					message = "Congratulations! You completed the workout."
				}
				label := material.Body1(a.theme, message)
				label.Alignment = text.Middle
				label.Color = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
				return label.Layout(gtx)
			}),

			// Spacing
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Spacer{Height: unit.Dp(50)}.Layout(gtx)
			}),

			// Done button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(a.theme, &a.completionDoneBtn, "Return to Form")
				btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}     // White text
				btn.Background = color.NRGBA{R: 33, G: 150, B: 243, A: 255} // Blue
				btn.CornerRadius = unit.Dp(4)
				inset := layout.Inset{
					Left:  unit.Dp(20),
					Right: unit.Dp(20),
				}
				return inset.Layout(gtx, btn.Layout)
			}),
		)
	})
}

// layoutWorkoutPreview displays the generated workout for confirmation before starting
func (a *App) layoutWorkoutPreview(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if a.confirmWorkoutBtn.Clicked(gtx) {
		a.handleConfirmWorkout()
	}
	if a.backToFormBtn.Clicked(gtx) {
		a.handleBackToForm()
	}

	inset := layout.Inset{
		Top:    unit.Dp(20),
		Bottom: unit.Dp(20),
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Use List widget for scrollable content
		a.list.Axis = layout.Vertical
		return material.List(a.theme, &a.list).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceStart,
				Alignment: layout.Middle,
			}.Layout(gtx,
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(a.theme, "Workout Preview")
					title.Alignment = text.Middle
					return title.Layout(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Workout summary
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutWorkoutSummary(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
				}),

				// Rounds list
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutWorkoutRoundsList(gtx)
				}),

				// Spacing
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
				}),

				// Action buttons
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Spacing:   layout.SpaceAround,
						Alignment: layout.Middle,
					}.Layout(gtx,
						// Back button
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							inset := layout.Inset{
								Left:  unit.Dp(10),
								Right: unit.Dp(10),
							}
							return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(a.theme, &a.backToFormBtn, "Back to Form")
								btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
								btn.Background = color.NRGBA{R: 158, G: 158, B: 158, A: 255} // Gray
								btn.CornerRadius = unit.Dp(4)
								return btn.Layout(gtx)
							})
						}),

						// Spacing
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Width: unit.Dp(40)}.Layout(gtx)
						}),

						// Confirm button
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							inset := layout.Inset{
								Left:  unit.Dp(10),
								Right: unit.Dp(10),
							}
							return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(a.theme, &a.confirmWorkoutBtn, "Confirm and Start")
								btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
								btn.Background = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green
								btn.CornerRadius = unit.Dp(4)
								return btn.Layout(gtx)
							})
						}),
					)
				}),
			)
		})
	})
}

// layoutWorkoutSummary displays workout summary information
func (a *App) layoutWorkoutSummary(gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{
		Left:  unit.Dp(40),
		Right: unit.Dp(40),
	}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var summaryText string
		if len(a.workout.Rounds) > 0 {
			firstRound := a.workout.Rounds[0]
			workSec := int(firstRound.WorkDuration.Seconds())
			restSec := int(firstRound.RestDuration.Seconds())
			totalDuration := a.workout.TotalDuration()
			minutes := int(totalDuration.Minutes())
			seconds := int(totalDuration.Seconds()) % 60

			summaryText = fmt.Sprintf("Total Rounds: %d | Work: %ds | Rest: %ds | Total Time: %dm %ds",
				len(a.workout.Rounds), workSec, restSec, minutes, seconds)
		} else {
			summaryText = "No workout data"
		}

		label := material.Body1(a.theme, summaryText)
		label.Alignment = text.Middle
		label.Color = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
		return label.Layout(gtx)
	})
}

// layoutWorkoutRoundsList displays all rounds with their combos in a scrollable list
func (a *App) layoutWorkoutRoundsList(gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{
		Left:  unit.Dp(40),
		Right: unit.Dp(40),
		Top:   unit.Dp(10),
	}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Create a list of rounds
		roundsList := widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		}

		return material.List(a.theme, &roundsList).Layout(gtx, len(a.workout.Rounds), func(gtx layout.Context, index int) layout.Dimensions {
			if index >= len(a.workout.Rounds) {
				return layout.Dimensions{}
			}
			round := a.workout.Rounds[index]

			return layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceStart,
				Alignment: layout.Start,
			}.Layout(gtx,
				// Round header
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					roundHeader := material.H6(a.theme, fmt.Sprintf("Round %d:", round.RoundNumber))
					roundHeader.Color = color.NRGBA{R: 33, G: 150, B: 243, A: 255} // Blue
					return roundHeader.Layout(gtx)
				}),

				// Combo moves (with stance-specific names)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					comboText := a.formatComboWithStance(round.Combo, a.selectedStance)
					if comboText == "" {
						comboText = "No moves"
					}
					comboLabel := material.Body2(a.theme, "  "+comboText)
					comboLabel.Color = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
					inset := layout.Inset{
						Left: unit.Dp(20),
						Top:  unit.Dp(4),
					}
					return inset.Layout(gtx, comboLabel.Layout)
				}),

				// Spacing between rounds
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(12)}.Layout(gtx)
				}),
			)
		})
	})
}

// handleConfirmWorkout confirms the workout and starts it
func (a *App) handleConfirmWorkout() {
	// PRIORITY: Switch to workout display FIRST to ensure UI appears before audio starts
	a.showWorkoutPreview = false
	a.showWorkoutDisplay = true
	a.statusMessage = ""

	// Invalidate window to trigger immediate redraw
	if a.window != nil {
		a.window.Invalidate()
	}

	// Create workout timer
	a.workoutTimer = timer.NewWorkoutTimer(a.workout)
	a.workoutTimer.SetStance(a.selectedStance)

	// Update character sprite stance and tempo (Tasks 32-34, 54)
	if a.characterSprite != nil {
		a.characterSprite.SetStance(a.selectedStance)
		a.characterSprite.SetTempo(a.selectedTempo)
	}

	// Set display handler (App implements TimerDisplayHandler)
	a.workoutTimer.SetDisplayHandler(a)

	// Set audio handler (default audio cues)
	audioHandler := timer.NewDefaultAudioCueHandler(true)
	a.audioHandler = audioHandler // Store for tempo ticker
	a.workoutTimer.SetAudioHandler(audioHandler)

	// Set workout completion callback
	a.workoutTimer.OnWorkoutComplete(func() {
		a.handleWorkoutComplete()
	})

	// Start the timer in a goroutine with a small delay to allow UI to render first
	go func() {
		// Small delay to ensure window has time to redraw
		time.Sleep(50 * time.Millisecond)

		// Start the timer (this will trigger audio callbacks)
		if err := a.workoutTimer.Start(); err != nil {
			// Handle error - we need to update UI from main thread
			// For now, we'll just log it since we're in a goroutine
			// In a production app, you'd want to use a channel or similar to communicate back
			return
		}
	}()
}

// handleBackToForm goes back to the form screen
func (a *App) handleBackToForm() {
	a.showWorkoutPreview = false
	a.showWorkoutDisplay = false
	// Keep the generated workout in case user wants to regenerate with different params
	// but allow them to go back and adjust settings
}

// formatComboWithStance formats a combo with stance-specific punch names
func (a *App) formatComboWithStance(combo models.Combo, stance models.Stance) string {
	if combo.IsEmpty() {
		return ""
	}

	result := ""
	for i, move := range combo.Moves {
		if i > 0 {
			result += ", "
		}

		if move.IsPunch() && move.Punch != nil {
			// Use stance-specific name for punches
			punchName := move.Punch.NameForStance(stance)
			// Capitalize first letter
			if len(punchName) > 0 {
				punchName = strings.ToUpper(string(punchName[0])) + punchName[1:]
			}
			result += punchName
		} else if move.IsDefensive() && move.Defensive != nil {
			// For defensive moves, show the full name
			result += move.String()
		}
	}

	return result
}

// layoutRoundNumber displays the current round number prominently (Task 23)
func (a *App) layoutRoundNumber(gtx layout.Context) layout.Dimensions {
	var roundText string
	if a.totalRounds > 0 && a.currentRound > 0 {
		// Display "Round X of Y" format
		roundText = fmt.Sprintf("Round %d of %d", a.currentRound, a.totalRounds)
	} else if a.totalRounds > 0 {
		// Workout ready but not started yet
		roundText = fmt.Sprintf("Ready - %d rounds", a.totalRounds)
	} else {
		// No workout data yet
		roundText = "Round -"
	}

	// Use H2 for prominent display
	label := material.H2(a.theme, roundText)
	label.Alignment = text.Middle
	label.Color = color.NRGBA{R: 50, G: 50, B: 50, A: 255} // Dark gray for good visibility
	return label.Layout(gtx)
}

// layoutWorkoutProgress displays workout progress with progress bar (Task 24)
func (a *App) layoutWorkoutProgress(gtx layout.Context) layout.Dimensions {
	// Calculate rounds completed (current round - 1, since we're currently in a round)
	roundsCompleted := a.currentRound - 1
	if roundsCompleted < 0 {
		roundsCompleted = 0
	}
	if a.currentRound > a.totalRounds && a.totalRounds > 0 {
		roundsCompleted = a.totalRounds
	}

	// Calculate progress percentage
	var progress float64
	if a.totalRounds > 0 {
		progress = float64(roundsCompleted) / float64(a.totalRounds)
		if progress > 1.0 {
			progress = 1.0
		}
		if progress < 0.0 {
			progress = 0.0
		}
	}

	// Layout progress bar and text in a column
	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Progress bar container with padding
			inset := layout.Inset{
				Left:  unit.Dp(40),
				Right: unit.Dp(40),
				Top:   unit.Dp(10),
			}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Progress bar height
				barHeight := gtx.Dp(unit.Dp(20))

				// Background (unfilled portion)
				backgroundRect := image.Rectangle{
					Max: image.Point{
						X: gtx.Constraints.Max.X,
						Y: barHeight,
					},
				}
				paint.FillShape(gtx.Ops, color.NRGBA{R: 220, G: 220, B: 220, A: 255}, clip.Rect(backgroundRect).Op())

				// Filled portion based on progress
				if progress > 0 {
					filledWidth := int(float64(gtx.Constraints.Max.X) * progress)
					if filledWidth > gtx.Constraints.Max.X {
						filledWidth = gtx.Constraints.Max.X
					}
					filledRect := image.Rectangle{
						Max: image.Point{
							X: filledWidth,
							Y: barHeight,
						},
					}
					// Use a nice blue color for the progress bar
					paint.FillShape(gtx.Ops, color.NRGBA{R: 33, G: 150, B: 243, A: 255}, clip.Rect(filledRect).Op())
				}

				return layout.Dimensions{
					Size: image.Point{
						X: gtx.Constraints.Max.X,
						Y: barHeight,
					},
				}
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Progress text below the bar
			var progressText string
			if a.totalRounds > 0 {
				progressPercent := int(progress * 100)
				progressText = fmt.Sprintf("%d%% (%d/%d rounds completed)", progressPercent, roundsCompleted, a.totalRounds)
			} else {
				progressText = "Progress: - / -"
			}

			label := material.Body1(a.theme, progressText)
			label.Alignment = text.Middle
			label.Color = color.NRGBA{R: 70, G: 70, B: 70, A: 255}

			inset := layout.Inset{
				Top:    unit.Dp(8),
				Bottom: unit.Dp(10),
			}
			return inset.Layout(gtx, label.Layout)
		}),
	)
}

// layoutCountdownTimer displays the countdown timer in MM:SS format (Task 25 placeholder)
func (a *App) layoutCountdownTimer(gtx layout.Context) layout.Dimensions {
	// Format remaining time as MM:SS. If no data, show "--:--"
	timerText := "--:--"
	if a.remainingTime > 0 {
		timerText = formatDurationMMSS(a.remainingTime)
	}

	label := material.H3(a.theme, timerText)
	label.Alignment = text.Middle
	label.Color = color.NRGBA{R: 0, G: 120, B: 210, A: 255} // Blue color for timer
	return label.Layout(gtx)
}

// layoutPeriodIndicator displays the current period (Work/Rest) (Task 26 placeholder)
func (a *App) layoutPeriodIndicator(gtx layout.Context) layout.Dimensions {
	periodLabel := "Period: -"
	switch a.currentPeriod {
	case types.PeriodWork:
		periodLabel = "Period: Work"
	case types.PeriodRest:
		periodLabel = "Period: Rest"
	}

	label := material.H5(a.theme, periodLabel)
	label.Alignment = text.Middle
	return label.Layout(gtx)
}

// formatDurationMMSS formats a duration as MM:SS (rounded down to seconds)
func formatDurationMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// layoutComboMoves displays the current combo moves (Task 27)
func (a *App) layoutComboMoves(gtx layout.Context) layout.Dimensions {
	// Only show combo during work periods
	if a.currentPeriod != types.PeriodWork {
		return layout.Dimensions{}
	}

	inset := layout.Inset{
		Left:   unit.Dp(40),
		Right:  unit.Dp(40),
		Top:    unit.Dp(10),
		Bottom: unit.Dp(10),
	}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.currentCombo.IsEmpty() {
			// No combo data yet
			label := material.Body1(a.theme, "Combo moves will appear here")
			label.Alignment = text.Middle
			label.Color = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
			return label.Layout(gtx)
		}

		// Format combo moves as a bulleted list
		return layout.Flex{
			Axis:      layout.Vertical,
			Spacing:   layout.SpaceStart,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Title
				title := material.H6(a.theme, "Current Combo:")
				title.Alignment = text.Middle
				title.Color = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
				return title.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Spacing
				return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Display moves as a formatted string
				comboText := a.formatComboForDisplay(a.currentCombo)
				label := material.Body1(a.theme, comboText)
				label.Alignment = text.Middle
				label.Color = color.NRGBA{R: 40, G: 40, B: 40, A: 255}
				// Add some padding around the text
				textInset := layout.Inset{
					Left:   unit.Dp(20),
					Right:  unit.Dp(20),
					Top:    unit.Dp(8),
					Bottom: unit.Dp(8),
				}
				return textInset.Layout(gtx, label.Layout)
			}),
			// Show "go!" indicator when combo should be performed (at tempo intervals during work period)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Only show "go!" during work periods
				if !a.showGo || a.currentPeriod != types.PeriodWork {
					return layout.Dimensions{}
				}
				return layout.Flex{
					Axis:      layout.Vertical,
					Spacing:   layout.SpaceStart,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Spacing
						return layout.Spacer{Height: unit.Dp(8)}.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// "go!" text
						goLabel := material.H5(a.theme, "go!")
						goLabel.Alignment = text.Middle
						goLabel.Color = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green color
						return goLabel.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// formatComboForDisplay formats a combo for display in the GUI with stance-specific names
func (a *App) formatComboForDisplay(combo models.Combo) string {
	if combo.IsEmpty() {
		return "No moves"
	}

	// Use stance-specific formatting
	return a.formatComboWithStance(combo, a.selectedStance)
}

// layoutCharacterAnimation renders the Scrappy Doo character animation (Tasks 32-34)
func (a *App) layoutCharacterAnimation(gtx layout.Context) layout.Dimensions {
	// Update animation based on current time
	if a.characterSprite != nil {
		// During rest periods, ensure we're always in idle state
		if a.currentPeriod == types.PeriodRest {
			if a.characterSprite.GetCurrentState() != AnimationStateIdle {
				a.characterSprite.SetAnimation(AnimationStateIdle)
			}
		}

		a.characterSprite.Update(time.Now())

		// Animation sequence is now managed by timers in animation_sequence.go
		// No need for continuous checking
	}

	// Render the character
	if a.characterRenderer != nil {
		return a.characterRenderer.Layout(gtx)
	}

	return layout.Dimensions{}
}

// layoutWorkoutControls displays pause/resume and stop buttons (Tasks 28-29)
func (a *App) layoutWorkoutControls(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if a.pauseResumeBtn.Clicked(gtx) {
		a.handlePauseResume()
	}
	if a.stopBtn.Clicked(gtx) {
		a.handleStop()
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Spacing:   layout.SpaceAround,
		Alignment: layout.Middle,
	}.Layout(gtx,
		// Pause/Resume button (Task 28)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var btnText string
				var btnColor color.NRGBA
				if a.isPaused {
					btnText = "Resume"
					btnColor = color.NRGBA{R: 33, G: 150, B: 243, A: 255} // Blue for resume
				} else {
					btnText = "Pause"
					btnColor = color.NRGBA{R: 255, G: 152, B: 0, A: 255} // Orange for pause
				}

				btn := material.Button(a.theme, &a.pauseResumeBtn, btnText)
				btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White text
				btn.Background = btnColor
				btn.CornerRadius = unit.Dp(4)
				return btn.Layout(gtx)
			})
		}),

		// Spacing
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Spacer{Width: unit.Dp(40)}.Layout(gtx)
		}),

		// Stop button (Task 29)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(a.theme, &a.stopBtn, "Stop")
				btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}    // White text
				btn.Background = color.NRGBA{R: 244, G: 67, B: 54, A: 255} // Red for stop
				btn.CornerRadius = unit.Dp(4)
				return btn.Layout(gtx)
			})
		}),
	)
}

// handlePauseResume toggles the pause/resume state (Task 28, Task 58)
func (a *App) handlePauseResume() {
	if a.workoutTimer == nil {
		return
	}

	if a.isPaused {
		// Resume - restart tempo ticker if in work period
		if err := a.workoutTimer.Resume(); err != nil {
			a.setStatusMessage(fmt.Sprintf("Error resuming workout: %v", err), true)
			return
		}
		a.isPaused = false
		// Animation timers will resume automatically when workout timer resumes
	} else {
		// Pause - pause workout timer (animation timers are managed by workout timer)
		a.workoutTimer.Pause()
		a.isPaused = true
	}
}

// handleStop stops the workout and returns to the form (Task 29, Task 58)
func (a *App) handleStop() {
	// Stop audio handler BEFORE stopping workout timer to ensure audio is cancelled
	// The workout timer's Stop() also calls audioHandler.Stop(), but we want to be explicit
	if a.audioHandler != nil {
		a.audioHandler.Stop()
	}

	if a.workoutTimer != nil {
		a.workoutTimer.Stop()
		a.workoutTimer = nil
	}
	a.resetWorkoutState()
}

// resetWorkoutState resets all workout-related state
func (a *App) resetWorkoutState() {
	// Stop timer if running
	if a.workoutTimer != nil {
		a.workoutTimer.Stop()
		a.workoutTimer = nil
	}

	// Stop animation sequence
	a.stopAnimationSequence()

	a.showWorkoutDisplay = false
	a.showCompletion = false
	a.showWorkoutPreview = false
	a.isPaused = false
	// Reset workout state
	a.currentRound = 0
	a.totalRounds = 0
	a.currentPeriod = types.PeriodWork
	a.remainingTime = 0
	a.currentCombo = models.Combo{} // Reset combo
	a.workout = models.Workout{}    // Reset generated workout
	a.showGo = false                // Reset "go!" indicator
	a.stopAnimationSequence()       // Stop any running animation timers
	a.audioHandler = nil            // Clear audio handler

	// Reset character animation to idle (Tasks 32-34)
	if a.characterSprite != nil {
		a.characterSprite.SetAnimation(AnimationStateIdle)
		a.characterSprite.Reset()
	}
}

// handleWorkoutComplete is called when the workout completes (Task 31, Task 64)
func (a *App) handleWorkoutComplete() {
	a.showCompletion = true
	a.showWorkoutDisplay = false
	a.isPaused = false
	// Reset character to idle animation
	if a.characterSprite != nil {
		a.characterSprite.SetAnimation(AnimationStateIdle)
	}
	// Timer will be cleaned up when user clicks "Return to Form"
}

// updateCharacterAnimationForMove updates the character animation based on the current move
// Tasks 35-51: Maps all punches and defensive moves to their corresponding animations
func (a *App) updateCharacterAnimationForMove(move models.Move) {
	if a.characterSprite == nil {
		return
	}

	animationState := a.getAnimationStateForMove(move)
	if animationState != AnimationStateIdle || !move.IsPunch() {
		// Only set animation if it's not idle, or if it's a punch (to trigger punch animations)
		a.characterSprite.SetAnimation(animationState)
	}
}

// getAnimationStateForMove maps a move to its corresponding animation state based on stance
// Tasks 35-51: Complete mapping for all punches and defensive moves
func (a *App) getAnimationStateForMove(move models.Move) AnimationState {
	if move.IsPunch() && move.Punch != nil {
		return a.getPunchAnimationState(*move.Punch)
	} else if move.IsDefensive() && move.Defensive != nil {
		return a.getDefensiveAnimationState(*move.Defensive)
	}
	return AnimationStateIdle
}

// getPunchAnimationState returns the animation state for a punch based on stance
func (a *App) getPunchAnimationState(punch models.Punch) AnimationState {
	stance := a.selectedStance

	switch punch {
	case models.Jab:
		// Task 34-35: Jab animations
		// In orthodox: jab = left hand (lead)
		// In southpaw: jab = right hand (lead)
		if stance == models.Orthodox {
			return AnimationStateJabLeft
		}
		return AnimationStateJabRight

	case models.Cross:
		// Task 36-37: Cross animations
		// In orthodox: cross = right hand (rear)
		// In southpaw: cross = left hand (rear)
		if stance == models.Orthodox {
			return AnimationStateCrossRight
		}
		return AnimationStateCrossLeft

	case models.LeadHook:
		// Task 38-39: Lead hook animations
		// In orthodox: lead hook = left hook
		// In southpaw: lead hook = right hook
		if stance == models.Orthodox {
			return AnimationStateLeadHookLeft
		}
		return AnimationStateLeadHookRight

	case models.RearHook:
		// Task 40-41: Rear hook animations
		// In orthodox: rear hook = right hook
		// In southpaw: rear hook = left hook
		if stance == models.Orthodox {
			return AnimationStateRearHookRight
		}
		return AnimationStateRearHookLeft

	case models.LeadUppercut:
		// Task 42-43: Lead uppercut animations
		// In orthodox: lead uppercut = left uppercut
		// In southpaw: lead uppercut = right uppercut
		if stance == models.Orthodox {
			return AnimationStateLeadUppercutLeft
		}
		return AnimationStateLeadUppercutRight

	case models.RearUppercut:
		// Task 44-45: Rear uppercut animations
		// In orthodox: rear uppercut = right uppercut
		// In southpaw: rear uppercut = left uppercut
		if stance == models.Orthodox {
			return AnimationStateRearUppercutRight
		}
		return AnimationStateRearUppercutLeft
	}

	return AnimationStateIdle
}

// getDefensiveAnimationState returns the animation state for a defensive move
func (a *App) getDefensiveAnimationState(move models.DefensiveMove) AnimationState {
	switch move {
	case models.LeftSlip:
		// Task 46: Left slip animation
		return AnimationStateSlipLeft
	case models.RightSlip:
		// Task 47: Right slip animation
		return AnimationStateSlipRight
	case models.LeftRoll:
		// Task 48: Left roll animation
		return AnimationStateRollLeft
	case models.RightRoll:
		// Task 49: Right roll animation
		return AnimationStateRollRight
	case models.PullBack:
		// Task 50: Pull back animation
		return AnimationStatePullBack
	case models.Duck:
		// Task 51: Duck animation
		return AnimationStateDuck
	}
	return AnimationStateIdle
}

// Old animation functions removed - replaced by timer-based system in animation_sequence.go

// layoutFormFieldWithValidation creates a labeled input field with validation error display
func (a *App) layoutFormFieldWithValidation(gtx layout.Context, label string, editor *widget.Editor, fieldName string) layout.Dimensions {
	// For maxMoves field, validate first to show appropriate error message
	// Then filter input in real-time to prevent values exceeding tempo limit
	if fieldName == "maxMoves" {
		// Capture original value before any filtering
		originalText := strings.TrimSpace(editor.Text())
		var hadOriginalValue bool
		var originalValueExceededLimit bool
		if originalText != "" {
			if val, err := strconv.Atoi(originalText); err == nil {
				hadOriginalValue = true
				tempoLimit := a.selectedTempo.MaxMovesLimit()
				originalValueExceededLimit = val > tempoLimit
			}
		}

		// Validate first on original value to show error message for values exceeding tempo limit
		// We validate before filtering so the error message is set based on what user actually typed
		a.validateField(fieldName)

		// Then filter input in real-time to prevent values exceeding tempo limit
		if hadOriginalValue && originalValueExceededLimit {
			// User entered a value exceeding tempo limit, prevent it by setting to limit
			tempoLimit := a.selectedTempo.MaxMovesLimit()
			editor.SetText(fmt.Sprintf("%d", tempoLimit))
			// Explicitly set the error message to show why value was corrected
			// This ensures the error is visible even after the value is auto-corrected
			a.validationErrors[fieldName] = fmt.Sprintf("Maximum moves cannot exceed %d for %s tempo", tempoLimit, a.selectedTempo.DisplayName())
		}
	} else {
		// Validate the field when it changes
		a.validateField(fieldName)
	}

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Field itself (with help text)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutFormFieldWithHelp(gtx, label, editor, fieldName)
		}),
		// Error message (if any)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if errMsg, hasError := a.validationErrors[fieldName]; hasError && errMsg != "" {
				return layout.Inset{
					Top:  unit.Dp(4),
					Left: unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					errorLabel := material.Body2(a.theme, errMsg)
					errorLabel.Color = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Red color
					return errorLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

// layoutFormFieldWithHelp creates a labeled input field with optional help text
func (a *App) layoutFormFieldWithHelp(gtx layout.Context, label string, editor *widget.Editor, fieldName string) layout.Dimensions {
	helpText := a.getFieldHelpText(fieldName)

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(a.theme, label)
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}),
		// Help text (if available)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if helpText != "" {
				return layout.Inset{
					Top:  unit.Dp(2),
					Left: unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					helpLabel := material.Caption(a.theme, helpText)
					helpLabel.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255} // Gray color for help text
					return helpLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Input field with background
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Ensure we have bounded constraints
			if gtx.Constraints.Max.X == 0 {
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(400)) // Default width
			}
			gtx.Constraints.Min.X = gtx.Constraints.Max.X

			// Use Stack to layer background and editor
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					// Draw background first (bottom layer)
					// Use a fixed reasonable height for single-line editor
					height := gtx.Dp(unit.Dp(40))
					rect := clip.Rect{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{
							X: gtx.Constraints.Max.X,
							Y: height,
						},
					}.Op()
					paint.FillShape(gtx.Ops,
						color.NRGBA{R: 245, G: 245, B: 245, A: 255}, // Light gray background
						rect,
					)
					return layout.Dimensions{
						Size: image.Point{
							X: gtx.Constraints.Max.X,
							Y: height,
						},
					}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					// Editor on top (with padding)
					return layout.Inset{
						Left:   unit.Dp(8),
						Right:  unit.Dp(8),
						Top:    unit.Dp(8),
						Bottom: unit.Dp(8),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						editor := material.Editor(a.theme, editor, "")
						editor.Editor.SingleLine = true
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return editor.Layout(gtx)
					})
				}),
			)
		}),
	)
}

// getFieldHelpText returns help text/tooltip for a given field name
func (a *App) getFieldHelpText(fieldName string) string {
	switch fieldName {
	case "workDuration":
		return "Duration of each work period in seconds (e.g., 20 for 20 seconds)"
	case "restDuration":
		return "Duration of each rest period in seconds (e.g., 10 for 10 seconds)"
	case "totalRounds":
		return "Total number of rounds in the workout"
	case "minMoves":
		return "Minimum number of moves per combo (must be positive)"
	case "maxMoves":
		tempoLimit := a.selectedTempo.MaxMovesLimit()
		return fmt.Sprintf("Maximum number of moves per combo (max %d for %s tempo)", tempoLimit, a.selectedTempo.DisplayName())
	case "openAIAPIKey":
		return "Optional OpenAI API key for LLM-powered workout generation. Can also be set via OPENAI_API_KEY environment variable."
	case "pattern":
		return "Workout pattern: Linear (increasing complexity), Pyramid (up then down), Random (varied), Constant (same complexity)"
	case "stance":
		return "Boxing stance: Orthodox (right-handed) or Southpaw (left-handed)"
	case "tempo":
		return "Tempo determines the speed of combo intervals: Slow (5s), Medium (4s), Fast (3s), Superfast (1s)"
	case "includeDefensive":
		return "Include defensive moves (slips, rolls, pull back, duck) in generated combos"
	case "useLLM":
		return "Use AI-powered workout generation (requires OpenAI API key)"
	case "preset":
		return "Quick-start with a preset workout configuration"
	default:
		return ""
	}
}

// layoutFormField creates a labeled input field
func (a *App) layoutFormField(gtx layout.Context, label string, editor *widget.Editor) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(a.theme, label)
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}),

		// Input field with background
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Ensure we have bounded constraints
			if gtx.Constraints.Max.X == 0 {
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(400)) // Default width
			}
			gtx.Constraints.Min.X = gtx.Constraints.Max.X

			// Use Stack to layer background and editor
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					// Draw background first (bottom layer)
					// Use a fixed reasonable height for single-line editor
					height := gtx.Dp(unit.Dp(40))
					rect := clip.Rect{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{
							X: gtx.Constraints.Max.X,
							Y: height,
						},
					}.Op()
					paint.FillShape(gtx.Ops,
						color.NRGBA{R: 245, G: 245, B: 245, A: 255}, // Light gray background
						rect,
					)
					return layout.Dimensions{
						Size: image.Point{
							X: gtx.Constraints.Max.X,
							Y: height,
						},
					}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					// Editor on top (with padding)
					return layout.Inset{
						Left:   unit.Dp(8),
						Right:  unit.Dp(8),
						Top:    unit.Dp(8),
						Bottom: unit.Dp(8),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						editor := material.Editor(a.theme, editor, "")
						editor.Editor.SingleLine = true
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return editor.Layout(gtx)
					})
				}),
			)
		}),
	)
}

// layoutPatternDropdown creates a dropdown selector for workout pattern
func (a *App) layoutPatternDropdown(gtx layout.Context) layout.Dimensions {
	// Check if pattern button was clicked
	if a.patternButton.Clicked(gtx) {
		a.patternDropdownOpen = !a.patternDropdownOpen
	}

	// Check if any pattern option was clicked
	patterns := []models.WorkoutPatternType{
		models.PatternLinear,
		models.PatternPyramid,
		models.PatternRandom,
		models.PatternConstant,
	}

	for i := range patterns {
		if a.patternOptions[i].Clicked(gtx) {
			a.selectedPattern = patterns[i]
			a.patternDropdownOpen = false
		}
	}

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(a.theme, "Workout Pattern")
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}),
		// Help text
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			helpText := a.getFieldHelpText("pattern")
			if helpText != "" {
				return layout.Inset{
					Top:  unit.Dp(2),
					Left: unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					helpLabel := material.Caption(a.theme, helpText)
					helpLabel.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255} // Gray color for help text
					return helpLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Dropdown button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			patternText := strings.Title(string(a.selectedPattern))
			btn := material.Button(a.theme, &a.patternButton, patternText)
			// Don't force full width - let button size naturally
			return btn.Layout(gtx)
		}),

		// Dropdown options (shown when open)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.patternDropdownOpen {
				return layout.Dimensions{}
			}

			return layout.Inset{
				Top:   unit.Dp(5),
				Left:  unit.Dp(10),
				Right: unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Filter out the selected pattern from options
				allPatterns := []struct {
					label   string
					pattern models.WorkoutPatternType
					index   int
				}{
					{"Linear", models.PatternLinear, 0},
					{"Pyramid", models.PatternPyramid, 1},
					{"Random", models.PatternRandom, 2},
					{"Constant", models.PatternConstant, 3},
				}

				var options []layout.FlexChild
				for _, p := range allPatterns {
					if p.pattern == a.selectedPattern {
						continue // Skip the selected pattern
					}
					if len(options) > 0 {
						options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Height: unit.Dp(5)}.Layout(gtx)
						}))
					}
					pattern := p.pattern
					index := p.index
					options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutPatternOption(gtx, p.label, pattern, &a.patternOptions[index])
					}))
				}

				return layout.Flex{
					Axis:      layout.Vertical,
					Spacing:   layout.SpaceStart,
					Alignment: layout.Start,
				}.Layout(gtx, options...)
			})
		}),
	)
}

// layoutPatternOption creates a clickable option for the pattern dropdown
func (a *App) layoutPatternOption(gtx layout.Context, label string, pattern models.WorkoutPatternType, clickable *widget.Clickable) layout.Dimensions {
	btn := material.Button(a.theme, clickable, label)
	if a.selectedPattern == pattern {
		// Highlight selected option
		btn.Background = a.theme.ContrastBg
	}

	// Don't force full width - let button size naturally
	return btn.Layout(gtx)
}

// layoutCheckbox creates a labeled checkbox with optional help text
func (a *App) layoutCheckbox(gtx layout.Context, label string, checkbox *widget.Bool, fieldName string) layout.Dimensions {
	helpText := a.getFieldHelpText(fieldName)

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Checkbox and label row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Spacing:   layout.SpaceSides,
				Alignment: layout.Middle,
			}.Layout(gtx,
				// Checkbox
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					chk := material.CheckBox(a.theme, checkbox, "")
					return chk.Layout(gtx)
				}),
				// Label text
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(a.theme, label)
					lbl.Alignment = text.Start
					return lbl.Layout(gtx)
				}),
			)
		}),
		// Help text (if available)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if helpText != "" {
				return layout.Inset{
					Top:  unit.Dp(2),
					Left: unit.Dp(28), // Indent to align with label text (checkbox width + spacing)
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					helpLabel := material.Caption(a.theme, helpText)
					helpLabel.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255} // Gray color for help text
					return helpLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

// layoutStanceDropdown creates a dropdown selector for stance
func (a *App) layoutStanceDropdown(gtx layout.Context) layout.Dimensions {
	// Check if stance button was clicked
	if a.stanceButton.Clicked(gtx) {
		a.stanceDropdownOpen = !a.stanceDropdownOpen
	}

	// Check if any stance option was clicked
	stances := []models.Stance{models.Orthodox, models.Southpaw}
	stanceLabels := []string{"Orthodox", "Southpaw"}

	for i := range stances {
		if a.stanceOptions[i].Clicked(gtx) {
			a.selectedStance = stances[i]
			a.stanceDropdownOpen = false
		}
	}

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(a.theme, "Stance")
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}),
		// Help text
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			helpText := a.getFieldHelpText("stance")
			if helpText != "" {
				return layout.Inset{
					Top:  unit.Dp(2),
					Left: unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					helpLabel := material.Caption(a.theme, helpText)
					helpLabel.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255} // Gray color for help text
					return helpLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Dropdown button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			stanceText := strings.Title(a.selectedStance.String())
			btn := material.Button(a.theme, &a.stanceButton, stanceText)
			return btn.Layout(gtx)
		}),

		// Dropdown options (shown when open)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.stanceDropdownOpen {
				return layout.Dimensions{}
			}

			return layout.Inset{
				Top:   unit.Dp(5),
				Left:  unit.Dp(10),
				Right: unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var options []layout.FlexChild
				for i, stance := range stances {
					if stance == a.selectedStance {
						continue // Skip the selected stance
					}
					if len(options) > 0 {
						options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Height: unit.Dp(5)}.Layout(gtx)
						}))
					}
					index := i
					options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(a.theme, &a.stanceOptions[index], stanceLabels[index])
						return btn.Layout(gtx)
					}))
				}

				return layout.Flex{
					Axis:      layout.Vertical,
					Spacing:   layout.SpaceStart,
					Alignment: layout.Start,
				}.Layout(gtx, options...)
			})
		}),
	)
}

// layoutTempoDropdown creates a dropdown selector for tempo
func (a *App) layoutTempoDropdown(gtx layout.Context) layout.Dimensions {
	// Check if tempo button was clicked
	if a.tempoButton.Clicked(gtx) {
		a.tempoDropdownOpen = !a.tempoDropdownOpen
	}

	// Check if any tempo option was clicked
	tempos := models.AllTempos()

	for i := range tempos {
		if a.tempoOptions[i].Clicked(gtx) {
			oldTempo := a.selectedTempo
			a.selectedTempo = tempos[i]
			a.tempoDropdownOpen = false
			// If tempo changed, re-validate max moves against new tempo limit
			if oldTempo != a.selectedTempo {
				a.validateField("maxMoves")
			}
		}
	}

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(a.theme, "Tempo")
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}),
		// Help text
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			helpText := a.getFieldHelpText("tempo")
			if helpText != "" {
				return layout.Inset{
					Top:  unit.Dp(2),
					Left: unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					helpLabel := material.Caption(a.theme, helpText)
					helpLabel.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255} // Gray color for help text
					return helpLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Dropdown button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(a.theme, &a.tempoButton, a.selectedTempo.DisplayName())
			return btn.Layout(gtx)
		}),

		// Dropdown options (shown when open)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.tempoDropdownOpen {
				return layout.Dimensions{}
			}

			return layout.Inset{
				Top:   unit.Dp(5),
				Left:  unit.Dp(10),
				Right: unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var options []layout.FlexChild
				for i, tempo := range tempos {
					if tempo == a.selectedTempo {
						continue // Skip the selected tempo
					}
					if len(options) > 0 {
						options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Height: unit.Dp(5)}.Layout(gtx)
						}))
					}
					tempoVal := tempo
					index := i
					options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(a.theme, &a.tempoOptions[index], tempoVal.DisplayName())
						return btn.Layout(gtx)
					}))
				}

				return layout.Flex{
					Axis:      layout.Vertical,
					Spacing:   layout.SpaceStart,
					Alignment: layout.Start,
				}.Layout(gtx, options...)
			})
		}),
	)
}

// layoutPresetDropdown creates a dropdown selector for workout presets
func (a *App) layoutPresetDropdown(gtx layout.Context) layout.Dimensions {
	// Check if preset button was clicked
	if a.presetButton.Clicked(gtx) {
		a.presetDropdownOpen = !a.presetDropdownOpen
	}

	// Available presets
	presets := models.AvailablePresets()
	presetLabels := []string{"Beta Style", "Endurance", "Power"}
	allPresets := []struct {
		preset *models.WorkoutPreset
		label  string
		index  int
	}{
		{nil, "Custom", 3}, // "Custom" option (nil preset)
	}
	for i, preset := range presets {
		p := preset
		allPresets = append(allPresets, struct {
			preset *models.WorkoutPreset
			label  string
			index  int
		}{&p, presetLabels[i], i})
	}

	// Check if any preset option was clicked
	for i := range allPresets {
		if i < len(a.presetOptions) && a.presetOptions[i].Clicked(gtx) {
			a.selectedPreset = allPresets[i].preset
			a.presetDropdownOpen = false
			// Populate fields if a preset was selected
			if a.selectedPreset != nil {
				a.applyPreset(*a.selectedPreset)
			}
		}
	}

	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceStart,
		Alignment: layout.Start,
	}.Layout(gtx,
		// Label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(a.theme, "Preset")
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}),
		// Help text
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			helpText := a.getFieldHelpText("preset")
			if helpText != "" {
				return layout.Inset{
					Top:  unit.Dp(2),
					Left: unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					helpLabel := material.Caption(a.theme, helpText)
					helpLabel.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255} // Gray color for help text
					return helpLabel.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Dropdown button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			presetText := "Custom"
			if a.selectedPreset != nil {
				switch *a.selectedPreset {
				case models.PresetBetaStyle:
					presetText = "Beta Style"
				case models.PresetEndurance:
					presetText = "Endurance"
				case models.PresetPower:
					presetText = "Power"
				}
			}
			btn := material.Button(a.theme, &a.presetButton, presetText)
			return btn.Layout(gtx)
		}),

		// Dropdown options (shown when open)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.presetDropdownOpen {
				return layout.Dimensions{}
			}

			return layout.Inset{
				Top:   unit.Dp(5),
				Left:  unit.Dp(10),
				Right: unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var options []layout.FlexChild
				for i, presetOption := range allPresets {
					// Skip the selected preset
					if (presetOption.preset == nil && a.selectedPreset == nil) ||
						(presetOption.preset != nil && a.selectedPreset != nil && *presetOption.preset == *a.selectedPreset) {
						continue
					}
					if len(options) > 0 {
						options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Height: unit.Dp(5)}.Layout(gtx)
						}))
					}
					label := presetOption.label
					index := i
					options = append(options, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(a.theme, &a.presetOptions[index], label)
						return btn.Layout(gtx)
					}))
				}

				return layout.Flex{
					Axis:      layout.Vertical,
					Spacing:   layout.SpaceStart,
					Alignment: layout.Start,
				}.Layout(gtx, options...)
			})
		}),
	)
}

// applyPreset populates form fields based on the selected preset
func (a *App) applyPreset(preset models.WorkoutPreset) {
	config := models.PresetWorkoutConfig(preset)
	a.workDurationEditor.SetText(fmt.Sprintf("%d", int(config.WorkDuration.Seconds())))
	a.restDurationEditor.SetText(fmt.Sprintf("%d", int(config.RestDuration.Seconds())))
	a.totalRoundsEditor.SetText(fmt.Sprintf("%d", config.TotalRounds))
	// Clear validation errors when preset is applied
	delete(a.validationErrors, "workDuration")
	delete(a.validationErrors, "restDuration")
	delete(a.validationErrors, "totalRounds")
}

// validateField validates a specific field and stores the error message
func (a *App) validateField(fieldName string) {
	switch fieldName {
	case "workDuration":
		text := a.workDurationEditor.Text()
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || val <= 0 {
			a.validationErrors[fieldName] = "Work duration must be a positive integer"
		} else {
			delete(a.validationErrors, fieldName)
		}

	case "restDuration":
		text := a.restDurationEditor.Text()
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || val < 0 {
			a.validationErrors[fieldName] = "Rest duration must be a non-negative integer"
		} else {
			delete(a.validationErrors, fieldName)
		}

	case "totalRounds":
		text := a.totalRoundsEditor.Text()
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || val <= 0 {
			a.validationErrors[fieldName] = "Total rounds must be a positive integer"
		} else {
			delete(a.validationErrors, fieldName)
		}

	case "minMoves":
		text := a.minMovesEditor.Text()
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || val <= 0 {
			a.validationErrors[fieldName] = "Minimum moves must be a positive integer"
		} else {
			delete(a.validationErrors, fieldName)
			// Also validate max moves relative to min moves
			if maxText := strings.TrimSpace(a.maxMovesEditor.Text()); maxText != "" {
				a.validateField("maxMoves")
			}
		}

	case "maxMoves":
		text := strings.TrimSpace(a.maxMovesEditor.Text())
		if text == "" {
			// Empty text is valid (user might be clearing the field)
			delete(a.validationErrors, fieldName)
		} else {
			maxVal, err := strconv.Atoi(text)
			if err != nil {
				// Invalid input (non-numeric)
				a.validationErrors[fieldName] = "Maximum moves must be a positive integer"
			} else if maxVal <= 0 {
				// Zero or negative
				a.validationErrors[fieldName] = "Maximum moves must be a positive integer"
			} else {
				// Valid number - check against tempo-based limit first
				tempoLimit := a.selectedTempo.MaxMovesLimit()
				if maxVal > tempoLimit {
					// Value exceeds tempo limit - set error message
					a.validationErrors[fieldName] = fmt.Sprintf("Maximum moves cannot exceed %d for %s tempo", tempoLimit, a.selectedTempo.DisplayName())
				} else if maxVal == tempoLimit {
					// Value is exactly at the limit - check if there's an existing error about exceeding limit
					// If so, keep it visible briefly to show why value was corrected
					existingError, hasError := a.validationErrors[fieldName]
					if hasError && strings.Contains(existingError, "cannot exceed") {
						// Keep the error message even though value is now valid
						// This helps user understand why the value was auto-corrected
						// Error will be cleared when user types a different value
					} else {
						// No existing error or different error - check if max >= min
						minText := strings.TrimSpace(a.minMovesEditor.Text())
						if minText != "" {
							minVal, minErr := strconv.Atoi(minText)
							if minErr == nil && maxVal < minVal {
								a.validationErrors[fieldName] = fmt.Sprintf("Maximum moves (%d) must be >= minimum moves (%d)", maxVal, minVal)
							} else {
								delete(a.validationErrors, fieldName)
							}
						} else {
							delete(a.validationErrors, fieldName)
						}
					}
				} else {
					// Value is below limit - check if max >= min
					minText := strings.TrimSpace(a.minMovesEditor.Text())
					if minText != "" {
						minVal, minErr := strconv.Atoi(minText)
						if minErr == nil && maxVal < minVal {
							a.validationErrors[fieldName] = fmt.Sprintf("Maximum moves (%d) must be >= minimum moves (%d)", maxVal, minVal)
						} else {
							delete(a.validationErrors, fieldName)
						}
					} else {
						delete(a.validationErrors, fieldName)
					}
				}
			}
		}

	case "openAIAPIKey":
		// API key is optional, but if LLM is enabled and provided, it should be non-empty
		text := strings.TrimSpace(a.openAIAPIKeyEditor.Text())
		if a.useLLM.Value && text == "" {
			// Note: API key can also come from environment, so this is just a warning
			// We'll allow empty and let the backend handle it
			delete(a.validationErrors, fieldName)
		} else {
			delete(a.validationErrors, fieldName)
		}
	}
}

// ValidateAllFields validates all form fields
func (a *App) ValidateAllFields() bool {
	fields := []string{"workDuration", "restDuration", "totalRounds", "minMoves", "maxMoves"}
	if a.useLLM.Value {
		fields = append(fields, "openAIAPIKey")
	}

	for _, field := range fields {
		a.validateField(field)
	}

	// Return true if no errors
	return len(a.validationErrors) == 0
}

// handleStartWorkout validates the form and generates/starts the workout (Task 57)
func (a *App) handleStartWorkout() {
	// Validate all fields
	if !a.ValidateAllFields() {
		a.setStatusMessage("Please fix validation errors before starting the workout", true)
		return
	}

	// Extract form values
	workSeconds, err := strconv.Atoi(strings.TrimSpace(a.workDurationEditor.Text()))
	if err != nil || workSeconds <= 0 {
		a.setStatusMessage("Invalid work duration", true)
		return
	}

	restSeconds, err := strconv.Atoi(strings.TrimSpace(a.restDurationEditor.Text()))
	if err != nil || restSeconds < 0 {
		a.setStatusMessage("Invalid rest duration", true)
		return
	}

	totalRounds, err := strconv.Atoi(strings.TrimSpace(a.totalRoundsEditor.Text()))
	if err != nil || totalRounds <= 0 {
		a.setStatusMessage("Invalid total rounds", true)
		return
	}

	minMoves, err := strconv.Atoi(strings.TrimSpace(a.minMovesEditor.Text()))
	if err != nil || minMoves <= 0 {
		a.setStatusMessage("Invalid minimum moves", true)
		return
	}

	maxMoves, err := strconv.Atoi(strings.TrimSpace(a.maxMovesEditor.Text()))
	if err != nil || maxMoves <= 0 {
		a.setStatusMessage("Invalid maximum moves", true)
		return
	}

	// Enforce tempo-based maximum limit
	tempoLimit := a.selectedTempo.MaxMovesLimit()
	if maxMoves > tempoLimit {
		a.setStatusMessage(fmt.Sprintf("Maximum moves per combo cannot exceed %d for %s tempo", tempoLimit, a.selectedTempo.DisplayName()), true)
		return
	}

	if maxMoves < minMoves {
		a.setStatusMessage("Maximum moves must be >= minimum moves", true)
		return
	}

	// Create workout configuration
	workoutConfig := models.NewWorkoutConfig(
		time.Duration(workSeconds)*time.Second,
		time.Duration(restSeconds)*time.Second,
		totalRounds,
	)

	// Create workout pattern
	includeDefensive := a.includeDefensive.Value
	workoutPattern := models.NewWorkoutPattern(
		a.selectedPattern,
		minMoves,
		maxMoves,
		includeDefensive,
	)

	// Generate workout
	var workout models.Workout
	if a.useLLM.Value {
		// LLM generation path
		apiKey := strings.TrimSpace(a.openAIAPIKeyEditor.Text())
		if apiKey == "" {
			// Try environment variable
			apiKey = a.getOpenAIAPIKeyFromEnv()
		}
		if apiKey == "" {
			a.setStatusMessage("OpenAI API key required for LLM generation", true)
			return
		}

		a.setStatusMessage("Generating workout with LLM...", false)
		// Note: In a real GUI, you'd want to show this in a non-blocking way
		// For now, we'll generate synchronously (could be improved with goroutines)
		llmGenerator := generator.NewLLMWorkoutGenerator(apiKey)
		var genErr error
		workout, genErr = llmGenerator.GenerateWorkoutWithStance(workoutConfig, workoutPattern, a.selectedStance)
		if genErr != nil {
			a.setStatusMessage(fmt.Sprintf("Error generating workout with LLM: %v", genErr), true)
			return
		}
	} else {
		// In-house generation path
		a.setStatusMessage("Generating workout...", false)
		inHouseGenerator := generator.NewWorkoutGenerator()
		var genErr error
		workout, genErr = inHouseGenerator.GenerateWorkout(workoutConfig, workoutPattern)
		if genErr != nil {
			a.setStatusMessage(fmt.Sprintf("Error generating workout: %v", genErr), true)
			return
		}
	}

	// Store the generated workout
	a.workout = workout

	// Initialize workout display state
	a.totalRounds = totalRounds
	a.currentRound = 0 // Will be updated when timer starts (task 58)
	a.currentPeriod = types.PeriodWork
	a.remainingTime = time.Duration(workSeconds) * time.Second

	// Switch to workout preview screen for confirmation
	a.showWorkoutPreview = true
	a.showWorkoutDisplay = false
	// Clear status messages when switching screens
	a.statusMessage = ""
}

// getOpenAIAPIKeyFromEnv retrieves OpenAI API key from environment variable
func (a *App) getOpenAIAPIKeyFromEnv() string {
	return os.Getenv("OPENAI_API_KEY")
}

// TimerDisplayHandler implementation
// These methods are called by the workout timer to update the GUI
// OnTimerUpdate is called on each timer tick to update the remaining time
func (a *App) OnTimerUpdate(remaining time.Duration, periodType types.PeriodType, roundNumber int) {
	a.remainingTime = remaining
	a.currentPeriod = periodType
	a.currentRound = roundNumber
	// Invalidate window to trigger redraw
	if a.window != nil {
		a.window.Invalidate()
	}
}

// OnPeriodStart is called when a period (work or rest) starts
func (a *App) OnPeriodStart(periodType types.PeriodType, roundNumber int, duration time.Duration) {
	a.currentPeriod = periodType
	a.currentRound = roundNumber
	a.remainingTime = duration

	// Update current combo for the round
	if roundNumber > 0 && roundNumber <= len(a.workout.Rounds) {
		round := a.workout.Rounds[roundNumber-1]
		a.currentCombo = round.Combo
	}

	if periodType == types.PeriodWork {
		// Store workout period start time and duration for period check
		a.workoutStartTime = time.Now()
		a.workoutPeriodDuration = duration

		// Set to idle initially, wait for first beep
		if a.characterSprite != nil {
			a.characterSprite.SetAnimation(AnimationStateIdle)
		}

		// Start animation sequence (beep will play at start of sequence)
		a.startAnimationSequence()
	} else {
		// Rest period - stop any work period animations
		a.stopAnimationSequence()
		a.showGo = false

		// Start rest period animation
		a.startRestPeriodAnimation()
	}

	// Invalidate window to trigger redraw
	if a.window != nil {
		a.window.Invalidate()
	}
}

// OnPeriodEnd is called when a period ends
func (a *App) OnPeriodEnd(periodType types.PeriodType, roundNumber int) {
	// Stop animation sequence when work period ends
	if periodType == types.PeriodWork {
		a.stopAnimationSequence()
		a.showGo = false
	}
	// Period end - timer will call OnPeriodStart for next period
}

// OnWorkoutStart is called when the workout starts
func (a *App) OnWorkoutStart(totalRounds int) {
	a.totalRounds = totalRounds
	a.currentRound = 1
	a.isPaused = false
	// Invalidate window to trigger redraw
	if a.window != nil {
		a.window.Invalidate()
	}
}

// OnWorkoutComplete is called when the workout completes
func (a *App) OnWorkoutComplete() {
	a.handleWorkoutComplete()
}

// handleLoadConfig loads configuration from a file
func (a *App) handleLoadConfig() {
	filePath := strings.TrimSpace(a.configFilePathEditor.Text())
	if filePath == "" {
		a.setStatusMessage("Please enter a config file path", true)
		return
	}

	// Import config package
	cfg, err := a.loadConfigFromFile(filePath)
	if err != nil {
		a.setStatusMessage(fmt.Sprintf("Error loading config: %v", err), true)
		return
	}

	// Populate form fields from config
	a.populateFromConfig(cfg)
	a.setStatusMessage(fmt.Sprintf("Config loaded from %s", filePath), false)
}

// handleSaveConfig saves the current form state to a config file
func (a *App) handleSaveConfig() {
	filePath := strings.TrimSpace(a.configFilePathEditor.Text())
	if filePath == "" {
		a.setStatusMessage("Please enter a config file path", true)
		return
	}

	// Validate all fields before saving
	if !a.ValidateAllFields() {
		a.setStatusMessage("Please fix validation errors before saving", true)
		return
	}

	// Create config from form
	cfg := a.createConfigFromForm()

	// Save to file
	if err := cfg.SaveToFile(filePath); err != nil {
		a.setStatusMessage(fmt.Sprintf("Error saving config: %v", err), true)
		return
	}

	a.setStatusMessage(fmt.Sprintf("Config saved to %s", filePath), false)
}

// setStatusMessage sets a status message to display
func (a *App) setStatusMessage(message string, isError bool) {
	a.statusMessage = message
	a.statusError = isError
	// Clear message after 5 seconds (this would need to be handled in the event loop)
	// For now, it will persist until the next action
}

// loadConfigFromFile loads a config file (wrapper to avoid circular import)
func (a *App) loadConfigFromFile(filePath string) (*config.AppConfig, error) {
	// We need to import the config package, but to avoid circular imports,
	// we'll use a type alias or import it here
	// For now, let's use the full path
	return config.LoadFromFile(filePath)
}

// populateFromConfig populates form fields from a config
func (a *App) populateFromConfig(cfg *config.AppConfig) {
	// Workout config
	a.workDurationEditor.SetText(fmt.Sprintf("%d", cfg.Workout.WorkDurationSeconds))
	a.restDurationEditor.SetText(fmt.Sprintf("%d", cfg.Workout.RestDurationSeconds))
	a.totalRoundsEditor.SetText(fmt.Sprintf("%d", cfg.Workout.TotalRounds))

	// Pattern config
	a.minMovesEditor.SetText(fmt.Sprintf("%d", cfg.Pattern.MinMoves))
	// Enforce tempo-based maximum limit when loading from config
	maxMovesFromConfig := cfg.Pattern.MaxMoves
	tempoLimit := a.selectedTempo.MaxMovesLimit()
	if maxMovesFromConfig > tempoLimit {
		maxMovesFromConfig = tempoLimit
	}
	a.maxMovesEditor.SetText(fmt.Sprintf("%d", maxMovesFromConfig))
	a.includeDefensive.Value = cfg.Pattern.IncludeDefensive

	// Set pattern type
	switch cfg.Pattern.Type {
	case "linear":
		a.selectedPattern = models.PatternLinear
	case "pyramid":
		a.selectedPattern = models.PatternPyramid
	case "random":
		a.selectedPattern = models.PatternRandom
	case "constant":
		a.selectedPattern = models.PatternConstant
	}

	// Set stance
	switch cfg.GetStance() {
	case "orthodox":
		a.selectedStance = models.Orthodox
	case "southpaw":
		a.selectedStance = models.Southpaw
	}

	// Generator config
	a.useLLM.Value = cfg.Generator.UseLLM
	if cfg.OpenAIAPIKey != "" {
		a.openAIAPIKeyEditor.SetText(cfg.OpenAIAPIKey)
	}

	// Clear preset selection when loading from file
	a.selectedPreset = nil

	// Clear validation errors
	a.validationErrors = make(map[string]string)
}

// createConfigFromForm creates a config from the current form state
func (a *App) createConfigFromForm() *config.AppConfig {
	workDuration, _ := strconv.Atoi(strings.TrimSpace(a.workDurationEditor.Text()))
	restDuration, _ := strconv.Atoi(strings.TrimSpace(a.restDurationEditor.Text()))
	totalRounds, _ := strconv.Atoi(strings.TrimSpace(a.totalRoundsEditor.Text()))
	minMoves, _ := strconv.Atoi(strings.TrimSpace(a.minMovesEditor.Text()))
	maxMoves, _ := strconv.Atoi(strings.TrimSpace(a.maxMovesEditor.Text()))

	patternType := string(a.selectedPattern)
	stance := a.selectedStance.String()

	return &config.AppConfig{
		Workout: config.WorkoutConfig{
			WorkDurationSeconds: workDuration,
			RestDurationSeconds: restDuration,
			TotalRounds:         totalRounds,
		},
		Pattern: config.PatternConfig{
			Type:             patternType,
			MinMoves:         minMoves,
			MaxMoves:         maxMoves,
			IncludeDefensive: a.includeDefensive.Value,
		},
		Generator: config.GeneratorConfig{
			UseLLM:   a.useLLM.Value,
			LLMModel: "gpt-4o-mini", // Default model
		},
		Stance:       stance,
		OpenAIAPIKey: strings.TrimSpace(a.openAIAPIKeyEditor.Text()),
	}
}
