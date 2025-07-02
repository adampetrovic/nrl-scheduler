import React, { useState } from 'react';
import { Layout } from './components/layout/Layout';
import { Dashboard } from './pages/Dashboard';
import { DrawGrid } from './components/draws/DrawGrid';
import { TeamScheduleView } from './components/draws/TeamScheduleView';
import { ConstraintBuilder } from './components/constraints/ConstraintBuilder';
import { OptimizationPanel } from './components/optimization/OptimizationPanel';
import { OptimizationHistory } from './components/optimization/OptimizationHistory';
import { useDrawStore } from './store/drawStore';
import { useTeamsStore } from './store/teamsStore';
import { useUIStore } from './store/uiStore';

type ViewType = 'dashboard' | 'draws' | 'constraints' | 'optimization';

function App() {
  const [currentView, setCurrentView] = useState<ViewType>('dashboard');
  const { currentDraw, matches } = useDrawStore();
  const { teams } = useTeamsStore();
  const { selectedView } = useUIStore();

  const renderMainContent = () => {
    switch (currentView) {
      case 'dashboard':
        return <Dashboard />;
      
      case 'draws':
        if (!currentDraw) {
          return (
            <div className="text-center py-12">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                No Draw Selected
              </h2>
              <p className="text-gray-600">
                Select a draw from the dashboard to view its schedule
              </p>
            </div>
          );
        }
        
        if (selectedView === 'team') {
          return (
            <TeamScheduleView
              matches={matches}
              teams={teams}
              onMatchClick={(match) => console.log('Match clicked:', match)}
            />
          );
        }
        
        return (
          <DrawGrid
            matches={matches}
            totalRounds={currentDraw.rounds}
            showConstraintViolations={true}
            onMatchClick={(match) => console.log('Match clicked:', match)}
          />
        );
      
      case 'constraints':
        return (
          <div className="space-y-6">
            <div className="text-center">
              <h1 className="text-3xl font-bold text-gray-900 mb-4">
                Constraint Management
              </h1>
              <p className="text-gray-600 mb-8">
                Configure scheduling constraints and optimization parameters
              </p>
            </div>
            <ConstraintBuilder />
          </div>
        );
      
      case 'optimization':
        return (
          <div className="space-y-6">
            <div className="text-center">
              <h1 className="text-3xl font-bold text-gray-900 mb-4">
                Optimization Center
              </h1>
              <p className="text-gray-600 mb-8">
                Optimize your draws using simulated annealing algorithms
              </p>
            </div>
            <OptimizationPanel 
              drawId={currentDraw?.id}
              onOptimizationComplete={(result) => {
                console.log('Optimization completed:', result);
              }}
            />
            <OptimizationHistory 
              onRetryJob={(job) => console.log('Retry job:', job)}
              onViewDetails={(job) => console.log('View details:', job)}
            />
          </div>
        );
      
      default:
        return <Dashboard />;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Layout>
        {/* Navigation Tabs */}
        <div className="mb-6">
          <nav className="flex space-x-8 border-b border-gray-200">
            {[
              { key: 'dashboard', label: 'Dashboard' },
              { key: 'draws', label: 'Draws' },
              { key: 'constraints', label: 'Constraints' },
              { key: 'optimization', label: 'Optimization' },
            ].map(({ key, label }) => (
              <button
                key={key}
                onClick={() => setCurrentView(key as ViewType)}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  currentView === key
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {label}
              </button>
            ))}
          </nav>
        </div>

        {/* Main Content */}
        {renderMainContent()}

        {/* Global Components */}
        <ConstraintBuilder />
      </Layout>
    </div>
  );
}

export default App;
