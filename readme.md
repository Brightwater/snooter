Snooter: Intelligent Homelab Inventory & Update Manager

Snooter is a Go-based orchestration tool designed to move beyond "blind" automated updates (like Watchtower) by providing context-aware analysis of software releases. It monitors Docker Compose stacks, Proxmox nodes, and networking gear, using LLMs to identify breaking changes and security vulnerabilities before alerting the user via Discord.

## 📖 Methodology & Philosophy
Snooter is built specifically for users of standard Docker Compose. It is not meant to have a fancy Web UI or replace platforms like Portainer, TrueNAS Scale, or Arcane. It is a tool designed to help lazy but competent homelabbers who are comfortable with the Linux CLI keep their things up to date without manually reading release notes.
🎯 Project Objectives

    Centralized Inventory: Automatically discover and track versions of all self-hosted services.

    Semantic Analysis: Use Gemini to parse release notes for "breaking changes" and "manual migrations."

    Vulnerability Auditing: Cross-reference current versions against the OSV (Open Source Vulnerability) database.

    Interactive ChatOps: Approve or reject updates directly via Discord buttons.

    Zero-Cost Infrastructure: Utilize free-tier APIs (GitHub, OSV, Gemini AI Studio) for a sustainable $0/month operation.

🏗 System Architecture

The application is structured as a modular Go binary intended to run as a container with access to the host's Docker socket and local filesystem.
Component	Responsibility
Config Manager	Reads the central snooter.yaml config to inventory organized deployments.
Watcher	Polls GitHub Releases and Anitya for version increments.
Analyzer	Sends release notes to Gemini and queries OSV.dev for CVE data.
Data Store	SQLite database to track notification history, state, and ignore-lists. Schema documented in source (No ORM).
Discord Bot	Handles the WebSocket connection for real-time alerts and interaction.
🛠 Technical Stack

    Language: Go (Golang)

    Database: SQLite (pure-Go, WAL mode)

    Orchestration: Docker SDK / Compose Go SDK

    Scheduling: robfig/cron/v3

    AI Engine: Gemini 3 Flash / Gemini 2.5 Flash (via Google AI Studio)

    APIs: GitHub REST API, OSV.dev, Anitya

    Notifications: Discord (via discordgo)

📂 Expected Directory Structure

Snooter expects an organized, Arcane-style directory structure. Snooter runs from its own subfolder, but reads a central config from the root:

    /opt/stacks/
    ├── docker-compose.yml         # Root compose file for base infra (e.g., Arcane)
    ├── arcane-config/             # Arcane configuration directory
    ├── snooter/                   # Snooter's own application folder
    │   └── docker-compose.yml     # Snooter's container definition
    |   └── snooter.yaml           # Central configuration file tracking all deployments
    ├── caddy/                     # Individual project folder
    │   └── docker-compose.yml
    ├── frigate/                   # Individual project folder
    │   └── docker-compose.yml
    └── jellyfin/                  # Individual project folder
        └── docker-compose.yml

🔄 Logic Flow
1. Configuration & Inventory

The app reads `snooter.yaml` on startup to register all defined deployments (Docker Compose external, internal git-builds, etc.) and establishes their absolute paths.

    Note: Use Docker labels (e.g., com.snooter.repo=owner/repo) in Compose files to explicitly map images to GitHub repositories if auto-detection fails.

2. Update Detection

    Schedule: A Cron job triggers the Watcher on a user-defined schedule.

    Poll: Check the remote registry for a digest change or a new semantic version on GitHub for the tracked deployments.

    Fetch: Download the body text from the latest GitHub Release.

    Audit: Send the version string to OSV.dev to check for known vulnerabilities.

3. AI Synthesis

The release notes are passed to Gemini with a system prompt:

    "Analyze the following release notes. Identify if there are breaking changes to configuration, database schemas, or API endpoints. Categorize the update as 'Safe', 'Warning', or 'Critical'. Return result in structured JSON."

4. Discord Interaction

Snooter pushes an embed to Discord (one message per application to avoid message size limits and keep interactions organized):

    Title: [App Name] Update Available: v1.1.0 -> v1.2.0

    Security: List of active CVEs (if any).

    AI Summary: A 2-sentence breakdown of what changed and why it might break.

    Buttons / User Options: 
    - [🚀 Proceed] (Runs the update pipeline and replies with logs/summary)
    - [💤 Snooze Version] (Ignores this update until the *next* version is released)
    - [⏰ Snooze Report] (Reminds the user again next time the report runs)
    
    *Note: For apps that cannot be auto-updated (e.g. Proxmox), this functions purely as a notifier.*

🗺 Implementation Roadmap
Phase 1: The Foundation (Inventory & OSV)

    *MVP Scope: Focus on Docker Compose deployments defined in the central config. Support for docker compose pull/build/up. External Dockerfile builds are Post-MVP.*

    [x] Implement parser for the central `snooter.yaml` configuration.

    [x] Integrate Docker Compose CLI/SDK to execute deployments.

    [x] Set up SQLite database schema using raw SQL (No ORM) for version tracking.

    [ ] Implement OSV.dev API client for basic security flags.

Phase 2: Intelligence (GitHub & Gemini)

    [x] Implement GitHub Releases API client.

    [x] Create Gemini API wrapper with structured prompt logic.

Phase 3: ChatOps (Discord)

    [ ] Set up Discord bot and command handlers.

    [ ] Implement interactive buttons for container restarts/rebuilds.

    [ ] Add support for "External Providers" (Proxmox API / OPNsense SSH).

🛡 Security & Permissions

    Docker Socket: Container must mount `/var/run/docker.sock` for orchestration (Docker-out-of-Docker).

    Filesystem: Snooter must mount the target compose directories using the exact same absolute path as the host system to ensure volume mounts resolve correctly via the host Docker daemon.

    API Keys: Environment variables for GITHUB_TOKEN, GEMINI_API_KEY, and DISCORD_TOKEN.