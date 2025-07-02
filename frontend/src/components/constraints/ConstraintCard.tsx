import React from 'react';
import { TrashIcon, CogIcon } from '@heroicons/react/24/outline';
import { HardConstraintConfig, SoftConstraintConfig } from '../../types';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';

interface ConstraintCardProps {
  constraint: HardConstraintConfig | SoftConstraintConfig;
  type: 'hard' | 'soft';
  onEdit: () => void;
  onDelete: () => void;
}

export const ConstraintCard: React.FC<ConstraintCardProps> = ({
  constraint,
  type,
  onEdit,
  onDelete,
}) => {
  const isHard = type === 'hard';
  const softConstraint = constraint as SoftConstraintConfig;

  const getConstraintDisplayName = (constraintType: string): string => {
    const names: Record<string, string> = {
      venue_availability: 'Venue Availability',
      bye_constraint: 'Bye Distribution',
      team_availability: 'Team Availability',
      double_up_constraint: 'Double-Up Prevention',
      travel_minimization: 'Travel Minimization',
      rest_period: 'Rest Period',
      prime_time_spread: 'Prime Time Distribution',
      home_away_balance: 'Home/Away Balance',
    };
    return names[constraintType] || constraintType;
  };

  const getConstraintDescription = (constraint: HardConstraintConfig | SoftConstraintConfig): string => {
    const params = constraint.params;
    
    switch (constraint.type) {
      case 'venue_availability':
        return `Venue: ${params.venue}, Dates: ${params.dates?.join(', ') || 'None'}`;
      case 'team_availability':
        return `Team: ${params.team}, Dates: ${params.dates?.join(', ') || 'None'}`;
      case 'travel_minimization':
        return `Max consecutive away: ${params.max_consecutive_away || 3}`;
      case 'rest_period':
        return `Minimum days: ${params.minimum_days || 7}`;
      case 'prime_time_spread':
        return `Max per team: ${params.max_per_team || 2}`;
      case 'double_up_constraint':
        return `Rounds separation: ${params.rounds_separation || 5}`;
      default:
        return JSON.stringify(params);
    }
  };

  return (
    <Card className="mb-4">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <div className="flex items-center space-x-2">
            <span
              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                isHard
                  ? 'bg-red-100 text-red-800'
                  : 'bg-blue-100 text-blue-800'
              }`}
            >
              {isHard ? 'Hard' : 'Soft'}
            </span>
            <h4 className="text-lg font-medium text-gray-900">
              {getConstraintDisplayName(constraint.type)}
            </h4>
            {!isHard && (
              <span className="text-sm text-gray-500 font-medium">
                Weight: {softConstraint.weight}
              </span>
            )}
          </div>
          <p className="text-sm text-gray-600 mt-1">
            {getConstraintDescription(constraint)}
          </p>
        </div>
        
        <div className="flex items-center space-x-2 ml-4">
          <Button
            variant="outline"
            size="sm"
            onClick={onEdit}
            className="flex items-center space-x-1"
          >
            <CogIcon className="h-4 w-4" />
            <span>Edit</span>
          </Button>
          <Button
            variant="danger"
            size="sm"
            onClick={onDelete}
            className="flex items-center space-x-1"
          >
            <TrashIcon className="h-4 w-4" />
            <span>Delete</span>
          </Button>
        </div>
      </div>
    </Card>
  );
};