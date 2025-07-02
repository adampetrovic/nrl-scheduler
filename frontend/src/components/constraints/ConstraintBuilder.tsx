import React, { useState } from 'react';
import { PlusIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { Dialog } from '@headlessui/react';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { ConstraintCard } from './ConstraintCard';
import { ConstraintWeightSlider } from './ConstraintWeightSlider';
import { useDrawStore } from '../../store/drawStore';
import { useUIStore } from '../../store/uiStore';
import { HardConstraintConfig, SoftConstraintConfig } from '../../types';

const CONSTRAINT_TYPES = [
  {
    type: 'venue_availability',
    name: 'Venue Availability',
    description: 'Restrict venue usage on specific dates',
    category: 'availability',
    isHard: true,
  },
  {
    type: 'team_availability',
    name: 'Team Availability',
    description: 'Prevent teams from playing on specific dates',
    category: 'availability',
    isHard: true,
  },
  {
    type: 'bye_constraint',
    name: 'Bye Distribution',
    description: 'Ensure proper bye week distribution',
    category: 'distribution',
    isHard: true,
  },
  {
    type: 'double_up_constraint',
    name: 'Double-Up Prevention',
    description: 'Prevent teams from playing twice in short periods',
    category: 'scheduling',
    isHard: true,
  },
  {
    type: 'travel_minimization',
    name: 'Travel Minimization',
    description: 'Minimize consecutive away games',
    category: 'optimization',
    isHard: false,
  },
  {
    type: 'rest_period',
    name: 'Rest Period',
    description: 'Ensure minimum rest between matches',
    category: 'welfare',
    isHard: false,
  },
  {
    type: 'prime_time_spread',
    name: 'Prime Time Distribution',
    description: 'Distribute prime-time games fairly',
    category: 'broadcast',
    isHard: false,
  },
  {
    type: 'home_away_balance',
    name: 'Home/Away Balance',
    description: 'Balance home and away games per team',
    category: 'fairness',
    isHard: false,
  },
];

export const ConstraintBuilder: React.FC = () => {
  const { constraints, setConstraints } = useDrawStore();
  const { isConstraintBuilderOpen, closeConstraintBuilder } = useUIStore();
  const [selectedType, setSelectedType] = useState<string>('');
  const [isHardConstraint, setIsHardConstraint] = useState(true);
  const [weight, setWeight] = useState(0.5);
  const [params, setParams] = useState<Record<string, any>>({});

  const handleAddConstraint = () => {
    if (!selectedType) return;

    const constraintType = CONSTRAINT_TYPES.find(c => c.type === selectedType);
    if (!constraintType) return;

    if (constraintType.isHard || isHardConstraint) {
      const newConstraint: HardConstraintConfig = {
        type: selectedType,
        params,
      };
      setConstraints({
        ...constraints,
        hard: [...constraints.hard, newConstraint],
      });
    } else {
      const newConstraint: SoftConstraintConfig = {
        type: selectedType,
        weight,
        params,
      };
      setConstraints({
        ...constraints,
        soft: [...constraints.soft, newConstraint],
      });
    }

    // Reset form
    setSelectedType('');
    setParams({});
    setWeight(0.5);
  };

  const handleDeleteConstraint = (type: 'hard' | 'soft', index: number) => {
    if (type === 'hard') {
      setConstraints({
        ...constraints,
        hard: constraints.hard.filter((_, i) => i !== index),
      });
    } else {
      setConstraints({
        ...constraints,
        soft: constraints.soft.filter((_, i) => i !== index),
      });
    }
  };

  const handleEditConstraint = (type: 'hard' | 'soft', index: number) => {
    // TODO: Implement edit functionality
    console.log('Edit constraint:', type, index);
  };

  const renderConstraintForm = () => {
    const constraintType = CONSTRAINT_TYPES.find(c => c.type === selectedType);
    if (!constraintType) return null;

    switch (selectedType) {
      case 'venue_availability':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Venue Name
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={params.venue || ''}
                onChange={(e) => setParams({ ...params, venue: e.target.value })}
                placeholder="Enter venue name"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Unavailable Dates (comma-separated)
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={params.dates?.join(', ') || ''}
                onChange={(e) => setParams({ 
                  ...params, 
                  dates: e.target.value.split(',').map(d => d.trim()).filter(Boolean)
                })}
                placeholder="2025-06-15, 2025-07-01"
              />
            </div>
          </div>
        );
      
      case 'travel_minimization':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Maximum Consecutive Away Games
              </label>
              <input
                type="number"
                min="1"
                max="10"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={params.max_consecutive_away || 3}
                onChange={(e) => setParams({ 
                  ...params, 
                  max_consecutive_away: parseInt(e.target.value)
                })}
              />
            </div>
          </div>
        );
      
      case 'rest_period':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Minimum Days Between Matches
              </label>
              <input
                type="number"
                min="1"
                max="14"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={params.minimum_days || 7}
                onChange={(e) => setParams({ 
                  ...params, 
                  minimum_days: parseInt(e.target.value)
                })}
              />
            </div>
          </div>
        );
      
      default:
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Parameters (JSON)
            </label>
            <textarea
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={3}
              value={JSON.stringify(params, null, 2)}
              onChange={(e) => {
                try {
                  setParams(JSON.parse(e.target.value));
                } catch {
                  // Invalid JSON, ignore
                }
              }}
              placeholder="{}"
            />
          </div>
        );
    }
  };

  return (
    <Dialog
      open={isConstraintBuilderOpen}
      onClose={closeConstraintBuilder}
      className="relative z-50"
    >
      <div className="fixed inset-0 bg-black/30" aria-hidden="true" />
      
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <Dialog.Panel className="mx-auto max-w-4xl w-full bg-white rounded-lg shadow-xl">
          <div className="flex items-center justify-between p-6 border-b border-gray-200">
            <Dialog.Title className="text-lg font-semibold text-gray-900">
              Constraint Builder
            </Dialog.Title>
            <button
              onClick={closeConstraintBuilder}
              className="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon className="h-6 w-6" />
            </button>
          </div>
          
          <div className="p-6 max-h-96 overflow-y-auto">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Add New Constraint */}
              <Card title="Add New Constraint">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Constraint Type
                    </label>
                    <select
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      value={selectedType}
                      onChange={(e) => {
                        setSelectedType(e.target.value);
                        const type = CONSTRAINT_TYPES.find(c => c.type === e.target.value);
                        setIsHardConstraint(type?.isHard || true);
                        setParams({});
                      }}
                    >
                      <option value="">Select a constraint type</option>
                      {CONSTRAINT_TYPES.map((type) => (
                        <option key={type.type} value={type.type}>
                          {type.name} ({type.isHard ? 'Hard' : 'Soft'})
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  {selectedType && !CONSTRAINT_TYPES.find(c => c.type === selectedType)?.isHard && (
                    <ConstraintWeightSlider
                      value={weight}
                      onChange={setWeight}
                      label="Constraint Weight"
                    />
                  )}
                  
                  {selectedType && renderConstraintForm()}
                  
                  <Button
                    onClick={handleAddConstraint}
                    disabled={!selectedType}
                    className="w-full flex items-center justify-center space-x-2"
                  >
                    <PlusIcon className="h-4 w-4" />
                    <span>Add Constraint</span>
                  </Button>
                </div>
              </Card>
              
              {/* Existing Constraints */}
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-4">
                  Current Constraints
                </h3>
                <div className="space-y-3 max-h-80 overflow-y-auto">
                  {constraints.hard.map((constraint, index) => (
                    <ConstraintCard
                      key={`hard-${index}`}
                      constraint={constraint}
                      type="hard"
                      onEdit={() => handleEditConstraint('hard', index)}
                      onDelete={() => handleDeleteConstraint('hard', index)}
                    />
                  ))}
                  {constraints.soft.map((constraint, index) => (
                    <ConstraintCard
                      key={`soft-${index}`}
                      constraint={constraint}
                      type="soft"
                      onEdit={() => handleEditConstraint('soft', index)}
                      onDelete={() => handleDeleteConstraint('soft', index)}
                    />
                  ))}
                  {constraints.hard.length === 0 && constraints.soft.length === 0 && (
                    <p className="text-gray-500 text-center py-8">
                      No constraints configured yet
                    </p>
                  )}
                </div>
              </div>
            </div>
          </div>
          
          <div className="flex justify-end space-x-3 p-6 border-t border-gray-200">
            <Button variant="outline" onClick={closeConstraintBuilder}>
              Cancel
            </Button>
            <Button onClick={closeConstraintBuilder}>
              Save Constraints
            </Button>
          </div>
        </Dialog.Panel>
      </div>
    </Dialog>
  );
};