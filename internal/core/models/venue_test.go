package models

import (
	"testing"
)

func TestVenue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		venue   Venue
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid venue",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "Brisbane",
				Capacity:  52500,
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			venue: Venue{
				Name:      "",
				City:      "Brisbane",
				Capacity:  52500,
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "venue name cannot be empty",
		},
		{
			name: "empty city",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "",
				Capacity:  52500,
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "venue city cannot be empty",
		},
		{
			name: "negative capacity",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "Brisbane",
				Capacity:  -100,
				Latitude:  -27.4649,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "venue capacity cannot be negative",
		},
		{
			name: "invalid latitude too low",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "Brisbane",
				Capacity:  52500,
				Latitude:  -91,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "venue latitude must be between -90 and 90",
		},
		{
			name: "invalid latitude too high",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "Brisbane",
				Capacity:  52500,
				Latitude:  91,
				Longitude: 153.0095,
			},
			wantErr: true,
			errMsg:  "venue latitude must be between -90 and 90",
		},
		{
			name: "invalid longitude too low",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "Brisbane",
				Capacity:  52500,
				Latitude:  -27.4649,
				Longitude: -181,
			},
			wantErr: true,
			errMsg:  "venue longitude must be between -180 and 180",
		},
		{
			name: "invalid longitude too high",
			venue: Venue{
				Name:      "Suncorp Stadium",
				City:      "Brisbane",
				Capacity:  52500,
				Latitude:  -27.4649,
				Longitude: 181,
			},
			wantErr: true,
			errMsg:  "venue longitude must be between -180 and 180",
		},
		{
			name: "boundary latitude -90",
			venue: Venue{
				Name:      "South Pole Stadium",
				City:      "Antarctica",
				Capacity:  100,
				Latitude:  -90,
				Longitude: 0,
			},
			wantErr: false,
		},
		{
			name: "boundary latitude 90",
			venue: Venue{
				Name:      "North Pole Stadium",
				City:      "Arctic",
				Capacity:  100,
				Latitude:  90,
				Longitude: 0,
			},
			wantErr: false,
		},
		{
			name: "boundary longitude -180",
			venue: Venue{
				Name:      "Date Line Stadium",
				City:      "Pacific",
				Capacity:  100,
				Latitude:  0,
				Longitude: -180,
			},
			wantErr: false,
		},
		{
			name: "boundary longitude 180",
			venue: Venue{
				Name:      "Date Line Stadium",
				City:      "Pacific",
				Capacity:  100,
				Latitude:  0,
				Longitude: 180,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.venue.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Venue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Venue.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}