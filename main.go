package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// ContributionGraph maps a date string (YYYY-MM-DD) to the number of commits on that day.
type ContributionGraph map[string]int

// Constants for TUI colors representing different contribution levels
const (
	ColorLevel0 = "[black:darkgray]"   // No commits
	ColorLevel1 = "[black:darkgreen]"  // 1-4 commits
	ColorLevel2 = "[black:green]"      // 5-9 commits
	ColorLevel3 = "[black:lightgreen]" // 10-19 commits
	ColorLevel4 = "[black:lime]"       // 20+ commits
)

func findGitRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// If there's an error accessing a path, log it and continue
			log.Printf("Error accessing path %q: %v", path, err)
			return filepath.SkipDir // Skip this directory
		}

		if d.IsDir() && d.Name() == ".git" {
			// Found a .git directory, add its parent directory (the repo root)
			repoPath := filepath.Dir(path)
			repos = append(repos, repoPath)
			return filepath.SkipDir // Skip scanning inside the .git directory
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %q: %w", root, err)
	}
	return repos, nil
}

func getCommits(repoPath string) ([]time.Time, error) {
	// Use --pretty=format:%ad to get only the author date
	// Use --date=short to get dates in YYYY-MM-DD format
	cmd := exec.Command("git", "-C", repoPath, "log", "--pretty=format:%ad", "--date=short")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Log stderr for more context on git command failures
		return nil, fmt.Errorf("failed to run git log in %q: %w\nStderr: %s", repoPath, err, stderr.String())
	}

	var commits []time.Time
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		t, err := time.Parse("2006-01-02", line)
		if err != nil {
			log.Printf("Warning: Could not parse date %q from repo %q: %v", line, repoPath, err)
			continue
		}
		commits = append(commits, t)
	}
	return commits, nil
}

func aggregateCommits(allCommits []time.Time) ContributionGraph {
	graph := make(ContributionGraph)
	for _, commitTime := range allCommits {
		dateStr := commitTime.Format("2006-01-02")
		graph[dateStr]++
	}
	return graph
}

func getColorForCommits(count int) string {
	switch {
	case count >= 20:
		return ColorLevel4
	case count >= 10:
		return ColorLevel3
	case count >= 5:
		return ColorLevel2
	case count >= 1:
		return ColorLevel1
	default:
		return ColorLevel0
	}
}

func renderContributionGraph(graph ContributionGraph) tview.Primitive {
	grid := tview.NewGrid().
		SetRows(1, 1, 1, 1, 1, 1, 1, 0).                                                                                                                                              // 1 row for month labels, 7 rows for days of week, 0 for flexible height
		SetColumns(2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0). // 2 for day labels, then 53 columns for weeks
		SetBorders(false)

	// Day of week labels (Sunday, Monday, ..., Saturday)
	dayLabels := []string{"", "Mon", "", "Wed", "", "Fri", ""} // Only show Mon, Wed, Fri for brevity
	for i, label := range dayLabels {
		if label != "" {
			grid.AddItem(tview.NewTextView().SetText(label).SetTextAlign(tview.AlignRight).SetDynamicColors(true),
				i+1, 0, 1, 1, 0, 0, false)
		}
	}

	now := time.Now()
	// Start date is 365 days ago from today
	startDate := now.AddDate(0, 0, -365)
	// Adjust startDate to be the first Sunday before or on the actual start date
	// This ensures the graph starts on a Sunday for consistent alignment
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	currentDate := startDate
	col := 1 // Start filling from the second column (after day labels)
	lastMonth := -1

	// Map to keep track of month label positions
	monthColStart := make(map[int]int) // month -> starting column

	for {
		if currentDate.After(now) {
			break
		}

		// Add month label if it's a new month and the first day of the week for that month
		if currentDate.Month() != time.Month(lastMonth) && currentDate.Weekday() == time.Sunday {
			monthColStart[int(currentDate.Month())] = col
			lastMonth = int(currentDate.Month())
		}

		dateStr := currentDate.Format("2006-01-02")
		commits := graph[dateStr]
		color := getColorForCommits(commits)

		// Add the day cell
		cell := tview.NewTextView().
			SetText(fmt.Sprintf("%s %s", color, tview.Escape("[ ]"))). // Use a space with background color
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true)

		// Row is based on Weekday (Sunday=0, Monday=1, ..., Saturday=6)
		// Add 1 to row index because row 0 is for month labels
		grid.AddItem(cell, int(currentDate.Weekday())+1, col, 1, 1, 0, 0, false)

		// Move to the next day
		currentDate = currentDate.AddDate(0, 0, 1)

		// If it's Saturday, move to the next column (next week)
		if currentDate.Weekday() == time.Sunday {
			col++
		}
	}

	// Add month labels to the top row
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	for month, startCol := range monthColStart {
		// Calculate approximate span for the month label.
		// This is a rough estimate and might need fine-tuning based on actual week lengths.
		endCol := startCol + 4 // Assume a month spans about 4-5 weeks
		if month < 12 {        // For all months except December, check if the next month starts soon
			if nextMonthStartCol, ok := monthColStart[month+1]; ok {
				endCol = nextMonthStartCol - 1
			}
		}
		if endCol < startCol { // Ensure endCol is not before startCol
			endCol = startCol
		}

		monthLabel := tview.NewTextView().
			SetText(monthNames[month-1]).
			SetTextAlign(tview.AlignLeft).
			SetDynamicColors(true)
		grid.AddItem(monthLabel, 0, startCol, 1, endCol-startCol+1, 0, 0, false)
	}

	return grid
}

func main() {
	// Determine the root path to scan. For now, use the current directory.
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Suppress console output when running in TUI mode
	// This is a temporary measure; a proper logger setup would be better.
	// For now, we'll redirect stdout/stderr to a null device or a log file
	// if we want to keep the TUI clean.
	// For development, keeping it for now to see scan progress.
	// fmt.Printf("Scanning for Git repositories in %q...\n", rootPath)

	repos, err := findGitRepos(rootPath)
	if err != nil {
		log.Fatalf("Failed to find Git repositories: %v", err)
	}

	app := tview.NewApplication()
	var rootPrimitive tview.Primitive

	if len(repos) == 0 {
		rootPrimitive = tview.NewTextView().
			SetText("No Git repositories found in the current directory or its subdirectories.").
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true)
	} else {
		// fmt.Printf("Found %d Git repositories:\n", len(repos))
		// for _, repo := range repos {
		// 	fmt.Printf("- %s\n", repo)
		// }

		var allCommits []time.Time
		for _, repoPath := range repos {
			// fmt.Printf("Collecting commits from %q...\n", repoPath)
			commits, err := getCommits(repoPath)
			if err != nil {
				log.Printf("Error collecting commits from %q: %v", repoPath, err)
				continue // Continue to the next repo even if one fails
			}
			allCommits = append(allCommits, commits...)
			// fmt.Printf("  Collected %d commits.\n", len(commits))
		}

		if len(allCommits) == 0 {
			rootPrimitive = tview.NewTextView().
				SetText("No commits found in any of the discovered Git repositories for the last year.").
				SetTextAlign(tview.AlignCenter).
				SetDynamicColors(true)
		} else {
			// fmt.Printf("\nAggregating %d total commits...\n", len(allCommits))
			contributionGraph := aggregateCommits(allCommits)
			rootPrimitive = renderContributionGraph(contributionGraph)
		}
	}

	if err := app.SetRoot(rootPrimitive, true).Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
