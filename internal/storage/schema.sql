CREATE TABLE IF NOT EXISTS replays (
    id TEXT PRIMARY KEY,
    format TEXT NOT NULL,
    gen INTEGER,
    gametype TEXT,
    rated BOOLEAN DEFAULT FALSE,
    winner TEXT,
    upload_time INTEGER,
    views INTEGER,
    log_raw TEXT,
    scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    replay_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    name TEXT NOT NULL,
    rating INTEGER,
    FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS turns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    replay_id TEXT NOT NULL,
    turn_number INTEGER NOT NULL,
    timestamp INTEGER,
    FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE,
    UNIQUE(replay_id, turn_number)
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    turn_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    pokemon TEXT,
    target TEXT,
    move TEXT,
    hp TEXT,
    stat TEXT,
    amount INTEGER,
    detail TEXT,
    FOREIGN KEY (turn_id) REFERENCES turns(id) ON DELETE CASCADE
);
