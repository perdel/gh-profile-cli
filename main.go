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

func main() {
	// Determine the root path to scan. For now, use the current directory.
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	fmt.Printf("Scanning for Git repositories in %q...\n", rootPath)

	repos, err := findGitRepos(rootPath)
	if err != nil {
		log.Fatalf("Failed to find Git repositories: %v", err)
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		// Still run the TUI, but with a message
		app := tview.NewApplication()
		textView := tview.NewTextView().
			SetText("No Git repositories found in the current directory or its subdirectories.").
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true)
		if err := app.SetRoot(textView, true).Run(); err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
		return
	}

	fmt.Printf("Found %d Git repositories:\n", len(repos))
	for _, repo := range repos {
		fmt.Printf("- %s\n", repo)
	}

	var allCommits []time.Time
	for _, repoPath := range repos {
		fmt.Printf("Collecting commits from %q...\n", repoPath)
		commits, err := getCommits(repoPath)
		if err != nil {
			log.Printf("Error collecting commits from %q: %v", repoPath, err)
			continue // Continue to the next repo even if one fails
		}
		allCommits = append(allCommits, commits...)
		fmt.Printf("  Collected %d commits.\n", len(commits))
	}

	if len(allCommits) == 0 {
		fmt.Println("No commits found across all repositories.")
		app := tview.NewApplication()
		textView := tview.NewTextView().
			SetText("No commits found in any of the discovered Git repositories.").
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true)
		if err := app.SetRoot(textView, true).Run(); err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
		return
	}

	fmt.Printf("\nAggregating %d total commits...\n", len(allCommits))
	contributionGraph := aggregateCommits(allCommits)

	fmt.Println("\n--- Contribution Summary (Top 10 Days) ---")
	// Sort and print top 10 for verification
	type dailyCount struct {
		Date  string
		Count int
	}
	var sortedDays []dailyCount
	for date, count := range contributionGraph {
		sortedDays = append(sortedDays, dailyCount{Date: date, Count: count})
	}
	// Simple sort for demonstration, can be improved
	// For now, just print unsorted or a few to verify
	count := 0
	for date, commits := range contributionGraph {
		fmt.Printf("%s: %d commits\n", date, commits)
		count++
		if count >= 10 { // Print only first 10 for brevity
			break
		}
	}
	fmt.Println("------------------------------------------")

	// The TUI part will be implemented in the next step, using the 'contributionGraph' data.
	// For now, it just shows a placeholder message.
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetText("Data collected! Next step: Render the contribution graph using tview.").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	if err := app.SetRoot(textView, true).Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
