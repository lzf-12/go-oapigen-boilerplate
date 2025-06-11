CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    is_active INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE auth_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL DEFAULT 'local', -- 'local', 'google', 'github', etc.
    provider_id TEXT,  -- social logins
    password_hash TEXT,  -- NULL for social logins
    UNIQUE(user_id, provider),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);