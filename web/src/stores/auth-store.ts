import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import toast from 'react-hot-toast';
import type { User } from '@/types';
import { apiClient } from '@/lib/api-client';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isHydrating: boolean;
  error: string | null;

  // Actions
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  loadUser: () => Promise<void>;
  clearError: () => void;
  setHydrating: (loading: boolean) => void;
  handleAuthError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      isHydrating: true, // Start as true to prevent premature redirects before hydration
      error: null,

      setHydrating: (loading: boolean) => set({ isHydrating: loading }),

      login: async (username: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiClient.login({ username, password });
          set({
            user: response.user,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      logout: async () => {
        set({ isLoading: true });
        try {
          await apiClient.logout();
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
        }
      },

      loadUser: async () => {
        set({ isLoading: true });
        try {
          const user = await apiClient.getCurrentUser();
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
          });
        }
      },

      clearError: () => set({ error: null }),

      handleAuthError: () => {
        // Clear auth state
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          isLoading: false,
        });
        
        // Show toast notification
        toast.error('Your session has expired. Please log in again.', {
          duration: 5000,
          id: 'auth-error', // Prevent duplicate toasts
        });
        
        // Redirect to login page with session expired flag
        if (typeof window !== 'undefined') {
          window.location.href = '/login?session=expired';
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        // Called when hydration finishes
        if (state) {
          state.setHydrating(false);
        }
      },
    }
  )
);

// Set up the auth error handler on the api client
// This needs to be done after store creation to avoid circular dependencies
if (typeof window !== 'undefined') {
  // Use a timeout to ensure the store is fully initialized
  setTimeout(() => {
    const { handleAuthError } = useAuthStore.getState();
    apiClient.setOnAuthErrorHandler(handleAuthError);
  }, 0);
}
