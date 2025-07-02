# NRL Scheduler Frontend

React frontend for the NRL scheduling system with constraint management and optimization features.

## Features

- **Dashboard**: Overview of draws, teams, and quick actions
- **Constraint Builder**: Drag-and-drop interface for managing scheduling constraints
- **Draw Visualization**: Grid and team-centric views of match schedules
- **Optimization Panel**: Real-time optimization progress tracking
- **Team Management**: CRUD operations for teams and venues

## Technology Stack

- React 18 with TypeScript
- Tailwind CSS + Headless UI for styling
- Zustand for state management
- Heroicons for iconography
- Responsive design for mobile and desktop

## Getting Started

```bash
# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build

# Run tests
npm test
```

## Configuration

Set the API base URL in `.env`:

```
REACT_APP_API_BASE_URL=http://localhost:8080
```

## Project Structure

```
src/
├── components/
│   ├── ui/              # Basic UI components (Button, Card, etc.)
│   ├── constraints/     # Constraint management components
│   ├── draws/          # Draw visualization components
│   ├── layout/         # App layout and navigation
│   └── optimization/   # Optimization tracking components
├── hooks/              # Custom React hooks
├── pages/              # Main page components
├── services/           # API integration
├── store/              # Zustand state management
└── types/              # TypeScript type definitions
```

## Key Components

### Constraint Management
- `ConstraintBuilder`: Modal interface for creating and editing constraints
- `ConstraintCard`: Individual constraint display with edit/delete actions
- `ConstraintWeightSlider`: Adjustable weight control for soft constraints

### Draw Visualization
- `DrawGrid`: Round-by-round schedule grid view
- `TeamScheduleView`: Team-centric schedule display
- `MatchCard`: Individual match display with violation indicators

### Optimization
- `OptimizationPanel`: Control panel for starting/stopping optimization
- `ProgressIndicator`: Real-time progress visualization
- `OptimizationHistory`: Historical optimization job tracking

## State Management

The app uses Zustand for state management with three main stores:

- `drawStore`: Draw data, matches, and optimization state
- `teamsStore`: Team and venue management
- `uiStore`: UI state (selected views, modals, etc.)

## API Integration

All API calls are handled through the `APIClient` class in `services/api.ts`, providing:

- Teams and venues CRUD operations
- Draw generation and management
- Constraint validation
- Optimization job control

## Development

The frontend is designed to work with the NRL Scheduler Go backend API. Ensure the backend is running on the configured port before starting the frontend development server.