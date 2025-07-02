import React, { useState } from 'react';
import { 
  PlayIcon, 
  StopIcon, 
  ClockIcon,
  CogIcon 
} from '@heroicons/react/24/outline';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { ProgressIndicator } from './ProgressIndicator';
import { useDrawStore } from '../../store/drawStore';
import { OptimizationProgress } from '../../types';

interface OptimizationPanelProps {
  drawId?: number;
  onOptimizationComplete?: (result: any) => void;
}

export const OptimizationPanel: React.FC<OptimizationPanelProps> = ({
  drawId,
  onOptimizationComplete,
}) => {
  const { 
    currentDraw, 
    optimizationStatus, 
    optimizing, 
    startOptimization,
    setOptimizationStatus 
  } = useDrawStore();
  
  const [jobId, setJobId] = useState<string | null>(null);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [optimizationParams, setOptimizationParams] = useState({
    maxIterations: 10000,
    temperature: 100,
    coolingRate: 0.95,
    timeout: 300, // 5 minutes
  });

  const selectedDrawId = drawId || currentDraw?.id;

  const handleStartOptimization = async () => {
    if (!selectedDrawId) return;
    
    const newJobId = await startOptimization(selectedDrawId);
    if (newJobId) {
      setJobId(newJobId);
      // Start polling for progress updates
      startProgressPolling(newJobId);
    }
  };

  const handleStopOptimization = () => {
    if (jobId) {
      // TODO: Implement stop optimization API call
      console.log('Stop optimization:', jobId);
      setJobId(null);
      setOptimizationStatus(null);
    }
  };

  const startProgressPolling = (jobIdToTrack: string) => {
    // TODO: Implement WebSocket or polling for real-time updates
    // For now, simulate progress updates
    let iteration = 0;
    const interval = setInterval(() => {
      iteration += Math.floor(Math.random() * 100) + 50;
      
      const mockProgress: OptimizationProgress = {
        iteration,
        temperature: 100 * Math.exp(-iteration / 1000),
        current_score: 150 - (iteration / 100),
        best_score: Math.max(50, 150 - (iteration / 80)),
        acceptance_rate: Math.max(0.1, 0.8 * Math.exp(-iteration / 500)),
        estimated_time: `${Math.max(0, Math.floor((10000 - iteration) / 200))}s`,
      };
      
      setOptimizationStatus(mockProgress);
      
      if (iteration >= optimizationParams.maxIterations) {
        clearInterval(interval);
        setJobId(null);
        onOptimizationComplete?.(mockProgress);
      }
    }, 1000);
  };

  const getOptimizationStatusText = () => {
    if (!optimizationStatus) return 'Ready to optimize';
    if (optimizing || jobId) return 'Optimization in progress...';
    return 'Optimization completed';
  };

  const canStartOptimization = selectedDrawId && !optimizing && !jobId;

  return (
    <div className="space-y-6">
      {/* Control Panel */}
      <Card title="Optimization Control">
        <div className="space-y-4">
          {/* Status */}
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className={`w-3 h-3 rounded-full ${
                optimizing || jobId 
                  ? 'bg-yellow-500 animate-pulse' 
                  : optimizationStatus 
                    ? 'bg-green-500' 
                    : 'bg-gray-300'
              }`} />
              <span className="text-sm font-medium text-gray-900">
                {getOptimizationStatusText()}
              </span>
            </div>
            
            {selectedDrawId && (
              <span className="text-sm text-gray-500">
                Draw: {currentDraw?.name || `#${selectedDrawId}`}
              </span>
            )}
          </div>

          {/* Controls */}
          <div className="flex items-center space-x-3">
            {canStartOptimization ? (
              <Button
                onClick={handleStartOptimization}
                className="flex items-center space-x-2"
                disabled={!selectedDrawId}
              >
                <PlayIcon className="h-4 w-4" />
                <span>Start Optimization</span>
              </Button>
            ) : (
              <Button
                onClick={handleStopOptimization}
                variant="danger"
                className="flex items-center space-x-2"
                disabled={!jobId}
              >
                <StopIcon className="h-4 w-4" />
                <span>Stop Optimization</span>
              </Button>
            )}
            
            <Button
              variant="outline"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="flex items-center space-x-2"
            >
              <CogIcon className="h-4 w-4" />
              <span>Advanced Settings</span>
            </Button>
          </div>

          {/* Advanced Settings */}
          {showAdvanced && (
            <div className="p-4 bg-gray-50 rounded-lg space-y-4">
              <h4 className="text-sm font-semibold text-gray-900">
                Optimization Parameters
              </h4>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Max Iterations
                  </label>
                  <input
                    type="number"
                    min="1000"
                    max="100000"
                    step="1000"
                    value={optimizationParams.maxIterations}
                    onChange={(e) => setOptimizationParams({
                      ...optimizationParams,
                      maxIterations: parseInt(e.target.value)
                    })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={optimizing || !!jobId}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Initial Temperature
                  </label>
                  <input
                    type="number"
                    min="1"
                    max="1000"
                    step="10"
                    value={optimizationParams.temperature}
                    onChange={(e) => setOptimizationParams({
                      ...optimizationParams,
                      temperature: parseInt(e.target.value)
                    })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={optimizing || !!jobId}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Cooling Rate
                  </label>
                  <input
                    type="number"
                    min="0.1"
                    max="0.99"
                    step="0.01"
                    value={optimizationParams.coolingRate}
                    onChange={(e) => setOptimizationParams({
                      ...optimizationParams,
                      coolingRate: parseFloat(e.target.value)
                    })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={optimizing || !!jobId}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Timeout (seconds)
                  </label>
                  <input
                    type="number"
                    min="60"
                    max="3600"
                    step="60"
                    value={optimizationParams.timeout}
                    onChange={(e) => setOptimizationParams({
                      ...optimizationParams,
                      timeout: parseInt(e.target.value)
                    })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={optimizing || !!jobId}
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Progress Display */}
      {optimizationStatus && (
        <ProgressIndicator 
          progress={optimizationStatus}
          maxIterations={optimizationParams.maxIterations}
        />
      )}

      {/* Quick Stats */}
      {optimizationStatus && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Card className="text-center">
            <div className="text-2xl font-bold text-blue-600">
              {optimizationStatus.iteration.toLocaleString()}
            </div>
            <div className="text-sm text-gray-600">Iterations</div>
          </Card>
          
          <Card className="text-center">
            <div className="text-2xl font-bold text-green-600">
              {optimizationStatus.best_score.toFixed(1)}
            </div>
            <div className="text-sm text-gray-600">Best Score</div>
          </Card>
          
          <Card className="text-center">
            <div className="text-2xl font-bold text-orange-600">
              {(optimizationStatus.acceptance_rate * 100).toFixed(1)}%
            </div>
            <div className="text-sm text-gray-600">Acceptance Rate</div>
          </Card>
          
          <Card className="text-center">
            <div className="text-2xl font-bold text-purple-600 flex items-center justify-center space-x-1">
              <ClockIcon className="h-6 w-6" />
              <span>{optimizationStatus.estimated_time}</span>
            </div>
            <div className="text-sm text-gray-600">Time Remaining</div>
          </Card>
        </div>
      )}
    </div>
  );
};