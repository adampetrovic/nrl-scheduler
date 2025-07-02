import React from 'react';
import { ExclamationTriangleIcon, MapPinIcon, CalendarIcon } from '@heroicons/react/24/outline';
import { Match } from '../../types';
import { Card } from '../ui/Card';

interface MatchCardProps {
  match: Match;
  showConstraintViolations?: boolean;
  violations?: string[];
  onClick?: () => void;
  className?: string;
}

export const MatchCard: React.FC<MatchCardProps> = ({
  match,
  showConstraintViolations = false,
  violations = [],
  onClick,
  className = '',
}) => {
  const hasViolations = violations.length > 0;
  
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

  if (match.is_bye) {
    return (
      <Card 
        className={`bg-gray-50 border-gray-200 ${className}`}
        onClick={onClick}
      >
        <div className="text-center py-2">
          <p className="text-sm font-medium text-gray-600">
            {match.home_team?.name || `Team ${match.home_team_id}`} - BYE
          </p>
          <p className="text-xs text-gray-500 mt-1">
            Round {match.round}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Card 
      className={`cursor-pointer transition-all hover:shadow-md ${
        hasViolations && showConstraintViolations 
          ? 'border-red-300 bg-red-50' 
          : 'hover:border-blue-300'
      } ${className}`}
      onClick={onClick}
    >
      <div className="space-y-2">
        {/* Teams */}
        <div className="text-center">
          <div className="space-y-1">
            <div className="flex items-center justify-center space-x-2">
              <span className="font-semibold text-gray-900">
                {match.home_team?.name || `Team ${match.home_team_id}`}
              </span>
              <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                HOME
              </span>
            </div>
            <div className="text-xs text-gray-500">vs</div>
            <div className="flex items-center justify-center space-x-2">
              <span className="font-semibold text-gray-900">
                {match.away_team?.name || `Team ${match.away_team_id}`}
              </span>
              <span className="text-xs bg-gray-100 text-gray-800 px-2 py-1 rounded">
                AWAY
              </span>
            </div>
          </div>
        </div>

        {/* Match Details */}
        <div className="space-y-1 pt-2 border-t border-gray-200">
          <div className="flex items-center justify-center space-x-3 text-xs text-gray-600">
            <div className="flex items-center space-x-1">
              <CalendarIcon className="h-3 w-3" />
              <span>{formatDate(match.date)}</span>
            </div>
            <div className="flex items-center space-x-1">
              <MapPinIcon className="h-3 w-3" />
              <span>{match.venue?.name || 'TBD'}</span>
            </div>
          </div>
          <div className="text-center">
            <span className="text-xs text-gray-500">Round {match.round}</span>
          </div>
        </div>

        {/* Constraint Violations */}
        {hasViolations && showConstraintViolations && (
          <div className="pt-2 border-t border-red-200">
            <div className="flex items-start space-x-1">
              <ExclamationTriangleIcon className="h-4 w-4 text-red-500 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <p className="text-xs text-red-700 font-medium mb-1">
                  Constraint Violations:
                </p>
                <ul className="text-xs text-red-600 space-y-0.5">
                  {violations.map((violation, index) => (
                    <li key={index}>• {violation}</li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        )}
      </div>
    </Card>
  );
};