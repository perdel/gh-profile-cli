package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"
)

// Placeholder for future functions to scan git repos and collect commit data

func main() {
	// We will later add logic here to find git repositories and process commits
	// For now, let's just display a placeholder message.

	app := tview.NewApplication()

	textView := tview.NewTextView().
		SetText("Scanning for Git repositories and building contribution graph... (Not yet implemented)").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	if err := app.SetRoot(textView, true).Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
