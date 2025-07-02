-- Venues table (created first as teams references it)
CREATE TABLE venues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    capacity INTEGER,
    latitude REAL,
    longitude REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Teams table
CREATE TABLE teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    short_name TEXT NOT NULL,
    city TEXT NOT NULL,
    venue_id INTEGER,
    latitude REAL,
    longitude REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (venue_id) REFERENCES venues(id)
);

-- Draws table
CREATE TABLE draws (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    season_year INTEGER NOT NULL,
    rounds INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft', 'optimizing', 'completed')),
    constraint_config TEXT, -- JSON configuration
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Matches table
CREATE TABLE matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draw_id INTEGER NOT NULL,
    round INTEGER NOT NULL,
    home_team_id INTEGER,
    away_team_id INTEGER,
    venue_id INTEGER,
    match_date DATE,
    match_time TIME,
    is_prime_time BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draw_id) REFERENCES draws(id) ON DELETE CASCADE,
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id),
    FOREIGN KEY (venue_id) REFERENCES venues(id),
    CHECK (home_team_id != away_team_id OR (home_team_id IS NULL AND away_team_id IS NULL))
);

-- Create indexes
CREATE INDEX idx_matches_draw_id ON matches(draw_id);
CREATE INDEX idx_matches_round ON matches(round);
CREATE INDEX idx_matches_home_team ON matches(home_team_id);
CREATE INDEX idx_matches_away_team ON matches(away_team_id);
CREATE INDEX idx_matches_date ON matches(match_date);

-- Create triggers for updated_at
CREATE TRIGGER update_teams_updated_at AFTER UPDATE ON teams
BEGIN
    UPDATE teams SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_venues_updated_at AFTER UPDATE ON venues
BEGIN
    UPDATE venues SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_draws_updated_at AFTER UPDATE ON draws
BEGIN
    UPDATE draws SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_matches_updated_at AFTER UPDATE ON matches
BEGIN
    UPDATE matches SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;