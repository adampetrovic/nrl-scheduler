import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface UIState {
  // View states
  selectedRound: number;
  selectedTeam: number | null;
  showConstraintViolations: boolean;
  selectedView: 'grid' | 'team' | 'venue';
  
  // Modal states
  isConstraintBuilderOpen: boolean;
  isMatchDetailsOpen: boolean;
  selectedMatchId: number | null;
  
  // Sidebar and layout
  sidebarCollapsed: boolean;
  darkMode: boolean;
  
  // Actions
  setSelectedRound: (round: number) => void;
  setSelectedTeam: (teamId: number | null) => void;
  setShowConstraintViolations: (show: boolean) => void;
  setSelectedView: (view: 'grid' | 'team' | 'venue') => void;
  
  // Modal actions
  openConstraintBuilder: () => void;
  closeConstraintBuilder: () => void;
  openMatchDetails: (matchId: number) => void;
  closeMatchDetails: () => void;
  
  // Layout actions
  toggleSidebar: () => void;
  toggleDarkMode: () => void;
  
  // Reset
  reset: () => void;
}

export const useUIStore = create<UIState>()(
  devtools(
    (set) => ({
      // Initial state
      selectedRound: 1,
      selectedTeam: null,
      showConstraintViolations: true,
      selectedView: 'grid',
      
      isConstraintBuilderOpen: false,
      isMatchDetailsOpen: false,
      selectedMatchId: null,
      
      sidebarCollapsed: false,
      darkMode: false,
      
      // View actions
      setSelectedRound: (round) => set({ selectedRound: round }),
      setSelectedTeam: (teamId) => set({ selectedTeam: teamId }),
      setShowConstraintViolations: (show) => set({ showConstraintViolations: show }),
      setSelectedView: (view) => set({ selectedView: view }),
      
      // Modal actions
      openConstraintBuilder: () => set({ isConstraintBuilderOpen: true }),
      closeConstraintBuilder: () => set({ isConstraintBuilderOpen: false }),
      openMatchDetails: (matchId) => set({ 
        isMatchDetailsOpen: true, 
        selectedMatchId: matchId 
      }),
      closeMatchDetails: () => set({ 
        isMatchDetailsOpen: false, 
        selectedMatchId: null 
      }),
      
      // Layout actions
      toggleSidebar: () => set((state) => ({ 
        sidebarCollapsed: !state.sidebarCollapsed 
      })),
      toggleDarkMode: () => set((state) => ({ 
        darkMode: !state.darkMode 
      })),
      
      // Reset
      reset: () => set({
        selectedRound: 1,
        selectedTeam: null,
        showConstraintViolations: true,
        selectedView: 'grid',
        isConstraintBuilderOpen: false,
        isMatchDetailsOpen: false,
        selectedMatchId: null,
        sidebarCollapsed: false,
        darkMode: false,
      }),
    }),
    { name: 'ui-store' }
  )
);