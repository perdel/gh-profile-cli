package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// GitHubUser represents the structure of a GitHub user profile
type GitHubUser struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Followers   int    `json:"followers"`
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

	fmt.Printf("User: %s (%s)\nFollowers: %d\nPublic Repos: %d\n", user.Name, user.Login, user.Followers, user.PublicRepos)
}
