package main

import (
	"testing"
)

func Test_getGitHubRepoInfo(t *testing.T) {
	// Using a highly available public homelab repo: immich
	repoURL := "https://github.com/immich-app/immich"

	info, err := getGitHubRepoInfo(repoURL)
	if err != nil {
		t.Fatalf("Failed to fetch repo info: %v", err)
	}

	if info.Description == "" {
		t.Error("Expected a description, but got an empty string")
	}

	t.Logf("Fetched Description: %s", info.Description)
}

func Test_getReleasesSince(t *testing.T) {
	repoURL := "https://github.com/immich-app/immich"

	// Passing an empty string for currentVersion should fetch the latest release page
	releases, err := getReleasesSince(repoURL, "")
	if err != nil {
		t.Fatalf("Failed to fetch releases: %v", err)
	}

	if len(releases) == 0 {
		t.Fatalf("Expected at least one release, got 0")
	}

	latest := releases[0]
	if latest.TagName == "" || latest.Body == "" {
		t.Errorf("Release data is missing critical fields (TagName or Body): %+v", latest)
	}

	t.Logf("Successfully fetched %d releases. Latest tag: %s", len(releases), latest.TagName)
	t.Logf("Release Notes Snippet:\n%.200s...\n", latest.Body)
}

func TestGetGitHubUpdateContext(t *testing.T) {
	repoURL := "https://github.com/immich-app/immich"

	// Pass an empty string to just fetch the latest updates
	ctx, err := GetGitHubUpdateContext(repoURL, "")
	if err != nil {
		t.Fatalf("Failed to fetch update context: %v", err)
	}

	if ctx.Description == "" || len(ctx.Releases) == 0 {
		t.Errorf("Wrapper is missing data. Description: %s, Releases: %d", ctx.Description, len(ctx.Releases))
	}

	t.Logf("Wrapper successfully composed!\nRepo Description: %s\nLatest Release: %s", ctx.Description, ctx.Releases[0].TagName)
	t.Logf("Release Notes Snippet:\n%.200s...\n", ctx.Releases[0].Body)
}
