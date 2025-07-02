package draw

import (
	"testing"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

func TestNewGenerator(t *testing.T) {
	teams := createTestTeams(4)

	tests := []struct {
		name    string
		teams   []*models.Team
		rounds  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid generator",
			teams:   teams,
			rounds:  3,
			wantErr: false,
		},
		{
			name:    "no teams",
			teams:   []*models.Team{},
			rounds:  3,
			wantErr: true,
			errMsg:  "need at least 2 teams to generate a draw",
		},
		{
			name:    "one team",
			teams:   teams[:1],
			rounds:  3,
			wantErr: true,
			errMsg:  "need at least 2 teams to generate a draw",
		},
		{
			name:    "zero rounds",
			teams:   teams,
			rounds:  0,
			wantErr: true,
			errMsg:  "rounds must be positive",
		},
		{
			name:    "negative rounds",
			teams:   teams,
			rounds:  -1,
			wantErr: true,
			errMsg:  "rounds must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.teams, tt.rounds)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("NewGenerator() error = %v, want %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && gen == nil {
				t.Error("NewGenerator() returned nil generator")
			}
		})
	}
}

func TestGenerateRoundRobin_EvenTeams(t *testing.T) {
	tests := []struct {
		name      string
		numTeams  int
		rounds    int
		wantStats drawStats
	}{
		{
			name:     "4 teams, 3 rounds",
			numTeams: 4,
			rounds:   3,
			wantStats: drawStats{
				totalMatches:   6, // 2 matches per round * 3 rounds
				matchesPerTeam: 3, // Each team plays once per round
				homeGamesMin:   0, // With only 3 rounds, some teams might not get home games
				homeGamesMax:   3, // Some teams might get all home games
			},
		},
		{
			name:     "6 teams, 5 rounds",
			numTeams: 6,
			rounds:   5,
			wantStats: drawStats{
				totalMatches:   15, // 3 matches per round * 5 rounds
				matchesPerTeam: 5,  // Each team plays once per round
				homeGamesMin:   1,  // With 5 rounds, minimum 1 home game
				homeGamesMax:   4,  // Maximum 4 home games
			},
		},
		{
			name:     "16 teams, full season (15 rounds)",
			numTeams: 16,
			rounds:   15,
			wantStats: drawStats{
				totalMatches:   120, // 8 matches per round * 15 rounds
				matchesPerTeam: 15,  // Each team plays once per round
				homeGamesMin:   6,   // With 15 rounds, should be fairly balanced
				homeGamesMax:   9,   // Allow some variance in home games
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams := createTestTeams(tt.numTeams)
			gen, err := NewGenerator(teams, tt.rounds)
			if err != nil {
				t.Fatalf("NewGenerator() error = %v", err)
			}

			draw, err := gen.GenerateRoundRobin()
			if err != nil {
				t.Fatalf("GenerateRoundRobin() error = %v", err)
			}

			// Verify total matches
			if len(draw.Matches) != tt.wantStats.totalMatches {
				t.Errorf("total matches = %d, want %d", len(draw.Matches), tt.wantStats.totalMatches)
			}

			// Verify each team plays correct number of matches
			teamMatchCounts := make(map[int]int)
			teamHomeGames := make(map[int]int)
			
			for _, match := range draw.Matches {
				if match.HomeTeamID != nil {
					teamMatchCounts[*match.HomeTeamID]++
					teamHomeGames[*match.HomeTeamID]++
				}
				if match.AwayTeamID != nil {
					teamMatchCounts[*match.AwayTeamID]++
				}
			}

			// Check match counts
			for teamID, count := range teamMatchCounts {
				if count != tt.wantStats.matchesPerTeam {
					t.Errorf("team %d played %d matches, want %d", teamID, count, tt.wantStats.matchesPerTeam)
				}
			}

			// Check home/away balance
			minHome, maxHome := tt.numTeams, 0
			for _, homeCount := range teamHomeGames {
				if homeCount < minHome {
					minHome = homeCount
				}
				if homeCount > maxHome {
					maxHome = homeCount
				}
			}

			if minHome < tt.wantStats.homeGamesMin {
				t.Errorf("minimum home games = %d, want >= %d", minHome, tt.wantStats.homeGamesMin)
			}
			if maxHome > tt.wantStats.homeGamesMax {
				t.Errorf("maximum home games = %d, want <= %d", maxHome, tt.wantStats.homeGamesMax)
			}

			// Verify no team plays itself
			for _, match := range draw.Matches {
				if match.HomeTeamID != nil && match.AwayTeamID != nil &&
					*match.HomeTeamID == *match.AwayTeamID {
					t.Errorf("team %d plays against itself", *match.HomeTeamID)
				}
			}

			// Verify no duplicate matches in same round
			roundMatches := make(map[int]map[string]bool)
			for _, match := range draw.Matches {
				if roundMatches[match.Round] == nil {
					roundMatches[match.Round] = make(map[string]bool)
				}
				
				if match.HomeTeamID != nil && match.AwayTeamID != nil {
					key1 := matchKey(*match.HomeTeamID, *match.AwayTeamID)
					key2 := matchKey(*match.AwayTeamID, *match.HomeTeamID)
					
					if roundMatches[match.Round][key1] || roundMatches[match.Round][key2] {
						t.Errorf("duplicate match in round %d: %s", match.Round, key1)
					}
					roundMatches[match.Round][key1] = true
					roundMatches[match.Round][key2] = true
				}
			}

			// Verify each team plays at most once per round
			for round := 1; round <= tt.rounds; round++ {
				teamAppearances := make(map[int]int)
				for _, match := range draw.Matches {
					if match.Round != round {
						continue
					}
					if match.HomeTeamID != nil {
						teamAppearances[*match.HomeTeamID]++
					}
					if match.AwayTeamID != nil {
						teamAppearances[*match.AwayTeamID]++
					}
				}
				
				for teamID, appearances := range teamAppearances {
					if appearances > 1 {
						t.Errorf("team %d appears %d times in round %d", teamID, appearances, round)
					}
				}
			}
		})
	}
}

func TestGenerateRoundRobin_OddTeams(t *testing.T) {
	tests := []struct {
		name        string
		numTeams    int
		rounds      int
		wantMatches int
	}{
		{
			name:        "3 teams, 3 rounds",
			numTeams:    3,
			rounds:      3,
			wantMatches: 3, // 1 match per round (1 team has bye)
		},
		{
			name:        "5 teams, 5 rounds",
			numTeams:    5,
			rounds:      5,
			wantMatches: 10, // 2 matches per round (1 team has bye)
		},
		{
			name:        "17 teams, 17 rounds",
			numTeams:    17,
			rounds:      17,
			wantMatches: 136, // 8 matches per round (1 team has bye)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams := createTestTeams(tt.numTeams)
			gen, err := NewGenerator(teams, tt.rounds)
			if err != nil {
				t.Fatalf("NewGenerator() error = %v", err)
			}

			draw, err := gen.GenerateRoundRobin()
			if err != nil {
				t.Fatalf("GenerateRoundRobin() error = %v", err)
			}

			// Verify total matches
			if len(draw.Matches) != tt.wantMatches {
				t.Errorf("total matches = %d, want %d", len(draw.Matches), tt.wantMatches)
			}

			// Verify each team has exactly one bye per round
			for round := 1; round <= tt.rounds; round++ {
				playingTeams := make(map[int]bool)
				
				for _, match := range draw.Matches {
					if match.Round != round {
						continue
					}
					if match.HomeTeamID != nil {
						playingTeams[*match.HomeTeamID] = true
					}
					if match.AwayTeamID != nil {
						playingTeams[*match.AwayTeamID] = true
					}
				}

				// Should have exactly numTeams-1 teams playing
				if len(playingTeams) != tt.numTeams-1 {
					t.Errorf("round %d has %d teams playing, want %d", 
						round, len(playingTeams), tt.numTeams-1)
				}

				// Find which team has bye
				byeTeam := -1
				for _, team := range teams {
					if !playingTeams[team.ID] {
						if byeTeam != -1 {
							t.Errorf("multiple teams have bye in round %d", round)
						}
						byeTeam = team.ID
					}
				}

				if byeTeam == -1 {
					t.Errorf("no team has bye in round %d", round)
				}
			}

			// Verify bye distribution is fair
			byeCounts := make(map[int]int)
			for round := 1; round <= tt.rounds; round++ {
				playingTeams := make(map[int]bool)
				
				for _, match := range draw.Matches {
					if match.Round != round {
						continue
					}
					if match.HomeTeamID != nil {
						playingTeams[*match.HomeTeamID] = true
					}
					if match.AwayTeamID != nil {
						playingTeams[*match.AwayTeamID] = true
					}
				}

				for _, team := range teams {
					if !playingTeams[team.ID] {
						byeCounts[team.ID]++
					}
				}
			}

			// Check bye fairness
			minByes, maxByes := tt.rounds, 0
			for _, count := range byeCounts {
				if count < minByes {
					minByes = count
				}
				if count > maxByes {
					maxByes = count
				}
			}

			// Byes should be distributed fairly (difference of at most 1)
			if maxByes-minByes > 1 {
				t.Errorf("unfair bye distribution: min=%d, max=%d", minByes, maxByes)
			}
		})
	}
}

func TestGenerateDoubleRoundRobin(t *testing.T) {
	tests := []struct {
		name          string
		numTeams      int
		wantMatches   int
		wantHomeGames int // per team
	}{
		{
			name:          "4 teams double round-robin",
			numTeams:      4,
			wantMatches:   12, // 6 matches in single RR * 2
			wantHomeGames: 3,  // Each team should have 3 home games
		},
		{
			name:          "6 teams double round-robin",
			numTeams:      6,
			wantMatches:   30, // 15 matches in single RR * 2
			wantHomeGames: 5,  // Each team should have 5 home games
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams := createTestTeams(tt.numTeams)
			singleRounds := tt.numTeams - 1
			gen, err := NewGenerator(teams, singleRounds*2)
			if err != nil {
				t.Fatalf("NewGenerator() error = %v", err)
			}

			draw, err := gen.GenerateDoubleRoundRobin()
			if err != nil {
				t.Fatalf("GenerateDoubleRoundRobin() error = %v", err)
			}

			// Verify total matches
			if len(draw.Matches) != tt.wantMatches {
				t.Errorf("total matches = %d, want %d", len(draw.Matches), tt.wantMatches)
			}

			// Verify each team plays each other exactly twice
			matchups := make(map[string]int)
			teamHomeGames := make(map[int]int)
			
			for _, match := range draw.Matches {
				if match.HomeTeamID != nil && match.AwayTeamID != nil {
					key := matchKey(*match.HomeTeamID, *match.AwayTeamID)
					matchups[key]++
					teamHomeGames[*match.HomeTeamID]++
				}
			}

			// Each ordered pair should appear exactly once
			for matchup, count := range matchups {
				if count != 1 {
					t.Errorf("matchup %s appears %d times, want 1", matchup, count)
				}
			}

			// Each team should have exactly the expected home games
			for teamID, homeGames := range teamHomeGames {
				if homeGames != tt.wantHomeGames {
					t.Errorf("team %d has %d home games, want %d", 
						teamID, homeGames, tt.wantHomeGames)
				}
			}
		})
	}
}

// Helper functions

type drawStats struct {
	totalMatches   int
	matchesPerTeam int
	homeGamesMin   int
	homeGamesMax   int
}

func createTestTeams(n int) []*models.Team {
	teams := make([]*models.Team, n)
	for i := 0; i < n; i++ {
		venueID := i + 1
		teams[i] = &models.Team{
			ID:        i + 1,
			Name:      teamName(i + 1),
			ShortName: teamShortName(i + 1),
			City:      "City",
			VenueID:   &venueID,
			Latitude:  -27.0 + float64(i)*0.1,
			Longitude: 153.0 + float64(i)*0.1,
		}
	}
	return teams
}

func teamName(id int) string {
	names := []string{
		"Broncos", "Cowboys", "Dolphins", "Titans",
		"Raiders", "Bulldogs", "Dragons", "Eels",
		"Sea Eagles", "Storm", "Knights", "Panthers",
		"Rabbitohs", "Roosters", "Sharks", "Warriors",
		"Tigers", "Team18", "Team19", "Team20",
	}
	if id <= len(names) {
		return names[id-1]
	}
	return "Team" + string(rune('A'+id-1))
}

func teamShortName(id int) string {
	shorts := []string{
		"BRI", "NQL", "DOL", "GLD",
		"CAN", "CBY", "SGI", "PAR",
		"MAN", "MEL", "NEW", "PEN",
		"SOU", "SYD", "CRO", "WAR",
		"WST", "T18", "T19", "T20",
	}
	if id <= len(shorts) {
		return shorts[id-1]
	}
	return "T" + string(rune('A'+id-1))
}

func matchKey(home, away int) string {
	return string(rune('A'+home-1)) + string(rune('A'+away-1))
}