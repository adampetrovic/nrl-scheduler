import React from 'react';
import { ChevronLeftIcon, ChevronRightIcon } from '@heroicons/react/24/outline';
import { Match } from '../../types';
import { MatchCard } from './MatchCard';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { useUIStore } from '../../store/uiStore';

interface DrawGridProps {
  matches: Match[];
  totalRounds: number;
  showConstraintViolations?: boolean;
  onMatchClick?: (match: Match) => void;
}

export const DrawGrid: React.FC<DrawGridProps> = ({
  matches,
  totalRounds,
  showConstraintViolations = true,
  onMatchClick,
}) => {
  const { selectedRound, setSelectedRound } = useUIStore();

  // Group matches by round
  const matchesByRound = matches.reduce((acc, match) => {
    if (!acc[match.round]) {
      acc[match.round] = [];
    }
    acc[match.round].push(match);
    return acc;
  }, {} as Record<number, Match[]>);

  const currentRoundMatches = matchesByRound[selectedRound] || [];
  
  const handlePreviousRound = () => {
    if (selectedRound > 1) {
      setSelectedRound(selectedRound - 1);
    }
  };

  const handleNextRound = () => {
    if (selectedRound < totalRounds) {
      setSelectedRound(selectedRound + 1);
    }
  };

  const getRoundStats = (round: number) => {
    const roundMatches = matchesByRound[round] || [];
    const totalMatches = roundMatches.length;
    const byeMatches = roundMatches.filter(m => m.is_bye).length;
    const regularMatches = totalMatches - byeMatches;
    
    return { totalMatches, byeMatches, regularMatches };
  };

  const currentStats = getRoundStats(selectedRound);

  return (
    <div className="space-y-6">
      {/* Round Navigation */}
      <Card>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              variant="outline"
              size="sm"
              onClick={handlePreviousRound}
              disabled={selectedRound <= 1}
              className="flex items-center space-x-1"
            >
              <ChevronLeftIcon className="h-4 w-4" />
              <span>Previous</span>
            </Button>
            
            <div className="text-center">
              <h2 className="text-xl font-bold text-gray-900">
                Round {selectedRound}
              </h2>
              <p className="text-sm text-gray-600">
                {currentStats.regularMatches} matches, {currentStats.byeMatches} byes
              </p>
            </div>
            
            <Button
              variant="outline"
              size="sm"
              onClick={handleNextRound}
              disabled={selectedRound >= totalRounds}
              className="flex items-center space-x-1"
            >
              <span>Next</span>
              <ChevronRightIcon className="h-4 w-4" />
            </Button>
          </div>
          
          <div className="flex items-center space-x-2">
            <select
              value={selectedRound}
              onChange={(e) => setSelectedRound(parseInt(e.target.value))}
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {Array.from({ length: totalRounds }, (_, i) => i + 1).map(round => (
                <option key={round} value={round}>
                  Round {round}
                </option>
              ))}
            </select>
          </div>
        </div>
      </Card>

      {/* Round Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="text-center">
          <div className="text-2xl font-bold text-blue-600">
            {currentStats.regularMatches}
          </div>
          <div className="text-sm text-gray-600">Regular Matches</div>
        </Card>
        
        <Card className="text-center">
          <div className="text-2xl font-bold text-gray-600">
            {currentStats.byeMatches}
          </div>
          <div className="text-sm text-gray-600">Bye Weeks</div>
        </Card>
        
        <Card className="text-center">
          <div className="text-2xl font-bold text-green-600">
            {currentStats.totalMatches}
          </div>
          <div className="text-sm text-gray-600">Total Fixtures</div>
        </Card>
      </div>

      {/* Matches Grid */}
      {currentRoundMatches.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {currentRoundMatches
            .sort((a, b) => {
              // Sort by date, then by venue, then by team names
              const dateA = new Date(a.date).getTime();
              const dateB = new Date(b.date).getTime();
              if (dateA !== dateB) return dateA - dateB;
              
              const venueA = a.venue?.name || '';
              const venueB = b.venue?.name || '';
              if (venueA !== venueB) return venueA.localeCompare(venueB);
              
              const teamA = a.home_team?.name || '';
              const teamB = b.home_team?.name || '';
              return teamA.localeCompare(teamB);
            })
            .map((match) => (
              <MatchCard
                key={match.id}
                match={match}
                showConstraintViolations={showConstraintViolations}
                violations={[]} // TODO: Add actual constraint violations
                onClick={() => onMatchClick?.(match)}
              />
            ))}
        </div>
      ) : (
        <Card>
          <div className="text-center py-12">
            <p className="text-gray-500 text-lg">
              No matches scheduled for Round {selectedRound}
            </p>
            <p className="text-gray-400 text-sm mt-2">
              Generate a draw to see matches here
            </p>
          </div>
        </Card>
      )}
    </div>
  );
};