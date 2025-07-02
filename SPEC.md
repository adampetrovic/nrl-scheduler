# NRL Scheduler - Implementation Specification

## Overview
Professional-grade NRL season scheduling system with constraint management, optimization engines, and interactive web interface. This document outlines the remaining implementation phases after completing the core foundation (models, draw generation, storage layer).

## Completed Foundation (Phases 1-3)
✅ Go module setup with go-migrate database migrations  
✅ Core domain models (Team, Match, Draw, Venue) with validation  
✅ Round-robin draw generation with bye support and home/away balance  
✅ SQLite repository pattern with transaction support  
✅ Comprehensive test suite (55+ passing tests)  

---

# PHASE 4: Basic Constraint System
**Session Goal**: Implement pluggable constraint framework with core NRL constraints

## 4.1 Constraint Framework
```go
// Create internal/core/constraints/constraint.go
type Constraint interface {
    Validate(match *models.Match, draw *models.Draw) error
    Score(draw *models.Draw) float64
    IsHard() bool
    Name() string
    Description() string
}

type ConstraintEngine struct {
    hardConstraints []Constraint
    softConstraints []WeightedConstraint
}
```

## 4.2 Core NRL Constraints
Implement these specific constraints:

### Hard Constraints
- **VenueAvailabilityConstraint**: Venue unavailable on specific dates
- **ByeConstraint**: Teams get exactly one bye per full round-robin
- **TeamAvailabilityConstraint**: Team unavailable on specific dates
- **DoubleUpConstraint**: Teams can't play each other twice in X rounds

### Soft Constraints  
- **TravelMinimizationConstraint**: Minimize consecutive away games
- **RestPeriodConstraint**: Ensure minimum days between matches
- **PrimeTimeSpreadConstraint**: Distribute prime-time games fairly
- **HomeAwayBalanceConstraint**: Balance home/away games per team

## 4.3 Constraint Configuration
```go
// JSON-based constraint configuration
type ConstraintConfig struct {
    Hard []HardConstraintConfig `json:"hard"`
    Soft []SoftConstraintConfig `json:"soft"`
}

type SoftConstraintConfig struct {
    Type   string                 `json:"type"`
    Weight float64               `json:"weight"`
    Params map[string]interface{} `json:"params"`
}
```

## 4.4 Testing Requirements
- Unit tests for each constraint type
- Integration tests with draw generator
- Constraint violation detection tests
- Configuration parsing tests

## 4.5 Deliverables
- [ ] Constraint interface and engine
- [ ] 8 core NRL constraints implemented
- [ ] JSON configuration system
- [ ] Constraint validation integration with draw generator
- [ ] Full test suite (target: 25+ new tests)

---

# PHASE 5: REST API Foundation  
**Session Goal**: Create HTTP API with teams/venues/draws CRUD operations

## 5.1 API Architecture
```go
// internal/api/server.go - Gin HTTP server
// internal/api/handlers/ - HTTP handlers by resource
// internal/api/middleware/ - Auth, logging, CORS
// pkg/types/ - API request/response types
```

## 5.2 Core Endpoints
```
# Teams Management
GET    /api/v1/teams
POST   /api/v1/teams
GET    /api/v1/teams/:id
PUT    /api/v1/teams/:id
DELETE /api/v1/teams/:id

# Venues Management  
GET    /api/v1/venues
POST   /api/v1/venues
GET    /api/v1/venues/:id
PUT    /api/v1/venues/:id
DELETE /api/v1/venues/:id

# Draws Management
GET    /api/v1/draws
POST   /api/v1/draws
GET    /api/v1/draws/:id
PUT    /api/v1/draws/:id
DELETE /api/v1/draws/:id
GET    /api/v1/draws/:id/matches

# Draw Generation
POST   /api/v1/draws/:id/generate
POST   /api/v1/draws/:id/validate-constraints
```

## 5.3 Request/Response Types
```go
// pkg/types/api.go
type CreateDrawRequest struct {
    Name       string              `json:"name" validate:"required"`
    SeasonYear int                 `json:"season_year" validate:"required,min=2000,max=2100"`
    Rounds     int                 `json:"rounds" validate:"required,min=1,max=52"`
    TeamIDs    []int               `json:"team_ids" validate:"required,min=2"`
    Constraints ConstraintConfig   `json:"constraints"`
}

type ErrorResponse struct {
    Error   string            `json:"error"`
    Details map[string]string `json:"details,omitempty"`
}
```

## 5.4 Middleware Stack
- **Logging**: Request/response logging with correlation IDs
- **Validation**: JSON request validation using validator library
- **Error Handling**: Consistent error response format
- **CORS**: Cross-origin support for frontend

## 5.5 Testing Requirements
- HTTP handler unit tests with test database
- Integration tests for full request/response cycle
- Error handling and validation tests
- API contract compliance tests

## 5.6 Dependencies to Add
```bash
go get github.com/gin-gonic/gin
go get github.com/go-playground/validator/v10
go get github.com/rs/cors
```

## 5.7 Deliverables
- [ ] Gin HTTP server with middleware stack
- [ ] CRUD endpoints for teams, venues, draws
- [ ] Draw generation API endpoint
- [ ] Request validation and error handling
- [ ] API integration tests (target: 20+ new tests)

---

# PHASE 6: Simulated Annealing Optimizer
**Session Goal**: Implement optimization engine to improve constraint satisfaction

## 6.1 Optimizer Architecture  
```go
// internal/core/optimizer/simulated_annealing.go
type SimulatedAnnealing struct {
    Temperature     float64
    CoolingRate     float64
    MaxIterations   int
    ConstraintEngine *constraints.ConstraintEngine
}

type OptimizationResult struct {
    InitialScore    float64
    FinalScore      float64
    Iterations      int
    Improvements    int
    Duration        time.Duration
}
```

## 6.2 Optimization Operations
Implement these draw modification operations:
- **SwapMatches**: Swap two matches between rounds
- **RescheduleMatch**: Move match to different round
- **SwapVenues**: Change venue assignments
- **SwapHomeAway**: Flip home/away for matches

## 6.3 Temperature Schedule
```go
type CoolingSchedule interface {
    NextTemperature(current float64, iteration int) float64
}

// Exponential cooling: T = T0 * (cooling_rate)^iteration
// Linear cooling: T = T0 - (cooling_rate * iteration)
// Adaptive cooling: Adjust based on acceptance rate
```

## 6.4 Progress Tracking
```go
type OptimizationProgress struct {
    Iteration       int     `json:"iteration"`
    Temperature     float64 `json:"temperature"`
    CurrentScore    float64 `json:"current_score"`
    BestScore       float64 `json:"best_score"`
    AcceptanceRate  float64 `json:"acceptance_rate"`
    EstimatedTime   string  `json:"estimated_time"`
}
```

## 6.5 API Integration
```
POST   /api/v1/optimize/:drawId/start
GET    /api/v1/optimize/:jobId/status  
POST   /api/v1/optimize/:jobId/cancel
GET    /api/v1/optimize/:jobId/result
```

## 6.6 Testing Requirements
- Unit tests for each optimization operation
- Temperature schedule validation tests
- Convergence tests with known optimal solutions
- Performance benchmarks for different draw sizes

## 6.7 Deliverables
- [ ] Simulated annealing implementation
- [ ] Draw modification operations
- [ ] Progress tracking and reporting
- [ ] Optimization API endpoints
- [ ] Performance benchmarks (target: 15+ new tests)

---

# PHASE 7: WebSocket Real-time Updates
**Session Goal**: Add real-time optimization progress and live draw updates

## 7.1 WebSocket Architecture
```go
// internal/api/websocket/hub.go - Connection management
// internal/api/websocket/client.go - Individual client handling
// internal/api/websocket/messages.go - Message types

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}
```

## 7.2 Message Types
```go
type MessageType string

const (
    OptimizationProgress MessageType = "optimization_progress"
    DrawUpdate          MessageType = "draw_update"
    ErrorMessage        MessageType = "error"
    KeepAlive          MessageType = "ping"
)

type WebSocketMessage struct {
    Type      MessageType `json:"type"`
    DrawID    int         `json:"draw_id,omitempty"`
    JobID     string      `json:"job_id,omitempty"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
}
```

## 7.3 WebSocket Endpoints
```
WS /ws/optimization/:jobId  # Optimization progress
WS /ws/draws/:drawId        # Draw updates 
WS /ws/global              # Global notifications
```

## 7.4 Integration Points
- Connect optimizer progress callbacks to WebSocket broadcasts
- Send draw modification events in real-time
- Handle client disconnections gracefully
- Implement connection authentication/authorization

## 7.5 Client Management
```go
type Client struct {
    hub     *Hub
    conn    *websocket.Conn
    send    chan []byte
    drawID  int
    jobID   string
    userID  string
}
```

## 7.6 Testing Requirements
- WebSocket connection tests
- Message broadcasting tests  
- Client lifecycle management tests
- Concurrent connection handling tests

## 7.7 Dependencies to Add
```bash
go get github.com/gorilla/websocket
```

## 7.8 Deliverables
- [ ] WebSocket hub and client management
- [ ] Real-time optimization progress updates
- [ ] Draw modification broadcasting
- [ ] Connection lifecycle management
- [ ] WebSocket integration tests (target: 12+ new tests)

---

# PHASE 8: Basic Constraint System Frontend
**Session Goal**: Create React frontend for constraint management and draw visualization

## 8.1 Frontend Architecture
```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/              # Basic UI components
│   │   ├── constraints/     # Constraint management
│   │   ├── draws/          # Draw visualization  
│   │   └── layout/         # App layout
│   ├── hooks/              # Custom React hooks
│   ├── services/           # API integration
│   ├── store/              # Zustand state management
│   └── types/              # TypeScript definitions
```

## 8.2 Core Pages
### Dashboard
- Active draws overview
- Constraint satisfaction summary
- Recent optimization jobs
- Quick actions (generate draw, start optimization)

### Constraint Builder
- Drag-and-drop constraint management
- Hard vs soft constraint toggles
- Weight sliders for soft constraints
- Constraint validation and preview

### Draw Visualizer  
- Round-by-round schedule grid
- Team-centric view with home/away indicators
- Constraint violation highlighting
- Match details modal

## 8.3 Key Components
```tsx
// Constraint management
<ConstraintBuilder />
<ConstraintCard />
<ConstraintWeightSlider />

// Draw visualization
<DrawGrid />
<MatchCard />
<TeamScheduleView />
<ConstraintViolationIndicator />

// Optimization
<OptimizationPanel />
<ProgressIndicator />
<OptimizationHistory />
```

## 8.4 State Management
```typescript
// Zustand stores
interface DrawStore {
  draws: Draw[]
  currentDraw: Draw | null
  constraints: ConstraintConfig
  optimizationStatus: OptimizationProgress
}

interface UIStore {
  selectedRound: number
  selectedTeam: number | null
  showConstraintViolations: boolean
}
```

## 8.5 API Integration
```typescript
// services/api.ts
class APIClient {
  async getDraws(): Promise<Draw[]>
  async createDraw(request: CreateDrawRequest): Promise<Draw>
  async startOptimization(drawId: number): Promise<string>
  async getOptimizationProgress(jobId: string): Promise<OptimizationProgress>
}
```

## 8.6 Technology Stack
- **React 18** with TypeScript
- **Tailwind CSS** + **Headless UI** for styling
- **Zustand** for state management
- **React Query** for API data management
- **React Hook Form** for form handling
- **Recharts** for data visualization

## 8.7 Setup Commands
```bash
npx create-react-app frontend --template typescript
cd frontend
npm install @tailwindcss/ui zustand @tanstack/react-query
npm install react-hook-form recharts
```

## 8.8 Deliverables
- [ ] React application with TypeScript setup
- [ ] Core page layouts and navigation
- [ ] Constraint builder interface
- [ ] Draw visualization components
- [ ] API integration layer
- [ ] Basic responsive design

---

# PHASE 9: CLI Tool for Batch Operations
**Session Goal**: Create command-line interface for scripting and automation

## 9.1 CLI Architecture
```go
// cmd/cli/main.go - CLI entry point
// cmd/cli/commands/ - Command implementations
// internal/cli/ - CLI-specific logic
```

## 9.2 Command Structure
```bash
nrl-scheduler --help

# Database operations
nrl-scheduler db migrate
nrl-scheduler db seed --teams=data/nrl_teams.json

# Team management
nrl-scheduler teams list
nrl-scheduler teams import --file=teams.csv
nrl-scheduler teams export --format=json

# Draw generation
nrl-scheduler generate \
  --teams=16 \
  --rounds=26 \
  --constraints=config/nrl_2025.json \
  --output=nrl_2025_draw.json

# Optimization
nrl-scheduler optimize \
  --draw=nrl_2025_draw.json \
  --max-iterations=10000 \
  --temperature=100 \
  --output=optimized_draw.json

# Validation and reporting
nrl-scheduler validate --draw=draw.json --constraints=config.json
nrl-scheduler report --draw=draw.json --format=html --output=report.html
```

## 9.3 Configuration Files
```yaml
# config/nrl_2025.yaml
teams:
  - { name: "Brisbane Broncos", venue: "Suncorp Stadium", city: "Brisbane" }
  - { name: "Melbourne Storm", venue: "AAMI Park", city: "Melbourne" }
  
constraints:
  hard:
    - type: "venue_availability"
      params: { venue: "Suncorp Stadium", dates: ["2025-06-15"] }
  soft:
    - type: "travel_minimization" 
      weight: 0.8
      params: { max_consecutive_away: 3 }
```

## 9.4 Data Import/Export
```go
// Support multiple formats
type DataFormat string
const (
    JSON DataFormat = "json"
    CSV  DataFormat = "csv" 
    YAML DataFormat = "yaml"
    XML  DataFormat = "xml"
)

// Import/export interfaces
type Importer interface {
    ImportTeams(filename string) ([]*models.Team, error)
    ImportVenues(filename string) ([]*models.Venue, error)
}
```

## 9.5 Reporting System
Generate comprehensive reports:
- **HTML reports** with interactive charts
- **PDF exports** for distribution
- **CSV data** for further analysis
- **Constraint satisfaction** summaries
- **Travel analysis** per team
- **Venue utilization** statistics

## 9.6 Dependencies to Add
```bash
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get gopkg.in/yaml.v3
```

## 9.7 Testing Requirements
- CLI command execution tests
- File import/export validation tests
- Configuration parsing tests
- Error handling and user feedback tests

## 9.8 Deliverables
- [ ] Cobra CLI framework setup
- [ ] Database and team management commands
- [ ] Draw generation and optimization commands
- [ ] Multi-format import/export
- [ ] Reporting and validation commands
- [ ] CLI integration tests (target: 18+ new tests)

---

# PHASE 10: Advanced Optimization & Performance
**Session Goal**: Enhance optimization with genetic algorithms and performance tuning

## 10.1 Genetic Algorithm Implementation
```go
// internal/core/optimizer/genetic.go
type GeneticAlgorithm struct {
    PopulationSize    int
    Generations      int
    MutationRate     float64
    CrossoverRate    float64
    ElitismPercent   float64
    ConstraintEngine *constraints.ConstraintEngine
}

type Individual struct {
    Draw    *models.Draw
    Fitness float64
}
```

## 10.2 Genetic Operations
```go
// Selection methods
type SelectionMethod interface {
    Select(population []Individual) []Individual
}

// Tournament, Roulette, Rank-based selection
// Crossover operations for schedule mixing
// Mutation operations for local optimization
```

## 10.3 Multi-threaded Optimization
```go
type ParallelOptimizer struct {
    Workers     int
    JobQueue    chan OptimizationJob
    ResultQueue chan OptimizationResult
}

// Distribute optimization across CPU cores
// Progress aggregation from multiple workers
// Load balancing for optimal performance
```

## 10.4 Performance Benchmarking
```go
// benchmarks/optimizer_test.go
func BenchmarkSimulatedAnnealing(b *testing.B)
func BenchmarkGeneticAlgorithm(b *testing.B)
func BenchmarkConstraintEvaluation(b *testing.B)

// Profile memory usage and CPU utilization
// Compare algorithm performance on different draw sizes
// Optimization convergence analysis
```

## 10.5 Caching and Optimization
- **Constraint evaluation caching** for repeated calculations
- **Match permutation pre-computation** for faster operations
- **Database query optimization** with proper indexing
- **Result memoization** for common optimization scenarios

## 10.6 Algorithm Comparison API
```
POST /api/v1/compare-algorithms
{
  "draw_id": 123,
  "algorithms": ["simulated_annealing", "genetic"],
  "iterations": 1000,
  "runs": 5
}
```

## 10.7 Deliverables
- [ ] Genetic algorithm implementation
- [ ] Multi-threaded optimization engine
- [ ] Performance benchmarking suite
- [ ] Optimization caching system
- [ ] Algorithm comparison tools
- [ ] Performance analysis (target: 10+ new tests)

---

# SESSION OPTIMIZATION GUIDELINES

## Session Preparation
Each phase should start with:
1. **Context Setup**: "Continue NRL scheduler project at phase X"
2. **Goal Statement**: Clear objective for the session  
3. **Dependencies Check**: Verify previous phase completion
4. **Todo List**: Track progress within the session

## Implementation Strategy
- **Test-Driven Development**: Write tests before implementation
- **Incremental Progress**: Complete one feature before starting next
- **Error Handling**: Comprehensive error scenarios
- **Documentation**: Code comments and API documentation

## Quality Gates
Each phase must achieve:
- ✅ All tests passing (including existing tests)
- ✅ No compilation errors or warnings
- ✅ Proper error handling and validation
- ✅ Code follows established patterns
- ✅ Integration with existing components verified

## Handoff Between Sessions
At session end, provide:
1. **Completion Summary**: What was implemented
2. **Test Results**: Number of tests added/passing  
3. **Next Session**: Specific starting point for phase X+1
4. **Blockers**: Any issues that need resolution

This specification ensures each Claude Code session has clear, achievable goals while building toward a production-ready NRL scheduling system.