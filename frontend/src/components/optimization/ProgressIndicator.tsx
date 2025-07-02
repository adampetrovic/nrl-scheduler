import React from 'react';
import { 
  FireIcon, 
  ArrowTrendingUpIcon, 
  ArrowTrendingDownIcon 
} from '@heroicons/react/24/outline';
import { OptimizationProgress } from '../../types';
import { Card } from '../ui/Card';

interface ProgressIndicatorProps {
  progress: OptimizationProgress;
  maxIterations: number;
}

export const ProgressIndicator: React.FC<ProgressIndicatorProps> = ({
  progress,
  maxIterations,
}) => {
  const progressPercentage = Math.min((progress.iteration / maxIterations) * 100, 100);
  const temperaturePercentage = Math.min((progress.temperature / 100) * 100, 100);
  
  const formatNumber = (num: number, decimals: number = 1): string => {
    return num.toFixed(decimals);
  };

  const getScoreTrend = (): 'up' | 'down' | 'neutral' => {
    const scoreDiff = progress.current_score - progress.best_score;
    if (Math.abs(scoreDiff) < 0.1) return 'neutral';
    return scoreDiff > 0 ? 'down' : 'up'; // Lower scores are better
  };

  const scoreTrend = getScoreTrend();

  return (
    <Card title="Optimization Progress">
      <div className="space-y-6">
        {/* Overall Progress */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-gray-700">
              Overall Progress
            </span>
            <span className="text-sm text-gray-500">
              {progress.iteration.toLocaleString()} / {maxIterations.toLocaleString()}
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-3">
            <div
              className="bg-blue-600 h-3 rounded-full transition-all duration-300 ease-out"
              style={{ width: `${progressPercentage}%` }}
            />
          </div>
          <div className="text-right mt-1">
            <span className="text-xs text-gray-500">
              {formatNumber(progressPercentage)}% complete
            </span>
          </div>
        </div>

        {/* Temperature */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <FireIcon className="h-4 w-4 text-orange-500" />
              <span className="text-sm font-medium text-gray-700">
                Temperature
              </span>
            </div>
            <span className="text-sm text-gray-900 font-mono">
              {formatNumber(progress.temperature, 2)}
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-gradient-to-r from-red-500 to-orange-300 h-2 rounded-full transition-all duration-300 ease-out"
              style={{ width: `${temperaturePercentage}%` }}
            />
          </div>
        </div>

        {/* Score Tracking */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium text-gray-700">
                Current Score
              </span>
              {scoreTrend === 'up' && (
                <ArrowTrendingUpIcon className="h-4 w-4 text-green-500" />
              )}
              {scoreTrend === 'down' && (
                <ArrowTrendingDownIcon className="h-4 w-4 text-red-500" />
              )}
            </div>
            <div className="text-2xl font-bold text-gray-900 font-mono">
              {formatNumber(progress.current_score)}
            </div>
          </div>
          
          <div className="space-y-2">
            <span className="text-sm font-medium text-gray-700">
              Best Score
            </span>
            <div className="text-2xl font-bold text-green-600 font-mono">
              {formatNumber(progress.best_score)}
            </div>
          </div>
        </div>

        {/* Acceptance Rate */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-gray-700">
              Acceptance Rate
            </span>
            <span className="text-sm text-gray-900 font-mono">
              {formatNumber(progress.acceptance_rate * 100)}%
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all duration-300 ease-out ${
                progress.acceptance_rate > 0.5
                  ? 'bg-green-500'
                  : progress.acceptance_rate > 0.2
                    ? 'bg-yellow-500'
                    : 'bg-red-500'
              }`}
              style={{ width: `${progress.acceptance_rate * 100}%` }}
            />
          </div>
        </div>

        {/* Status and Time */}
        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
          <div className="flex items-center space-x-2">
            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            <span className="text-sm font-medium text-gray-700">
              Optimizing...
            </span>
          </div>
          <div className="text-sm text-gray-600">
            <span className="font-mono">
              ETA: {progress.estimated_time}
            </span>
          </div>
        </div>

        {/* Progress Details */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
          <div className="p-3 bg-blue-50 rounded-lg">
            <div className="text-lg font-bold text-blue-600">
              {formatNumber(progressPercentage)}%
            </div>
            <div className="text-xs text-blue-700">Complete</div>
          </div>
          
          <div className="p-3 bg-green-50 rounded-lg">
            <div className="text-lg font-bold text-green-600">
              {formatNumber(progress.best_score)}
            </div>
            <div className="text-xs text-green-700">Best Score</div>
          </div>
          
          <div className="p-3 bg-orange-50 rounded-lg">
            <div className="text-lg font-bold text-orange-600">
              {formatNumber(progress.temperature, 1)}
            </div>
            <div className="text-xs text-orange-700">Temperature</div>
          </div>
          
          <div className="p-3 bg-purple-50 rounded-lg">
            <div className="text-lg font-bold text-purple-600">
              {formatNumber(progress.acceptance_rate * 100)}%
            </div>
            <div className="text-xs text-purple-700">Acceptance</div>
          </div>
        </div>
      </div>
    </Card>
  );
};