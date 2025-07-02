package models

import (
	"testing"
)

func TestTeam_Validate(t *testing.T) {
	tests := []struct {
		name    string
		team    Team
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid team",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "BRI",
				City:      "Brisbane",
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			team: Team{
				Name:      "",
				ShortName: "BRI",
				City:      "Brisbane",
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "team name cannot be empty",
		},
		{
			name: "empty short name",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "",
				City:      "Brisbane",
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "team short name cannot be empty",
		},
		{
			name: "short name too long",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "BRIS",
				City:      "Brisbane",
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "team short name cannot be longer than 3 characters",
		},
		{
			name: "empty city",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "BRI",
				City:      "",
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "team city cannot be empty",
		},
		{
			name: "invalid latitude",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "BRI",
				City:      "Brisbane",
				Latitude:  -91,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "team latitude must be between -90 and 90",
		},
		{
			name: "invalid longitude",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "BRI",
				City:      "Brisbane",
				Latitude:  -27.4649,
				Longitude: 181,
			},
			wantErr: true,
			errMsg:  "team longitude must be between -180 and 180",
		},
		{
			name: "with venue ID",
			team: Team{
				Name:      "Brisbane Broncos",
				ShortName: "BRI",
				City:      "Brisbane",
				Latitude:  -27.4649,
				Longitude: 153.0095,
				VenueID:   intPtr(1),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.team.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Team.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Team.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTeam_HasBye(t *testing.T) {
	tests := []struct {
		name string
		team *Team
		want bool
	}{
		{
			name: "nil team",
			team: nil,
			want: true,
		},
		{
			name: "team with ID 0",
			team: &Team{ID: 0},
			want: true,
		},
		{
			name: "team with valid ID",
			team: &Team{ID: 1},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.team.HasBye(); got != tt.want {
				t.Errorf("Team.HasBye() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}