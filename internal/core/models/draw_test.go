package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDraw_Validate(t *testing.T) {
	tests := []struct {
		name    string
		draw    Draw
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid draw",
			draw: Draw{
				Name:       "NRL 2025 Season",
				SeasonYear: 2025,
				Rounds:     27,
				Status:     DrawStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			draw: Draw{
				Name:       "",
				SeasonYear: 2025,
				Rounds:     27,
				Status:     DrawStatusDraft,
			},
			wantErr: true,
			errMsg:  "draw name cannot be empty",
		},
		{
			name: "year too early",
			draw: Draw{
				Name:       "NRL 1999 Season",
				SeasonYear: 1999,
				Rounds:     27,
				Status:     DrawStatusDraft,
			},
			wantErr: true,
			errMsg:  "season year must be between 2000 and 2100",
		},
		{
			name: "year too late",
			draw: Draw{
				Name:       "NRL 2101 Season",
				SeasonYear: 2101,
				Rounds:     27,
				Status:     DrawStatusDraft,
			},
			wantErr: true,
			errMsg:  "season year must be between 2000 and 2100",
		},
		{
			name: "no rounds",
			draw: Draw{
				Name:       "NRL 2025 Season",
				SeasonYear: 2025,
				Rounds:     0,
				Status:     DrawStatusDraft,
			},
			wantErr: true,
			errMsg:  "rounds must be between 1 and 52",
		},
		{
			name: "too many rounds",
			draw: Draw{
				Name:       "NRL 2025 Season",
				SeasonYear: 2025,
				Rounds:     53,
				Status:     DrawStatusDraft,
			},
			wantErr: true,
			errMsg:  "rounds must be between 1 and 52",
		},
		{
			name: "invalid status",
			draw: Draw{
				Name:       "NRL 2025 Season",
				SeasonYear: 2025,
				Rounds:     27,
				Status:     "invalid",
			},
			wantErr: true,
			errMsg:  "invalid draw status",
		},
		{
			name: "optimizing status",
			draw: Draw{
				Name:       "NRL 2025 Season",
				SeasonYear: 2025,
				Rounds:     27,
				Status:     DrawStatusOptimizing,
			},
			wantErr: false,
		},
		{
			name: "completed status",
			draw: Draw{
				Name:       "NRL 2025 Season",
				SeasonYear: 2025,
				Rounds:     27,
				Status:     DrawStatusCompleted,
			},
			wantErr: false,
		},
		{
			name: "with constraint config",
			draw: Draw{
				Name:             "NRL 2025 Season",
				SeasonYear:       2025,
				Rounds:           27,
				Status:           DrawStatusDraft,
				ConstraintConfig: json.RawMessage(`{"max_consecutive_away": 3}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.draw.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Draw.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Draw.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDraw_GetMatchesByRound(t *testing.T) {
	now := time.Now()
	draw := Draw{
		Matches: []*Match{
			{ID: 1, Round: 1, HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			{ID: 2, Round: 1, HomeTeamID: intPtr(3), AwayTeamID: intPtr(4)},
			{ID: 3, Round: 2, HomeTeamID: intPtr(1), AwayTeamID: intPtr(3)},
			{ID: 4, Round: 2, HomeTeamID: intPtr(2), AwayTeamID: intPtr(4)},
			{ID: 5, Round: 3, HomeTeamID: intPtr(1), AwayTeamID: intPtr(4)},
		},
	}

	tests := []struct {
		name      string
		round     int
		wantCount int
		wantIDs   []int
	}{
		{
			name:      "round 1",
			round:     1,
			wantCount: 2,
			wantIDs:   []int{1, 2},
		},
		{
			name:      "round 2",
			round:     2,
			wantCount: 2,
			wantIDs:   []int{3, 4},
		},
		{
			name:      "round 3",
			round:     3,
			wantCount: 1,
			wantIDs:   []int{5},
		},
		{
			name:      "non-existent round",
			round:     99,
			wantCount: 0,
			wantIDs:   []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := draw.GetMatchesByRound(tt.round)
			if len(matches) != tt.wantCount {
				t.Errorf("GetMatchesByRound() returned %d matches, want %d", len(matches), tt.wantCount)
			}
			for i, match := range matches {
				if i < len(tt.wantIDs) && match.ID != tt.wantIDs[i] {
					t.Errorf("GetMatchesByRound() match[%d].ID = %d, want %d", i, match.ID, tt.wantIDs[i])
				}
			}
		})
	}

	// Test with scheduled dates
	draw.Matches[0].MatchDate = &now
	draw.Matches[2].MatchDate = &now
	matches := draw.GetMatchesByRound(1)
	if len(matches) != 2 {
		t.Errorf("GetMatchesByRound() with dates should still return all matches")
	}
}

func TestDraw_GetMatchesByTeam(t *testing.T) {
	draw := Draw{
		Matches: []*Match{
			{ID: 1, Round: 1, HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			{ID: 2, Round: 1, HomeTeamID: intPtr(3), AwayTeamID: intPtr(4)},
			{ID: 3, Round: 2, HomeTeamID: intPtr(1), AwayTeamID: intPtr(3)},
			{ID: 4, Round: 2, HomeTeamID: intPtr(2), AwayTeamID: intPtr(4)},
			{ID: 5, Round: 3, HomeTeamID: intPtr(4), AwayTeamID: intPtr(1)},
			{ID: 6, Round: 4}, // bye
		},
	}

	tests := []struct {
		name      string
		teamID    int
		wantCount int
		wantIDs   []int
	}{
		{
			name:      "team 1",
			teamID:    1,
			wantCount: 3,
			wantIDs:   []int{1, 3, 5},
		},
		{
			name:      "team 2",
			teamID:    2,
			wantCount: 2,
			wantIDs:   []int{1, 4},
		},
		{
			name:      "team 3",
			teamID:    3,
			wantCount: 2,
			wantIDs:   []int{2, 3},
		},
		{
			name:      "team 4",
			teamID:    4,
			wantCount: 3,
			wantIDs:   []int{2, 4, 5},
		},
		{
			name:      "non-existent team",
			teamID:    99,
			wantCount: 0,
			wantIDs:   []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := draw.GetMatchesByTeam(tt.teamID)
			if len(matches) != tt.wantCount {
				t.Errorf("GetMatchesByTeam() returned %d matches, want %d", len(matches), tt.wantCount)
			}
			for i, match := range matches {
				if i < len(tt.wantIDs) && match.ID != tt.wantIDs[i] {
					t.Errorf("GetMatchesByTeam() match[%d].ID = %d, want %d", i, match.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestDraw_IsComplete(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		draw     Draw
		complete bool
	}{
		{
			name:     "empty draw",
			draw:     Draw{},
			complete: false,
		},
		{
			name: "all matches scheduled",
			draw: Draw{
				Matches: []*Match{
					{ID: 1, MatchDate: &now},
					{ID: 2, MatchDate: &now},
					{ID: 3, MatchDate: &tomorrow},
				},
			},
			complete: true,
		},
		{
			name: "some matches unscheduled",
			draw: Draw{
				Matches: []*Match{
					{ID: 1, MatchDate: &now},
					{ID: 2, MatchDate: nil},
					{ID: 3, MatchDate: &tomorrow},
				},
			},
			complete: false,
		},
		{
			name: "all matches unscheduled",
			draw: Draw{
				Matches: []*Match{
					{ID: 1, MatchDate: nil},
					{ID: 2, MatchDate: nil},
					{ID: 3, MatchDate: nil},
				},
			},
			complete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.draw.IsComplete(); got != tt.complete {
				t.Errorf("Draw.IsComplete() = %v, want %v", got, tt.complete)
			}
		})
	}
}