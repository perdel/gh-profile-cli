package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rivo/tview"
)

// GitHubUser represents the structure of a GitHub user profile
type GitHubUser struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	PublicRepos int    `json:"public_repos"`
}

// fetchGitHubUser fetches the user data from GitHub API
func fetchGitHubUser(username string) (*GitHubUser, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-tui <GitHubUsername>")
		os.Exit(1)
	}

	username := os.Args[1]
	user, err := fetchGitHubUser(username)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	app := tview.NewApplication()

	textView := tview.NewTextView().
		SetText(fmt.Sprintf("GitHub User: %s\nName: %s\nFollowers: %d\nFollowing: %d\nPublic Repos: %d", user.Login, user.Name, user.Followers, user.Following, user.PublicRepos)).
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	if err := app.SetRoot(textView, true).Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}

}
