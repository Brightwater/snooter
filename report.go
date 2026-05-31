package main

import (
	"fmt"
	"strings"
)

// UpdateReport compiles the application details, version history, and AI analysis into a single cohesive payload.
type UpdateReport struct {
	AppName        string
	CurrentVersion string
	NewVersion     string
	ReleaseURL     string // URL to the GitHub releases page
	Analysis       *AIAnalysisResult
	UpdateMethod   string // e.g., "Docker Compose (Internal)", "Docker Compose (External)"
}

// ToMarkdown formats the UpdateReport into a clean Markdown string suitable for Discord or console output.
func (r *UpdateReport) ToMarkdown() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## 🚀 Update Available: %s\n\n", r.AppName)

	// Handle cases where the app is brand new and has no "current" version tracked yet
	if r.CurrentVersion == "" {
		fmt.Fprintf(&sb, "**Version:** `New Install` ➡️ `%s`\n", r.NewVersion)
	} else {
		fmt.Fprintf(&sb, "**Version:** `%s` ➡️ `%s`\n", r.CurrentVersion, r.NewVersion)
	}

	fmt.Fprintf(&sb, "**Update Method:** %s\n", r.UpdateMethod)

	if r.ReleaseURL != "" {
		fmt.Fprintf(&sb, "**Release Notes:** [View on GitHub](%s)\n", r.ReleaseURL)
	}
	fmt.Fprintf(&sb, "\n")

	if r.Analysis != nil {
		riskEmoji := "🟢"
		lowerRisk := strings.ToLower(r.Analysis.AutoUpdateRisk)
		if strings.Contains(lowerRisk, "caution") {
			riskEmoji = "🟡"
		} else if strings.Contains(lowerRisk, "high") || strings.Contains(lowerRisk, "critical") {
			riskEmoji = "🔴"
		}

		fmt.Fprintf(&sb, "### %s Risk Level: %s\n\n", riskEmoji, r.Analysis.AutoUpdateRisk)
		fmt.Fprintf(&sb, "**📝 AI Summary:**\n%s\n\n", r.Analysis.ChangeSummary)
		fmt.Fprintf(&sb, "**🛡️ Security Report:**\n%s\n", r.Analysis.SecurityReport)
	}

	return sb.String()
}
