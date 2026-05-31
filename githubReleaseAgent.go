package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GitHubRepoInfo holds repository metadata like the description.
type GitHubRepoInfo struct {
	Description string `json:"description"`
}

// GitHubRelease represents the JSON response from the GitHub Releases API.
// We capture the tag (version), URL, and most importantly, the Body (release notes) for Gemini to analyze.
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// GitHubUpdateContext combines repository metadata and its release history into a single composed wrapper.
type GitHubUpdateContext struct {
	GitHubRepoInfo
	Releases []GitHubRelease
}

// getGitHubRepoInfo fetches the main description of the repository for AI context.
func getGitHubRepoInfo(repoURL string) (*GitHubRepoInfo, error) {
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid GitHub URL: %s", repoURL)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	var info GitHubRepoInfo
	if err := doGitHubRequest(apiURL, &info); err != nil {
		return nil, err
	}

	// Handle missing or empty descriptions gracefully
	if strings.TrimSpace(info.Description) == "" {
		info.Description = "No description provided."
	}

	return &info, nil
}

// getReleasesSince fetches all releases published after the specified currentVersion tag.
// If currentVersion is empty, it returns only the latest release.
func getReleasesSince(repoURL, currentVersion string) ([]GitHubRelease, error) {
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid GitHub URL: %s", repoURL)
	}

	// Fetch up to 30 recent releases (which covers almost all multi-version skip scenarios)
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=30", owner, repo)

	var allReleases []GitHubRelease
	if err := doGitHubRequest(apiURL, &allReleases); err != nil {
		return nil, err
	}

	if len(allReleases) == 0 {
		return nil, errors.Errorf("no releases found for repository: %s", repoURL)
	}

	var newReleases []GitHubRelease
	for _, r := range allReleases {
		// Handle releases that were published without any release notes
		if strings.TrimSpace(r.Body) == "" {
			r.Body = "No release notes provided for this version."
		}

		if currentVersion != "" && r.TagName == currentVersion {
			break // Reached the user's currently installed version
		}
		newReleases = append(newReleases, r)

		// If no current version was provided, we just want the latest one to establish a baseline
		if currentVersion == "" {
			break
		}
	}

	return newReleases, nil
}

// GetGitHubUpdateContext fetches both the repository info and the recent releases, returning them in a composed wrapper.
func GetGitHubUpdateContext(repoURL, currentVersion string) (*GitHubUpdateContext, error) {
	info, err := getGitHubRepoInfo(repoURL)
	if err != nil {
		return nil, err
	}

	releases, err := getReleasesSince(repoURL, currentVersion)
	if err != nil {
		return nil, err
	}

	return &GitHubUpdateContext{
		GitHubRepoInfo: *info,
		Releases:       releases,
	}, nil
}

// doGitHubRequest is a helper to execute GitHub API calls with authentication and JSON decoding.
func doGitHubRequest(apiURL string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	// Use the standard v3 API headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// If the user has provided a GitHub token in their .env, use it to avoid strict rate limits
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return errors.Wrap(err, "failed to parse GitHub response JSON")
	}

	return nil
}

// parseGitHubURL extracts the owner and repository name from a standard GitHub URL.
func parseGitHubURL(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if u.Host != "github.com" || len(parts) < 2 {
		return "", "", errors.New("URL is not a valid github.com repository")
	}

	return parts[0], parts[1], nil
}
