export interface Team {
  id: number;
  name: string;
  city: string;
  venue: string;
  created_at: string;
  updated_at: string;
}

export interface Venue {
  id: number;
  name: string;
  city: string;
  capacity: number;
  created_at: string;
  updated_at: string;
}

export interface Match {
  id: number;
  draw_id: number;
  round: number;
  home_team_id: number;
  away_team_id: number;
  venue_id: number;
  date: string;
  is_bye: boolean;
  created_at: string;
  updated_at: string;
  home_team?: Team;
  away_team?: Team;
  venue?: Venue;
}

export interface Draw {
  id: number;
  name: string;
  season_year: number;
  rounds: number;
  is_complete: boolean;
  created_at: string;
  updated_at: string;
  matches?: Match[];
  teams?: Team[];
}

export interface HardConstraintConfig {
  type: string;
  params: Record<string, any>;
}

export interface SoftConstraintConfig {
  type: string;
  weight: number;
  params: Record<string, any>;
}

export interface ConstraintConfig {
  hard: HardConstraintConfig[];
  soft: SoftConstraintConfig[];
}

export interface CreateDrawRequest {
  name: string;
  season_year: number;
  rounds: number;
  team_ids: number[];
  constraints: ConstraintConfig;
}

export interface OptimizationProgress {
  iteration: number;
  temperature: number;
  current_score: number;
  best_score: number;
  acceptance_rate: number;
  estimated_time: string;
}

export interface OptimizationResult {
  initial_score: number;
  final_score: number;
  iterations: number;
  improvements: number;
  duration: string;
}

export interface ErrorResponse {
  error: string;
  details?: Record<string, string>;
}

export interface APIResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}