package draw

import (
	"errors"
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// Generator creates round-robin draws for sports competitions
type Generator struct {
	teams  []*models.Team
	rounds int
}

// NewGenerator creates a new draw generator
func NewGenerator(teams []*models.Team, rounds int) (*Generator, error) {
	if len(teams) < 2 {
		return nil, errors.New("need at least 2 teams to generate a draw")
	}
	if rounds < 1 {
		return nil, errors.New("rounds must be positive")
	}
	return &Generator{
		teams:  teams,
		rounds: rounds,
	}, nil
}

// GenerateRoundRobin creates a round-robin draw where each team plays each other team
func (g *Generator) GenerateRoundRobin() (*models.Draw, error) {
	numTeams := len(g.teams)
	isOdd := numTeams%2 == 1

	// For odd number of teams, add a virtual "bye" team
	workingTeams := make([]*models.Team, len(g.teams))
	copy(workingTeams, g.teams)
	
	if isOdd {
		workingTeams = append(workingTeams, nil) // nil represents bye
		numTeams++
	}

	draw := &models.Draw{
		Name:       fmt.Sprintf("Round Robin Draw - %d teams", len(g.teams)),
		SeasonYear: 2025, // Default, should be configurable
		Rounds:     g.rounds,
		Status:     models.DrawStatusDraft,
		Matches:    []*models.Match{},
	}

	// Calculate matches per round
	matchesPerRound := numTeams / 2
	
	// Calculate rounds needed for complete round-robin
	roundsInCycle := numTeams - 1
	if isOdd {
		roundsInCycle = numTeams - 1
	}

	// Standard round-robin algorithm using rotation
	for round := 1; round <= g.rounds; round++ {
		// Create matches for this round
		for match := 0; match < matchesPerRound; match++ {
			homeIdx := match
			awayIdx := numTeams - 1 - match

			homeTeam := workingTeams[homeIdx]
			awayTeam := workingTeams[awayIdx]

			// Skip if either team is bye (nil)
			if homeTeam == nil || awayTeam == nil {
				continue
			}

			// Determine home/away based on round for better balance
			// In even rounds, swap home/away for non-fixed matches
			actualHomeTeam := homeTeam
			actualAwayTeam := awayTeam
			
			// For the pairing involving the fixed team (index 0), alternate every cycle
			if homeIdx == 0 {
				cycleNum := ((round - 1) / roundsInCycle) % 2
				matchInCycle := (round - 1) % roundsInCycle
				if (matchInCycle % 2) == cycleNum {
					actualHomeTeam, actualAwayTeam = awayTeam, homeTeam
				}
			} else {
				// For other pairings, alternate each round
				if round % 2 == 0 {
					actualHomeTeam, actualAwayTeam = awayTeam, homeTeam
				}
			}

			matchModel := &models.Match{
				DrawID:     0, // Will be set when saved to DB
				Round:      round,
				HomeTeamID: &actualHomeTeam.ID,
				AwayTeamID: &actualAwayTeam.ID,
				VenueID:    actualHomeTeam.VenueID,
			}

			draw.Matches = append(draw.Matches, matchModel)
		}

		// Rotate teams for next round (keep first team fixed)
		g.rotateTeams(workingTeams)
	}

	return draw, nil
}

// rotateTeams performs the rotation for round-robin scheduling
// Keeps the first team fixed and rotates all others clockwise
func (g *Generator) rotateTeams(teams []*models.Team) {
	if len(teams) <= 2 {
		return
	}

	// Save the last team
	last := teams[len(teams)-1]

	// Shift all teams (except first) one position right
	for i := len(teams) - 1; i > 1; i-- {
		teams[i] = teams[i-1]
	}

	// Move the last team to position 1
	teams[1] = last
}

// GenerateDoubleRoundRobin creates a draw where each team plays each other twice (home and away)
func (g *Generator) GenerateDoubleRoundRobin() (*models.Draw, error) {
	// Calculate rounds needed for single round-robin
	singleRounds := len(g.teams) - 1
	if len(g.teams)%2 == 1 {
		singleRounds = len(g.teams)
	}

	// Create generator for single round-robin
	singleGen, err := NewGenerator(g.teams, singleRounds)
	if err != nil {
		return nil, err
	}

	// Generate first half
	draw, err := singleGen.GenerateRoundRobin()
	if err != nil {
		return nil, err
	}

	// Keep only matches from the single round-robin
	firstHalfMatches := make([]*models.Match, len(draw.Matches))
	copy(firstHalfMatches, draw.Matches)

	// Add reversed matches for second half
	for _, match := range firstHalfMatches {
		reversedMatch := &models.Match{
			DrawID:     match.DrawID,
			Round:      match.Round + singleRounds,
			HomeTeamID: match.AwayTeamID,
			AwayTeamID: match.HomeTeamID,
			VenueID:    nil, // Will be set based on new home team
		}

		// Set venue based on new home team
		if reversedMatch.HomeTeamID != nil {
			for _, team := range g.teams {
				if team.ID == *reversedMatch.HomeTeamID {
					reversedMatch.VenueID = team.VenueID
					break
				}
			}
		}

		draw.Matches = append(draw.Matches, reversedMatch)
	}

	draw.Name = fmt.Sprintf("Double Round Robin Draw - %d teams", len(g.teams))
	draw.Rounds = singleRounds * 2
	return draw, nil
}