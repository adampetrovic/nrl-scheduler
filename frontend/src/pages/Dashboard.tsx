import React, { useEffect } from 'react';
import { 
  PlusIcon, 
  PlayIcon, 
  CogIcon, 
  ChartBarIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon 
} from '@heroicons/react/24/outline';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { useDrawStore } from '../store/drawStore';
import { useTeamsStore } from '../store/teamsStore';
import { useUIStore } from '../store/uiStore';

export const Dashboard: React.FC = () => {
  const { 
    draws, 
    loading, 
    error,
    fetchDraws,
    setCurrentDraw,
    generateDraw,
    startOptimization 
  } = useDrawStore();
  
  const { teams, fetchTeams } = useTeamsStore();
  const { openConstraintBuilder } = useUIStore();

  useEffect(() => {
    fetchDraws();
    fetchTeams();
  }, [fetchDraws, fetchTeams]);

  const handleCreateDraw = () => {
    // TODO: Open create draw modal
    console.log('Create new draw');
  };

  const handleGenerateDraw = async (drawId: number) => {
    await generateDraw(drawId);
  };

  const handleStartOptimization = async (drawId: number) => {
    const jobId = await startOptimization(drawId);
    if (jobId) {
      console.log('Optimization started:', jobId);
      // TODO: Navigate to optimization progress page
    }
  };

  const getDrawStats = () => {
    const total = draws.length;
    const complete = draws.filter(d => d.is_complete).length;
    const incomplete = total - complete;
    
    return { total, complete, incomplete };
  };

  const stats = getDrawStats();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-600 mt-1">
            Manage your NRL season scheduling and optimization
          </p>
        </div>
        <div className="flex items-center space-x-3">
          <Button
            variant="outline"
            onClick={openConstraintBuilder}
            className="flex items-center space-x-2"
          >
            <CogIcon className="h-4 w-4" />
            <span>Manage Constraints</span>
          </Button>
          <Button
            onClick={handleCreateDraw}
            className="flex items-center space-x-2"
          >
            <PlusIcon className="h-4 w-4" />
            <span>New Draw</span>
          </Button>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <Card className="border-red-200 bg-red-50">
          <div className="flex items-center space-x-2">
            <ExclamationTriangleIcon className="h-5 w-5 text-red-500" />
            <p className="text-red-700">{error}</p>
          </div>
        </Card>
      )}

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card className="text-center">
          <div className="text-3xl font-bold text-blue-600">
            {stats.total}
          </div>
          <div className="text-sm text-gray-600 mt-1">Total Draws</div>
        </Card>
        
        <Card className="text-center">
          <div className="text-3xl font-bold text-green-600">
            {stats.complete}
          </div>
          <div className="text-sm text-gray-600 mt-1">Complete</div>
        </Card>
        
        <Card className="text-center">
          <div className="text-3xl font-bold text-orange-600">
            {stats.incomplete}
          </div>
          <div className="text-sm text-gray-600 mt-1">In Progress</div>
        </Card>
        
        <Card className="text-center">
          <div className="text-3xl font-bold text-purple-600">
            {teams.length}
          </div>
          <div className="text-sm text-gray-600 mt-1">Teams</div>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card title="Quick Actions">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="text-center p-6 border border-gray-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all">
            <PlusIcon className="h-12 w-12 text-blue-600 mx-auto mb-3" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              Create New Draw
            </h3>
            <p className="text-sm text-gray-600 mb-4">
              Start a new season draw with teams and constraints
            </p>
            <Button onClick={handleCreateDraw} className="w-full">
              Create Draw
            </Button>
          </div>
          
          <div className="text-center p-6 border border-gray-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all">
            <CogIcon className="h-12 w-12 text-green-600 mx-auto mb-3" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              Manage Constraints
            </h3>
            <p className="text-sm text-gray-600 mb-4">
              Configure scheduling rules and optimization parameters
            </p>
            <Button onClick={openConstraintBuilder} variant="outline" className="w-full">
              Open Builder
            </Button>
          </div>
          
          <div className="text-center p-6 border border-gray-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all">
            <ChartBarIcon className="h-12 w-12 text-purple-600 mx-auto mb-3" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              View Analytics
            </h3>
            <p className="text-sm text-gray-600 mb-4">
              Analyze draw quality and constraint satisfaction
            </p>
            <Button variant="outline" className="w-full" disabled>
              Coming Soon
            </Button>
          </div>
        </div>
      </Card>

      {/* Recent Draws */}
      <Card title="Recent Draws" subtitle="Your latest scheduling projects">
        {loading ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            <p className="text-gray-500 mt-2">Loading draws...</p>
          </div>
        ) : draws.length > 0 ? (
          <div className="space-y-4">
            {draws
              .sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
              .slice(0, 5)
              .map((draw) => (
                <div
                  key={draw.id}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:shadow-md cursor-pointer transition-all"
                  onClick={() => setCurrentDraw(draw)}
                >
                  <div className="flex items-center space-x-4">
                    <div className="flex-shrink-0">
                      {draw.is_complete ? (
                        <CheckCircleIcon className="h-8 w-8 text-green-500" />
                      ) : (
                        <ExclamationTriangleIcon className="h-8 w-8 text-orange-500" />
                      )}
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900">
                        {draw.name}
                      </h3>
                      <p className="text-sm text-gray-600">
                        {draw.season_year} Season • {draw.rounds} rounds
                      </p>
                      <p className="text-xs text-gray-500">
                        Last updated: {new Date(draw.updated_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      draw.is_complete
                        ? 'bg-green-100 text-green-800'
                        : 'bg-orange-100 text-orange-800'
                    }`}>
                      {draw.is_complete ? 'Complete' : 'In Progress'}
                    </span>
                    
                    <div className="flex items-center space-x-1">
                      {!draw.is_complete && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleGenerateDraw(draw.id);
                          }}
                          className="flex items-center space-x-1"
                        >
                          <PlayIcon className="h-3 w-3" />
                          <span>Generate</span>
                        </Button>
                      )}
                      
                      <Button
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStartOptimization(draw.id);
                        }}
                        className="flex items-center space-x-1"
                      >
                        <CogIcon className="h-3 w-3" />
                        <span>Optimize</span>
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <ChartBarIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              No draws yet
            </h3>
            <p className="text-gray-600 mb-4">
              Create your first draw to get started with NRL scheduling
            </p>
            <Button onClick={handleCreateDraw}>
              Create Your First Draw
            </Button>
          </div>
        )}
      </Card>
    </div>
  );
};