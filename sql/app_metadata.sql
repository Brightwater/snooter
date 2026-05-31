CREATE TABLE IF NOT EXISTS app_metadata (
    app_name TEXT PRIMARY KEY,
    current_version TEXT NOT NULL,
    deployment_path TEXT NOT NULL,
    last_checked DATETIME DEFAULT CURRENT_TIMESTAMP
);