-- Drop triggers
DROP TRIGGER IF EXISTS update_matches_updated_at;
DROP TRIGGER IF EXISTS update_draws_updated_at;
DROP TRIGGER IF EXISTS update_venues_updated_at;
DROP TRIGGER IF EXISTS update_teams_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_matches_date;
DROP INDEX IF EXISTS idx_matches_away_team;
DROP INDEX IF EXISTS idx_matches_home_team;
DROP INDEX IF EXISTS idx_matches_round;
DROP INDEX IF EXISTS idx_matches_draw_id;

-- Drop tables
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS draws;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS venues;