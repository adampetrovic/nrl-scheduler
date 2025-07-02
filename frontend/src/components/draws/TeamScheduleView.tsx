import React from 'react';
import { HomeIcon, MapPinIcon } from '@heroicons/react/24/outline';
import { Match, Team } from '../../types';
import { Card } from '../ui/Card';
import { useUIStore } from '../../store/uiStore';

interface TeamScheduleViewProps {
  matches: Match[];
  teams: Team[];
  onMatchClick?: (match: Match) => void;
}

export const TeamScheduleView: React.FC<TeamScheduleViewProps> = ({
  matches,
  teams,
  onMatchClick,
}) => {
  const { selectedTeam, setSelectedTeam } = useUIStore();

  const selectedTeamData = teams.find(t => t.id === selectedTeam);
  
  // Get matches for selected team
  const teamMatches = matches
    .filter(match => 
      match.home_team_id === selectedTeam || 
      match.away_team_id === selectedTeam
    )
    .sort((a, b) => a.round - b.round);

  const getOpponent = (match: Match): Team | null => {
    if (match.is_bye) return null;
    
    const opponentId = match.home_team_id === selectedTeam 
      ? match.away_team_id 
      : match.home_team_id;
    
    return teams.find(t => t.id === opponentId) || null;
  };

  const isHomeGame = (match: Match): boolean => {
    return match.home_team_id === selectedTeam;
  };

  const formatDate = (dateString: string) => {
    try {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-AU', {
        weekday: 'short',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return dateString;
    }
  };

  const getScheduleStats = () => {
    const homeGames = teamMatches.filter(m => !m.is_bye && isHomeGame(m)).length;
    const awayGames = teamMatches.filter(m => !m.is_bye && !isHomeGame(m)).length;
    const byes = teamMatches.filter(m => m.is_bye).length;
    
    return { homeGames, awayGames, byes, total: teamMatches.length };
  };

  const stats = getScheduleStats();

  return (
    <div className="space-y-6">
      {/* Team Selection */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Select Team
            </label>
            <select
              value={selectedTeam || ''}
              onChange={(e) => setSelectedTeam(e.target.value ? parseInt(e.target.value) : null)}
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 min-w-64"
            >
              <option value="">Choose a team...</option>
              {teams
                .sort((a, b) => a.name.localeCompare(b.name))
                .map(team => (
                  <option key={team.id} value={team.id}>
                    {team.name}
                  </option>
                ))}
            </select>
          </div>
          
          {selectedTeamData && (
            <div className="text-right">
              <h3 className="text-lg font-semibold text-gray-900">
                {selectedTeamData.name}
              </h3>
              <p className="text-sm text-gray-600">
                {selectedTeamData.city} • {selectedTeamData.venue}
              </p>
            </div>
          )}
        </div>
      </Card>

      {selectedTeam && (
        <>
          {/* Schedule Stats */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {stats.homeGames}
              </div>
              <div className="text-sm text-gray-600">Home Games</div>
            </Card>
            
            <Card className="text-center">
              <div className="text-2xl font-bold text-orange-600">
                {stats.awayGames}
              </div>
              <div className="text-sm text-gray-600">Away Games</div>
            </Card>
            
            <Card className="text-center">
              <div className="text-2xl font-bold text-gray-600">
                {stats.byes}
              </div>
              <div className="text-sm text-gray-600">Bye Weeks</div>
            </Card>
            
            <Card className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {stats.total}
              </div>
              <div className="text-sm text-gray-600">Total Fixtures</div>
            </Card>
          </div>

          {/* Schedule List */}
          <Card title={`${selectedTeamData?.name} Schedule`}>
            {teamMatches.length > 0 ? (
              <div className="space-y-3">
                {teamMatches.map((match) => {
                  const opponent = getOpponent(match);
                  const isHome = isHomeGame(match);
                  
                  return (
                    <div
                      key={match.id}
                      onClick={() => onMatchClick?.(match)}
                      className="p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:shadow-md cursor-pointer transition-all"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          {/* Round */}
                          <div className="text-center">
                            <div className="text-lg font-bold text-gray-900">
                              {match.round}
                            </div>
                            <div className="text-xs text-gray-500">Round</div>
                          </div>
                          
                          {/* Match Details */}
                          <div className="flex-1">
                            {match.is_bye ? (
                              <div className="flex items-center space-x-2">
                                <span className="text-lg font-semibold text-gray-700">
                                  BYE WEEK
                                </span>
                              </div>
                            ) : (
                              <div className="space-y-1">
                                <div className="flex items-center space-x-2">
                                  <div className="flex items-center space-x-1">
                                    {isHome ? (
                                      <HomeIcon className="h-4 w-4 text-blue-600" />
                                    ) : (
                                      <MapPinIcon className="h-4 w-4 text-orange-600" />
                                    )}
                                    <span className={`text-xs font-medium px-2 py-1 rounded ${
                                      isHome 
                                        ? 'bg-blue-100 text-blue-800' 
                                        : 'bg-orange-100 text-orange-800'
                                    }`}>
                                      {isHome ? 'HOME' : 'AWAY'}
                                    </span>
                                  </div>
                                  <span className="text-sm text-gray-500">vs</span>
                                  <span className="text-lg font-semibold text-gray-900">
                                    {opponent?.name || 'TBD'}
                                  </span>
                                </div>
                                <div className="flex items-center space-x-4 text-sm text-gray-600">
                                  <span>{formatDate(match.date)}</span>
                                  <span>•</span>
                                  <span>{match.venue?.name || 'TBD'}</span>
                                </div>
                              </div>
                            )}
                          </div>
                        </div>
                        
                        {/* Additional Info */}
                        <div className="text-right text-sm text-gray-500">
                          {!match.is_bye && (
                            <div>
                              <div>{match.venue?.city}</div>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500">
                  No matches found for {selectedTeamData?.name}
                </p>
              </div>
            )}
          </Card>
        </>
      )}
    </div>
  );
};