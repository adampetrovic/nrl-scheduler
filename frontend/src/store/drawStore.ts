import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { Draw, Match, Team, ConstraintConfig, OptimizationProgress } from '../types';
import { apiClient } from '../services/api';

interface DrawState {
  // Data
  draws: Draw[];
  currentDraw: Draw | null;
  matches: Match[];
  teams: Team[];
  constraints: ConstraintConfig;
  optimizationStatus: OptimizationProgress | null;
  
  // Loading states
  loading: boolean;
  optimizing: boolean;
  error: string | null;
  
  // Actions
  setCurrentDraw: (draw: Draw | null) => void;
  setConstraints: (constraints: ConstraintConfig) => void;
  setOptimizationStatus: (status: OptimizationProgress | null) => void;
  
  // Async actions
  fetchDraws: () => Promise<void>;
  fetchDraw: (id: number) => Promise<void>;
  fetchMatches: (drawId: number) => Promise<void>;
  createDraw: (drawData: any) => Promise<Draw | null>;
  generateDraw: (drawId: number) => Promise<void>;
  startOptimization: (drawId: number) => Promise<string | null>;
  validateConstraints: (drawId: number) => Promise<void>;
  
  // UI helpers
  clearError: () => void;
  reset: () => void;
}

const initialConstraints: ConstraintConfig = {
  hard: [],
  soft: []
};

export const useDrawStore = create<DrawState>()(
  devtools(
    (set, get) => ({
      // Initial state
      draws: [],
      currentDraw: null,
      matches: [],
      teams: [],
      constraints: initialConstraints,
      optimizationStatus: null,
      loading: false,
      optimizing: false,
      error: null,
      
      // Setters
      setCurrentDraw: (draw) => set({ currentDraw: draw }),
      setConstraints: (constraints) => set({ constraints }),
      setOptimizationStatus: (status) => set({ optimizationStatus: status }),
      
      // Async actions
      fetchDraws: async () => {
        set({ loading: true, error: null });
        try {
          const response = await apiClient.getDraws();
          // Handle paginated response
          const draws = Array.isArray(response) ? response : response.data || [];
          set({ draws, loading: false });
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to fetch draws',
            loading: false 
          });
        }
      },
      
      fetchDraw: async (id: number) => {
        set({ loading: true, error: null });
        try {
          const draw = await apiClient.getDraw(id);
          set({ currentDraw: draw, loading: false });
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to fetch draw',
            loading: false 
          });
        }
      },
      
      fetchMatches: async (drawId: number) => {
        set({ loading: true, error: null });
        try {
          const matches = await apiClient.getDrawMatches(drawId);
          set({ matches, loading: false });
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to fetch matches',
            loading: false 
          });
        }
      },
      
      createDraw: async (drawData: any) => {
        set({ loading: true, error: null });
        try {
          const newDraw = await apiClient.createDraw(drawData);
          set((state) => ({ 
            draws: [...state.draws, newDraw],
            currentDraw: newDraw,
            loading: false 
          }));
          return newDraw;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to create draw',
            loading: false 
          });
          return null;
        }
      },
      
      generateDraw: async (drawId: number) => {
        set({ loading: true, error: null });
        try {
          const updatedDraw = await apiClient.generateDraw(drawId);
          set((state) => ({
            draws: state.draws.map(d => d.id === drawId ? updatedDraw : d),
            currentDraw: state.currentDraw?.id === drawId ? updatedDraw : state.currentDraw,
            loading: false
          }));
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to generate draw',
            loading: false 
          });
        }
      },
      
      startOptimization: async (drawId: number) => {
        set({ optimizing: true, error: null });
        try {
          const result = await apiClient.startOptimization(drawId);
          set({ optimizing: false });
          return result.job_id;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to start optimization',
            optimizing: false 
          });
          return null;
        }
      },
      
      validateConstraints: async (drawId: number) => {
        set({ loading: true, error: null });
        try {
          await apiClient.validateConstraints(drawId);
          set({ loading: false });
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Constraint validation failed',
            loading: false 
          });
        }
      },
      
      // UI helpers
      clearError: () => set({ error: null }),
      reset: () => set({
        draws: [],
        currentDraw: null,
        matches: [],
        teams: [],
        constraints: initialConstraints,
        optimizationStatus: null,
        loading: false,
        optimizing: false,
        error: null,
      }),
    }),
    { name: 'draw-store' }
  )
);