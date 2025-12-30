package main

import (
	"heavybagworkout/internal/gui"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
)

func main() {
	// Create window in a goroutine - this is the correct pattern for Gio
	// The window event loop runs in this goroutine, while app.Main() runs on the main thread
	go func() {
		var w app.Window
		w.Option(app.Title("Puppy Power - Heavy Bag Workout"))
		w.Option(app.Size(800, 600))

		// Create GUI application
		guiApp := gui.NewApp()
		// Set window reference for invalidating frames on timer updates
		guiApp.SetWindow(&w)

		// Run the window event loop in this goroutine
		if err := run(&w, guiApp); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	// app.Main() must be called on the main thread
	// It hands over control to the OS and blocks until all windows are closed
	app.Main()
}

// run handles the window event loop
func run(w *app.Window, guiApp *gui.App) error {
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			// Window close event (Task 30)
			// Check if workout is in progress
			if guiApp.IsWorkoutInProgress() {
				// Workout is in progress - we can't easily prevent window close in Gio
				// without a modal dialog system. For now, we'll allow the close but
				// the workout state will be lost. In a production app, you might want
				// to implement a modal dialog system to confirm.
				// TODO: Implement modal dialog for confirmation if needed
				log.Println("Warning: Closing window during active workout - workout will be lost")
			}
			return e.Err
		case app.FrameEvent:
			// Frame event - render the UI
			ops.Reset()
			gtx := layout.Context{
				Ops:    &ops,
				Now:    e.Now,
				Metric: e.Metric,
				Source: e.Source,
			}
			gtx.Constraints = layout.Exact(e.Size)
			guiApp.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
