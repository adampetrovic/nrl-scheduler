package models

import (
	"testing"
	"time"
)

func TestMatch_Validate(t *testing.T) {
	tests := []struct {
		name    string
		match   Match
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid match",
			match: Match{
				DrawID:     1,
				Round:      1,
				HomeTeamID: intPtr(1),
				AwayTeamID: intPtr(2),
				VenueID:    intPtr(1),
			},
			wantErr: false,
		},
		{
			name: "valid bye",
			match: Match{
				DrawID:     1,
				Round:      1,
				HomeTeamID: nil,
				AwayTeamID: nil,
				VenueID:    nil,
			},
			wantErr: false,
		},
		{
			name: "no draw ID",
			match: Match{
				DrawID:     0,
				Round:      1,
				HomeTeamID: intPtr(1),
				AwayTeamID: intPtr(2),
				VenueID:    intPtr(1),
			},
			wantErr: true,
			errMsg:  "match must belong to a draw",
		},
		{
			name: "no round",
			match: Match{
				DrawID:     1,
				Round:      0,
				HomeTeamID: intPtr(1),
				AwayTeamID: intPtr(2),
				VenueID:    intPtr(1),
			},
			wantErr: true,
			errMsg:  "match round must be positive",
		},
		{
			name: "only home team",
			match: Match{
				DrawID:     1,
				Round:      1,
				HomeTeamID: intPtr(1),
				AwayTeamID: nil,
				VenueID:    intPtr(1),
			},
			wantErr: true,
			errMsg:  "match must have both home and away teams or be a bye",
		},
		{
			name: "only away team",
			match: Match{
				DrawID:     1,
				Round:      1,
				HomeTeamID: nil,
				AwayTeamID: intPtr(2),
				VenueID:    intPtr(1),
			},
			wantErr: true,
			errMsg:  "match must have both home and away teams or be a bye",
		},
		{
			name: "team playing itself",
			match: Match{
				DrawID:     1,
				Round:      1,
				HomeTeamID: intPtr(1),
				AwayTeamID: intPtr(1),
				VenueID:    intPtr(1),
			},
			wantErr: true,
			errMsg:  "team cannot play against itself",
		},
		{
			name: "match without venue",
			match: Match{
				DrawID:     1,
				Round:      1,
				HomeTeamID: intPtr(1),
				AwayTeamID: intPtr(2),
				VenueID:    nil,
			},
			wantErr: true,
			errMsg:  "match must have a venue",
		},
		{
			name: "match with date and time",
			match: Match{
				DrawID:      1,
				Round:       1,
				HomeTeamID:  intPtr(1),
				AwayTeamID:  intPtr(2),
				VenueID:     intPtr(1),
				MatchDate:   timePtr(time.Now()),
				MatchTime:   timePtr(time.Now()),
				IsPrimeTime: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.match.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Match.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Match.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestMatch_IsBye(t *testing.T) {
	tests := []struct {
		name  string
		match Match
		want  bool
	}{
		{
			name:  "bye match",
			match: Match{HomeTeamID: nil, AwayTeamID: nil},
			want:  true,
		},
		{
			name:  "regular match",
			match: Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			want:  false,
		},
		{
			name:  "invalid match with only home team",
			match: Match{HomeTeamID: intPtr(1), AwayTeamID: nil},
			want:  false,
		},
		{
			name:  "invalid match with only away team",
			match: Match{HomeTeamID: nil, AwayTeamID: intPtr(2)},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.match.IsBye(); got != tt.want {
				t.Errorf("Match.IsBye() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_HasTeam(t *testing.T) {
	tests := []struct {
		name   string
		match  Match
		teamID int
		want   bool
	}{
		{
			name:   "home team",
			match:  Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID: 1,
			want:   true,
		},
		{
			name:   "away team",
			match:  Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID: 2,
			want:   true,
		},
		{
			name:   "team not in match",
			match:  Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID: 3,
			want:   false,
		},
		{
			name:   "bye match",
			match:  Match{HomeTeamID: nil, AwayTeamID: nil},
			teamID: 1,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.match.HasTeam(tt.teamID); got != tt.want {
				t.Errorf("Match.HasTeam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_IsScheduled(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		match Match
		want  bool
	}{
		{
			name:  "scheduled match",
			match: Match{MatchDate: &now},
			want:  true,
		},
		{
			name:  "unscheduled match",
			match: Match{MatchDate: nil},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.match.IsScheduled(); got != tt.want {
				t.Errorf("Match.IsScheduled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_GetOpponent(t *testing.T) {
	tests := []struct {
		name      string
		match     Match
		teamID    int
		want      *int
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "get opponent for home team",
			match:   Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID:  1,
			want:    intPtr(2),
			wantErr: false,
		},
		{
			name:    "get opponent for away team",
			match:   Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID:  2,
			want:    intPtr(1),
			wantErr: false,
		},
		{
			name:    "team not in match",
			match:   Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID:  3,
			wantErr: true,
			errMsg:  "team not in this match",
		},
		{
			name:    "bye match",
			match:   Match{HomeTeamID: nil, AwayTeamID: nil},
			teamID:  1,
			wantErr: true,
			errMsg:  "bye matches have no opponent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.match.GetOpponent(tt.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Match.GetOpponent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Match.GetOpponent() error = %v, want %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && got != nil && tt.want != nil && *got != *tt.want {
				t.Errorf("Match.GetOpponent() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestMatch_IsHomeGame(t *testing.T) {
	tests := []struct {
		name    string
		match   Match
		teamID  int
		want    bool
		wantErr bool
		errMsg  string
	}{
		{
			name:    "home game",
			match:   Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID:  1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "away game",
			match:   Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID:  2,
			want:    false,
			wantErr: false,
		},
		{
			name:    "team not in match",
			match:   Match{HomeTeamID: intPtr(1), AwayTeamID: intPtr(2)},
			teamID:  3,
			want:    false,
			wantErr: true,
			errMsg:  "team not in this match",
		},
		{
			name:    "bye match",
			match:   Match{HomeTeamID: nil, AwayTeamID: nil},
			teamID:  1,
			want:    false,
			wantErr: true,
			errMsg:  "team not in this match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.match.IsHomeGame(tt.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Match.IsHomeGame() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Match.IsHomeGame() error = %v, want %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Match.IsHomeGame() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}