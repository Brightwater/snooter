CREATE TABLE IF NOT EXISTS event_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_name TEXT NOT NULL,
    target_version TEXT NOT NULL,
    status TEXT NOT NULL,
    discord_message_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(app_name) REFERENCES app_metadata(app_name) ON DELETE CASCADE
);