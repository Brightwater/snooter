package main

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestGenAIAnalyzer_Integration(t *testing.T) {
	// Load .env just in case the test is run directly without the main app loading it
	_ = godotenv.Load()

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping Gemini integration test: GEMINI_API_KEY is not set in environment or .env file")
	}

	// Let's use a popular repo and grab its latest updates
	repoURL := "https://github.com/immich-app/immich"

	t.Logf("Fetching GitHub releases for %s...", repoURL)
	updateCtx, err := GetGitHubUpdateContext(repoURL, "")
	if err != nil {
		t.Fatalf("Failed to fetch update context: %v", err)
	}

	t.Logf("Successfully fetched %d release(s). Sending to Gemini for analysis...", len(updateCtx.Releases))

	result, err := AnalyzeReleaseNotes(context.Background(), updateCtx)
	if err != nil {
		t.Fatalf("GenAIAnalyzer failed: %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result from GenAIAnalyzer")
	}

	t.Logf("\n=== 🤖 Gemini Analysis Result ===\n[Risk Level]: %s\n[Summary]: %s\n[Security]: %s\n=================================",
		result.AutoUpdateRisk, result.ChangeSummary, result.SecurityReport)
}
