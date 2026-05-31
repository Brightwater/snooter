# Snooter Devlog

## 4/6/26: Architecture, Integrations, and State

**Accomplished Today:**
- **Docker Compose Engine:** (THIS WASN'T TODAY I BUILT THIS PART LIKE A YEAR AGO) Built the execution engine for both `external` (existing stacks) and `internal` (Git-cloned source builds) deployments, complete with real-time `stdout`/`stderr` streaming.
- **GitHub Agent:** Wrote an intelligent fetcher that pulls repository metadata and *cumulative* missed release notes between the user's current version and the latest version.
- **AI Analysis:** Integrated `gemini-2.5-flash` to parse the compiled release notes. Instructed the LLM to output strict JSON containing an Auto-Update Risk level, a change summary, and a security/CVE report.
- **Markdown Reporting:** Created a formatter to combine app data, GitHub links, and the AI analysis into clean Markdown ready for Discord.
- **SQLite State Management:** Initialized a pure-Go SQLite database with WAL mode for concurrency. Created `app_metadata` and `event_state` tables via `//go:embed` SQL files, and wired up the boot sequence to sync `snooter.yaml` deployments into the database.

**Next Steps:**
- [x] Build the Docker Daemon Poller to extract currently running image versions.
- [x] Update the DB Sync logic to prune removed apps.
- Set up the master Cron loop and Discord integration.

## 5/2/26:
