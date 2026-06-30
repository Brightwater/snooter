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


## 5/30/26: Architectural Review (Code Audit)

Performed a full code audit via pair review. The overall architecture is sound — module
boundaries are intentional, the DB design is solid, and the cumulative release note pipeline
is correct. However, two real bugs were surfaced that need to be fixed before the cron/orchestrator
loop is wired in.

**Bug 1 — Data Race in `commandCapture.go`:**
Two goroutines (`stdout` and `stderr` readers) both write concurrently to the same `bytes.Buffer`
without synchronization. `bytes.Buffer` is not goroutine-safe. This is a real race condition that
will silently corrupt subprocess output under load. Fix: add a `sync.Mutex` around writes, use
`io.MultiWriter` to merge pipes before reading, or read them sequentially.

**Bug 2 — Type Assertion Leak in `database.go` (`SyncAppMetadata`):**
The function type-asserts `dw.Deployment.(DockerComposeDeployment)` to extract `ComposePath`
inside `database.go`. This punches through the `Deployment` interface abstraction — adding any
new deployment type (e.g., `ProxmoxDeployment`) will silently result in `path = ""` with no
error or warning. Fix: add `GetDeploymentPath() string` to the `Deployment` interface in
`types.go` and implement it on each concrete type.

**Discussion Point — `package main` Monolith:**
All ~12 source files live in `package main`. This is currently fine but will become painful when
the orchestrator, cron loop, and Discord interaction handlers land in the same package. At that
milestone, consider splitting into `internal/` sub-packages (e.g., `internal/github`,
`internal/docker`, `internal/ai`, `internal/discord`, `internal/db`) to enforce real API
boundaries and enable unit testing of internal helpers without making them public exports.
Not urgent — worth revisiting before wiring the orchestrator.
