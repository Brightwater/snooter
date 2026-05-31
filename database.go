package main

import (
	"database/sql"
	"embed"
	"os"
	"path/filepath"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// InitDB initializes the SQLite database, creating the file and schema if they don't exist.
func InitDB(dbPath string) (*sql.DB, error) {
	// Ensure the directory holding the DB exists (e.g., ./data/)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create database directory")
	}

	// Append PRAGMAs to enable Write-Ahead Logging (WAL) for concurrent reads/writes,
	// set a busy timeout to prevent "database is locked" errors, and enforce foreign keys.
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open sqlite database")
	}

	// Verify the connection is actually valid
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping sqlite database")
	}

	if err := createTables(db); err != nil {
		return nil, errors.Wrap(err, "failed to initialize database schema")
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	appMetadataTable, err := sqlFiles.ReadFile("sql/app_metadata.sql")
	if err != nil {
		return errors.Wrap(err, "failed to read app_metadata.sql")
	}

	eventStateTable, err := sqlFiles.ReadFile("sql/event_state.sql")
	if err != nil {
		return errors.Wrap(err, "failed to read event_state.sql")
	}

	if _, err := db.Exec(string(appMetadataTable)); err != nil {
		return errors.Wrap(err, "failed to create app_metadata table")
	}

	if _, err := db.Exec(string(eventStateTable)); err != nil {
		return errors.Wrap(err, "failed to create event_state table")
	}

	return nil
}

// SyncAppMetadata ensures all deployments defined in snooter.yaml exist in the SQLite database,
// and removes any that are no longer present in the configuration.
func SyncAppMetadata(db *sql.DB, config *SnooterConfig) error {
	// Step 1: Upsert all apps from the config and build a map of their names for quick lookups.
	upsertQuery := `
		INSERT INTO app_metadata (app_name, current_version, deployment_path)
		VALUES (?, 'Unknown', ?)
		ON CONFLICT(app_name) DO UPDATE SET deployment_path = excluded.deployment_path;`

	configAppNames := make(map[string]struct{})

	for _, dw := range config.Deployments {
		var path string
		if dcd, ok := dw.Deployment.(DockerComposeDeployment); ok {
			path = dcd.ComposePath
		}
		appName := dw.Deployment.GetName()
		if _, err := db.Exec(upsertQuery, appName, path); err != nil {
			return errors.Wrapf(err, "failed to sync metadata for app: %s", appName)
		}
		configAppNames[appName] = struct{}{}
	}

	// Step 2: Fetch all app names currently stored in the database.
	rows, err := db.Query(`SELECT app_name FROM app_metadata`)
	if err != nil {
		return errors.Wrap(err, "failed to query existing app names from database")
	}
	defer rows.Close()

	// Step 3: Iterate through the database apps and delete any that are not in the current config.
	deleteQuery := `DELETE FROM app_metadata WHERE app_name = ?`
	for rows.Next() {
		var dbAppName string
		if err := rows.Scan(&dbAppName); err != nil {
			return errors.Wrap(err, "failed to scan app name from database row")
		}

		// If the app from the DB doesn't exist in our config map, delete it.
		if _, exists := configAppNames[dbAppName]; !exists {
			if _, err := db.Exec(deleteQuery, dbAppName); err != nil {
				return errors.Wrapf(err, "failed to prune removed app: %s", dbAppName)
			}
		}
	}
	// Check for errors during row iteration.
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "error iterating over database app names")
	}

	return nil
}
