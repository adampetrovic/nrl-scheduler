import React from 'react';
import { 
  ClockIcon, 
  CheckCircleIcon, 
  XCircleIcon,
  ArrowPathIcon 
} from '@heroicons/react/24/outline';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';

interface OptimizationJob {
  id: string;
  drawId: number;
  drawName: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  startTime: string;
  endTime?: string;
  initialScore: number;
  finalScore?: number;
  iterations: number;
  improvements: number;
  parameters: {
    maxIterations: number;
    temperature: number;
    coolingRate: number;
  };
}

interface OptimizationHistoryProps {
  jobs?: OptimizationJob[];
  onRetryJob?: (job: OptimizationJob) => void;
  onViewDetails?: (job: OptimizationJob) => void;
}

// Mock data for demonstration
const mockJobs: OptimizationJob[] = [
  {
    id: 'job-001',
    drawId: 1,
    drawName: 'NRL 2025 Season',
    status: 'completed',
    startTime: '2025-07-02T10:30:00Z',
    endTime: '2025-07-02T10:45:00Z',
    initialScore: 245.8,
    finalScore: 127.3,
    iterations: 8500,
    improvements: 156,
    parameters: {
      maxIterations: 10000,
      temperature: 100,
      coolingRate: 0.95,
    },
  },
  {
    id: 'job-002',
    drawId: 1,
    drawName: 'NRL 2025 Season',
    status: 'running',
    startTime: '2025-07-02T12:15:00Z',
    initialScore: 189.4,
    iterations: 3200,
    improvements: 45,
    parameters: {
      maxIterations: 15000,
      temperature: 150,
      coolingRate: 0.92,
    },
  },
  {
    id: 'job-003',
    drawId: 2,
    drawName: 'Test Draw',
    status: 'failed',
    startTime: '2025-07-02T09:00:00Z',
    endTime: '2025-07-02T09:02:00Z',
    initialScore: 300.1,
    iterations: 150,
    improvements: 0,
    parameters: {
      maxIterations: 5000,
      temperature: 50,
      coolingRate: 0.98,
    },
  },
];

export const OptimizationHistory: React.FC<OptimizationHistoryProps> = ({
  jobs = mockJobs,
  onRetryJob,
  onViewDetails,
}) => {
  const formatDuration = (start: string, end?: string): string => {
    const startTime = new Date(start);
    const endTime = end ? new Date(end) : new Date();
    const diffMs = endTime.getTime() - startTime.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffSecs = Math.floor((diffMs % 60000) / 1000);
    
    if (diffMins > 0) {
      return `${diffMins}m ${diffSecs}s`;
    }
    return `${diffSecs}s`;
  };

  const formatDateTime = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-AU', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusIcon = (status: OptimizationJob['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
      case 'failed':
        return <XCircleIcon className="h-5 w-5 text-red-500" />;
      case 'cancelled':
        return <XCircleIcon className="h-5 w-5 text-gray-500" />;
      case 'running':
        return <ArrowPathIcon className="h-5 w-5 text-blue-500 animate-spin" />;
      default:
        return <ClockIcon className="h-5 w-5 text-gray-400" />;
    }
  };

  const getStatusColor = (status: OptimizationJob['status']) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'cancelled':
        return 'bg-gray-100 text-gray-800';
      case 'running':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getImprovementScore = (job: OptimizationJob): number | null => {
    if (job.status === 'completed' && job.finalScore) {
      return ((job.initialScore - job.finalScore) / job.initialScore) * 100;
    }
    return null;
  };

  return (
    <Card title="Optimization History" subtitle="Recent optimization jobs and their results">
      {jobs.length > 0 ? (
        <div className="space-y-4">
          {jobs
            .sort((a, b) => new Date(b.startTime).getTime() - new Date(a.startTime).getTime())
            .map((job) => {
              const improvement = getImprovementScore(job);
              
              return (
                <div
                  key={job.id}
                  className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 hover:shadow-md transition-all"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3">
                      <div className="flex-shrink-0 mt-1">
                        {getStatusIcon(job.status)}
                      </div>
                      
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2 mb-1">
                          <h3 className="text-sm font-semibold text-gray-900 truncate">
                            {job.drawName}
                          </h3>
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(job.status)}`}>
                            {job.status.charAt(0).toUpperCase() + job.status.slice(1)}
                          </span>
                        </div>
                        
                        <div className="space-y-1 text-sm text-gray-600">
                          <div className="flex items-center space-x-4">
                            <span>Started: {formatDateTime(job.startTime)}</span>
                            <span>•</span>
                            <span>Duration: {formatDuration(job.startTime, job.endTime)}</span>
                          </div>
                          
                          <div className="flex items-center space-x-4">
                            <span>Iterations: {job.iterations.toLocaleString()}</span>
                            <span>•</span>
                            <span>Improvements: {job.improvements}</span>
                            {improvement !== null && (
                              <>
                                <span>•</span>
                                <span className={`font-medium ${
                                  improvement > 0 ? 'text-green-600' : 'text-red-600'
                                }`}>
                                  {improvement > 0 ? '+' : ''}{improvement.toFixed(1)}% improvement
                                </span>
                              </>
                            )}
                          </div>
                        </div>
                        
                        {/* Score Progress */}
                        <div className="mt-3 space-y-2">
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>Score Progress</span>
                            <span>
                              {job.initialScore.toFixed(1)} → {job.finalScore?.toFixed(1) || 'In Progress'}
                            </span>
                          </div>
                          {job.status === 'completed' && job.finalScore && (
                            <div className="w-full bg-gray-200 rounded-full h-1">
                              <div
                                className="bg-green-500 h-1 rounded-full"
                                style={{ 
                                  width: `${Math.max(0, Math.min(100, improvement || 0))}%` 
                                }}
                              />
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                    
                    <div className="flex items-center space-x-2 ml-4">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => onViewDetails?.(job)}
                        className="text-xs"
                      >
                        Details
                      </Button>
                      
                      {(job.status === 'failed' || job.status === 'cancelled') && (
                        <Button
                          size="sm"
                          onClick={() => onRetryJob?.(job)}
                          className="text-xs flex items-center space-x-1"
                        >
                          <ArrowPathIcon className="h-3 w-3" />
                          <span>Retry</span>
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
        </div>
      ) : (
        <div className="text-center py-12">
          <ClockIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            No optimization history
          </h3>
          <p className="text-gray-600">
            Start your first optimization to see results here
          </p>
        </div>
      )}
    </Card>
  );
};