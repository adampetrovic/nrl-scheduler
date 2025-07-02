package optimizer

import (
	"errors"
	"math/rand"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// swapMatches swaps two matches between different rounds
func (sa *SimulatedAnnealing) swapMatches(draw *models.Draw) error {
	if len(draw.Matches) < 2 {
		return errors.New("not enough matches to swap")
	}
	
	// Find two different matches from different rounds
	var match1, match2 *models.Match
	maxAttempts := 50
	
	for attempts := 0; attempts < maxAttempts; attempts++ {
		idx1 := rand.Intn(len(draw.Matches))
		idx2 := rand.Intn(len(draw.Matches))
		
		if idx1 == idx2 {
			continue
		}
		
		match1 = draw.Matches[idx1]
		match2 = draw.Matches[idx2]
		
		// Only swap if they're in different rounds and both are regular matches (not byes)
		if match1.Round != match2.Round && !match1.IsBye() && !match2.IsBye() {
			break
		}
		
		match1, match2 = nil, nil
	}
	
	if match1 == nil || match2 == nil {
		return errors.New("could not find suitable matches to swap")
	}
	
	// Swap the rounds
	match1.Round, match2.Round = match2.Round, match1.Round
	
	return nil
}

// rescheduleMatch moves a match to a different round
func (sa *SimulatedAnnealing) rescheduleMatch(draw *models.Draw) error {
	if len(draw.Matches) == 0 {
		return errors.New("no matches to reschedule")
	}
	
	// Find a regular match (not a bye)
	var targetMatch *models.Match
	maxAttempts := 50
	
	for attempts := 0; attempts < maxAttempts; attempts++ {
		idx := rand.Intn(len(draw.Matches))
		match := draw.Matches[idx]
		
		if !match.IsBye() {
			targetMatch = match
			break
		}
	}
	
	if targetMatch == nil {
		return errors.New("could not find a regular match to reschedule")
	}
	
	// Choose a new round (different from current)
	originalRound := targetMatch.Round
	newRound := rand.Intn(draw.Rounds) + 1
	
	// Ensure it's different from the current round
	for newRound == originalRound {
		newRound = rand.Intn(draw.Rounds) + 1
	}
	
	targetMatch.Round = newRound
	
	return nil
}

// swapVenues changes venue assignments between two matches
func (sa *SimulatedAnnealing) swapVenues(draw *models.Draw) error {
	// Find two matches with venues that can be swapped
	var match1, match2 *models.Match
	maxAttempts := 50
	
	for attempts := 0; attempts < maxAttempts; attempts++ {
		idx1 := rand.Intn(len(draw.Matches))
		idx2 := rand.Intn(len(draw.Matches))
		
		if idx1 == idx2 {
			continue
		}
		
		m1 := draw.Matches[idx1]
		m2 := draw.Matches[idx2]
		
		// Both matches must have venues and not be byes
		if m1.VenueID != nil && m2.VenueID != nil && !m1.IsBye() && !m2.IsBye() {
			match1 = m1
			match2 = m2
			break
		}
	}
	
	if match1 == nil || match2 == nil {
		return errors.New("could not find suitable matches with venues to swap")
	}
	
	// Swap the venues
	match1.VenueID, match2.VenueID = match2.VenueID, match1.VenueID
	
	return nil
}

// swapHomeAway flips home/away designation for a match
func (sa *SimulatedAnnealing) swapHomeAway(draw *models.Draw) error {
	if len(draw.Matches) == 0 {
		return errors.New("no matches to modify")
	}
	
	// Find a regular match (not a bye)
	var targetMatch *models.Match
	maxAttempts := 50
	
	for attempts := 0; attempts < maxAttempts; attempts++ {
		idx := rand.Intn(len(draw.Matches))
		match := draw.Matches[idx]
		
		if !match.IsBye() && match.HomeTeamID != nil && match.AwayTeamID != nil {
			targetMatch = match
			break
		}
	}
	
	if targetMatch == nil {
		return errors.New("could not find a regular match to swap home/away")
	}
	
	// Swap home and away teams
	targetMatch.HomeTeamID, targetMatch.AwayTeamID = targetMatch.AwayTeamID, targetMatch.HomeTeamID
	
	return nil
}

// validateOperation checks if an operation maintains draw consistency
func (sa *SimulatedAnnealing) validateOperation(draw *models.Draw) error {
	// Check that all matches are still valid
	for _, match := range draw.Matches {
		if err := match.Validate(); err != nil {
			return err
		}
	}
	
	// Check that hard constraints are not violated
	violations := sa.ConstraintEngine.ValidateDraw(draw)
	if len(violations) > 0 {
		return errors.New("operation would violate hard constraints")
	}
	
	return nil
}

// applyMultipleOperations applies several operations in sequence
func (sa *SimulatedAnnealing) applyMultipleOperations(draw *models.Draw, count int) error {
	operations := []func(*models.Draw) error{
		sa.swapMatches,
		sa.rescheduleMatch,
		sa.swapVenues,
		sa.swapHomeAway,
	}
	
	for i := 0; i < count; i++ {
		operation := operations[rand.Intn(len(operations))]
		if err := operation(draw); err != nil {
			// If operation fails, continue with next one
			continue
		}
	}
	
	return nil
}

// revertLastOperation provides a mechanism to undo the last operation
// This is useful for more sophisticated optimization strategies
func (sa *SimulatedAnnealing) revertLastOperation(originalDraw, modifiedDraw *models.Draw) {
	// This is a simple implementation that just copies the original back
	// In a more sophisticated implementation, we might track specific operations
	*modifiedDraw = *sa.copyDraw(originalDraw)
}

// getRandomMatch returns a random match from the draw
func (sa *SimulatedAnnealing) getRandomMatch(draw *models.Draw) (*models.Match, error) {
	if len(draw.Matches) == 0 {
		return nil, errors.New("no matches available")
	}
	
	idx := rand.Intn(len(draw.Matches))
	return draw.Matches[idx], nil
}

// getRandomRegularMatch returns a random match that is not a bye
func (sa *SimulatedAnnealing) getRandomRegularMatch(draw *models.Draw) (*models.Match, error) {
	regularMatches := make([]*models.Match, 0)
	
	for _, match := range draw.Matches {
		if !match.IsBye() {
			regularMatches = append(regularMatches, match)
		}
	}
	
	if len(regularMatches) == 0 {
		return nil, errors.New("no regular matches available")
	}
	
	idx := rand.Intn(len(regularMatches))
	return regularMatches[idx], nil
}

// getMatchesByRound returns all matches in a specific round
func (sa *SimulatedAnnealing) getMatchesByRound(draw *models.Draw, round int) []*models.Match {
	matches := make([]*models.Match, 0)
	
	for _, match := range draw.Matches {
		if match.Round == round {
			matches = append(matches, match)
		}
	}
	
	return matches
}

// countMatchesInRound returns the number of matches in a specific round
func (sa *SimulatedAnnealing) countMatchesInRound(draw *models.Draw, round int) int {
	count := 0
	for _, match := range draw.Matches {
		if match.Round == round {
			count++
		}
	}
	return count
}