package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/genai"
)

// AIAnalysisResult represents the structured JSON response we expect from Gemini.
type AIAnalysisResult struct {
	AutoUpdateRisk string `json:"AutoUpdateRisk"` // "Safe", "Caution", or "High Risk"
	ChangeSummary  string `json:"ChangeSummary"`  // Summary of changes across all versions
	SecurityReport string `json:"SecurityReport"` // Security assessment of the skipped versions
}

// AnalyzeReleaseNotes sends the compiled release history to Gemini and returns a structured risk assessment.
func AnalyzeReleaseNotes(ctx context.Context, updateCtx *GitHubUpdateContext) (*AIAnalysisResult, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Gemini client")
	}

	// 1. Compile the release notes into a single string
	var releaseNotesBuilder strings.Builder
	for _, release := range updateCtx.Releases {
		fmt.Fprintf(&releaseNotesBuilder, "### Release: %s\n", release.TagName)
		releaseNotesBuilder.WriteString(release.Body + "\n\n")
	}

	// 2. Construct the System Prompt
	prompt := fmt.Sprintf(`You are an expert DevOps and Homelab systems administrator.
		I am providing you with the application description and a list of ALL release notes published since 
		my currently installed version.

		Application Description: %s

		Your task is to analyze these release notes collectively and identify:
		1. The risk of automatically applying all these updates at once.
		2. A summary of the major changes across all provided releases.
		3. A security report indicating if any of these newer releases patch critical vulnerabilities present in my older version.

		You must respond in a valid JSON object matching this schema exactly:
		{
		  "AutoUpdateRisk": "Safe" | "Caution" | "High Risk",
		  "ChangeSummary": "A concise summary of the major changes and breaking changes across all versions.",
		  "SecurityReport": "A report on whether these updates fix any critical security issues or CVEs."
		}

		Categorize AutoUpdateRisk as:
		- 'Safe' if it is just bug fixes, docs, or safe feature additions across all versions.
		- 'Caution' if there are minor configuration tweaks needed or deprecation notices.
		- 'High Risk' if there are explicit breaking changes, manual database migrations, or major architectural shifts.

		Here are the release notes since my current version:
		%s`,
		updateCtx.Description, releaseNotesBuilder.String())

	// 3. Call the API
	// Force the AI to respond in pure JSON
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate content from Gemini")
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("Gemini returned an empty response")
	}

	// 4. Extract and parse the JSON
	var result AIAnalysisResult
	part := resp.Candidates[0].Content.Parts[0]

	if part.Text != "" {
		jsonStr := part.Text
		// Clean up potential markdown formatting like ```json ... ``` that Gemini might still wrap it in
		jsonStr = strings.TrimPrefix(jsonStr, "```json\n")
		jsonStr = strings.TrimSuffix(jsonStr, "\n```")
		jsonStr = strings.TrimSuffix(jsonStr, "```") // sometimes it misses newline

		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal Gemini JSON output")
		}
		return &result, nil
	}

	return nil, errors.New("unexpected response format from Gemini")
}
