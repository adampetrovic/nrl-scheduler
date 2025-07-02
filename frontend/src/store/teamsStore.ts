import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { Team, Venue } from '../types';
import { apiClient } from '../services/api';

interface TeamsState {
  // Data
  teams: Team[];
  venues: Venue[];
  
  // Loading states
  loading: boolean;
  error: string | null;
  
  // Actions
  fetchTeams: () => Promise<void>;
  fetchVenues: () => Promise<void>;
  createTeam: (team: Omit<Team, 'id' | 'created_at' | 'updated_at'>) => Promise<Team | null>;
  updateTeam: (id: number, team: Partial<Team>) => Promise<Team | null>;
  deleteTeam: (id: number) => Promise<boolean>;
  createVenue: (venue: Omit<Venue, 'id' | 'created_at' | 'updated_at'>) => Promise<Venue | null>;
  updateVenue: (id: number, venue: Partial<Venue>) => Promise<Venue | null>;
  deleteVenue: (id: number) => Promise<boolean>;
  
  // UI helpers
  clearError: () => void;
  reset: () => void;
}

export const useTeamsStore = create<TeamsState>()(
  devtools(
    (set, get) => ({
      // Initial state
      teams: [],
      venues: [],
      loading: false,
      error: null,
      
      // Team actions
      fetchTeams: async () => {
        set({ loading: true, error: null });
        try {
          const teams = await apiClient.getTeams();
          set({ teams, loading: false });
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to fetch teams',
            loading: false 
          });
        }
      },
      
      createTeam: async (teamData) => {
        set({ loading: true, error: null });
        try {
          const newTeam = await apiClient.createTeam(teamData);
          set((state) => ({ 
            teams: [...state.teams, newTeam],
            loading: false 
          }));
          return newTeam;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to create team',
            loading: false 
          });
          return null;
        }
      },
      
      updateTeam: async (id, teamData) => {
        set({ loading: true, error: null });
        try {
          const updatedTeam = await apiClient.updateTeam(id, teamData);
          set((state) => ({
            teams: state.teams.map(t => t.id === id ? updatedTeam : t),
            loading: false
          }));
          return updatedTeam;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to update team',
            loading: false 
          });
          return null;
        }
      },
      
      deleteTeam: async (id) => {
        set({ loading: true, error: null });
        try {
          await apiClient.deleteTeam(id);
          set((state) => ({
            teams: state.teams.filter(t => t.id !== id),
            loading: false
          }));
          return true;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to delete team',
            loading: false 
          });
          return false;
        }
      },
      
      // Venue actions
      fetchVenues: async () => {
        set({ loading: true, error: null });
        try {
          const venues = await apiClient.getVenues();
          set({ venues, loading: false });
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to fetch venues',
            loading: false 
          });
        }
      },
      
      createVenue: async (venueData) => {
        set({ loading: true, error: null });
        try {
          const newVenue = await apiClient.createVenue(venueData);
          set((state) => ({ 
            venues: [...state.venues, newVenue],
            loading: false 
          }));
          return newVenue;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to create venue',
            loading: false 
          });
          return null;
        }
      },
      
      updateVenue: async (id, venueData) => {
        set({ loading: true, error: null });
        try {
          const updatedVenue = await apiClient.updateVenue(id, venueData);
          set((state) => ({
            venues: state.venues.map(v => v.id === id ? updatedVenue : v),
            loading: false
          }));
          return updatedVenue;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to update venue',
            loading: false 
          });
          return null;
        }
      },
      
      deleteVenue: async (id) => {
        set({ loading: true, error: null });
        try {
          await apiClient.deleteVenue(id);
          set((state) => ({
            venues: state.venues.filter(v => v.id !== id),
            loading: false
          }));
          return true;
        } catch (error) {
          set({ 
            error: error instanceof Error ? error.message : 'Failed to delete venue',
            loading: false 
          });
          return false;
        }
      },
      
      // UI helpers
      clearError: () => set({ error: null }),
      reset: () => set({
        teams: [],
        venues: [],
        loading: false,
        error: null,
      }),
    }),
    { name: 'teams-store' }
  )
);