CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    first_name TEXT,
    last_name TEXT,
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

CREATE TABLE user_sessions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' ||
         substr(hex(randomblob(2)), 2) || '-' ||
         substr('89ab', abs(random()) % 4 + 1, 1) ||
         substr(hex(randomblob(2)), 2) || '-' ||
         hex(randomblob(6)))),
    user_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    user_agent TEXT,
    ip_address TEXT,
    is_valid INTEGER DEFAULT 1, -- BOOLEAN as INTEGER
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);