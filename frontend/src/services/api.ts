import {
  Team,
  Venue,
  Draw,
  Match,
  CreateDrawRequest,
  OptimizationProgress,
  OptimizationResult,
  APIResponse,
} from '../types';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080';

class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'APIError';
  }
}

class APIClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}/api/v1${endpoint}`;
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new APIError(
          response.status,
          errorData.error || `HTTP ${response.status}: ${response.statusText}`
        );
      }

      return await response.json();
    } catch (error) {
      if (error instanceof APIError) {
        throw error;
      }
      throw new APIError(0, `Network error: ${error}`);
    }
  }

  // Teams API
  async getTeams(): Promise<Team[]> {
    return this.request<Team[]>('/teams');
  }

  async getTeam(id: number): Promise<Team> {
    return this.request<Team>(`/teams/${id}`);
  }

  async createTeam(team: Omit<Team, 'id' | 'created_at' | 'updated_at'>): Promise<Team> {
    return this.request<Team>('/teams', {
      method: 'POST',
      body: JSON.stringify(team),
    });
  }

  async updateTeam(id: number, team: Partial<Team>): Promise<Team> {
    return this.request<Team>(`/teams/${id}`, {
      method: 'PUT',
      body: JSON.stringify(team),
    });
  }

  async deleteTeam(id: number): Promise<void> {
    await this.request<void>(`/teams/${id}`, {
      method: 'DELETE',
    });
  }

  // Venues API
  async getVenues(): Promise<Venue[]> {
    return this.request<Venue[]>('/venues');
  }

  async getVenue(id: number): Promise<Venue> {
    return this.request<Venue>(`/venues/${id}`);
  }

  async createVenue(venue: Omit<Venue, 'id' | 'created_at' | 'updated_at'>): Promise<Venue> {
    return this.request<Venue>('/venues', {
      method: 'POST',
      body: JSON.stringify(venue),
    });
  }

  async updateVenue(id: number, venue: Partial<Venue>): Promise<Venue> {
    return this.request<Venue>(`/venues/${id}`, {
      method: 'PUT',
      body: JSON.stringify(venue),
    });
  }

  async deleteVenue(id: number): Promise<void> {
    await this.request<void>(`/venues/${id}`, {
      method: 'DELETE',
    });
  }

  // Draws API
  async getDraws(): Promise<{ data: Draw[]; total: number; page: number; per_page: number; total_pages: number }> {
    return this.request<{ data: Draw[]; total: number; page: number; per_page: number; total_pages: number }>('/draws');
  }

  async getDraw(id: number): Promise<Draw> {
    return this.request<Draw>(`/draws/${id}`);
  }

  async createDraw(draw: CreateDrawRequest): Promise<Draw> {
    return this.request<Draw>('/draws', {
      method: 'POST',
      body: JSON.stringify(draw),
    });
  }

  async updateDraw(id: number, draw: Partial<Draw>): Promise<Draw> {
    return this.request<Draw>(`/draws/${id}`, {
      method: 'PUT',
      body: JSON.stringify(draw),
    });
  }

  async deleteDraw(id: number): Promise<void> {
    await this.request<void>(`/draws/${id}`, {
      method: 'DELETE',
    });
  }

  async getDrawMatches(drawId: number): Promise<Match[]> {
    return this.request<Match[]>(`/draws/${drawId}/matches`);
  }

  // Draw Generation
  async generateDraw(drawId: number): Promise<Draw> {
    return this.request<Draw>(`/draws/${drawId}/generate`, {
      method: 'POST',
    });
  }

  async validateConstraints(drawId: number): Promise<APIResponse<any>> {
    return this.request<APIResponse<any>>(`/draws/${drawId}/validate-constraints`, {
      method: 'POST',
    });
  }

  // Optimization API
  async startOptimization(drawId: number): Promise<{ job_id: string }> {
    return this.request<{ job_id: string }>(`/optimize/${drawId}/start`, {
      method: 'POST',
    });
  }

  async getOptimizationStatus(jobId: string): Promise<OptimizationProgress> {
    return this.request<OptimizationProgress>(`/optimize/${jobId}/status`);
  }

  async cancelOptimization(jobId: string): Promise<void> {
    await this.request<void>(`/optimize/${jobId}/cancel`, {
      method: 'POST',
    });
  }

  async getOptimizationResult(jobId: string): Promise<OptimizationResult> {
    return this.request<OptimizationResult>(`/optimize/${jobId}/result`);
  }
}

export const apiClient = new APIClient();
export { APIError };