import { create } from 'zustand';
import { apiClient } from '@/lib/api-client';
import { wsService } from '@/lib/websocket';
import { DashboardStats, WebSocketMessage } from '@/types';

interface StatsState {
  stats: DashboardStats | null;
  isLoading: boolean;
  error: string | null;
  isInitialized: boolean;
  fetchStats: () => Promise<void>;
  initializeWebsocket: () => void;
}

export const useStatsStore = create<StatsState>((set, get) => ({
  stats: null,
  isLoading: true,
  error: null,
  isInitialized: false,

  fetchStats: async () => {
    // We only set isLoading to true if we don't have stats yet
    // This prevents UI flashing on background refresh
    const currentStats = get().stats;
    if (!currentStats) {
      set({ isLoading: true });
    }
    
    try {
      const data = await apiClient.getDashboardStats();
      set({ stats: data, error: null, isLoading: false });
    } catch (err: any) {
      console.error('Failed to fetch dashboard stats:', err);
      // Don't overwrite existing stats if a background refresh fails
      if (!get().stats) {
        set({ error: err.message || 'Failed to fetch dashboard stats', isLoading: false });
      } else {
        set({ isLoading: false });
      }
    }
  },

  initializeWebsocket: () => {
    const { isInitialized, fetchStats } = get();
    if (isInitialized) return;

    // Listen for incoming messages
    wsService.addListener((message: WebSocketMessage) => {
      if (message.type === 'stats_changed') {
        // When stats change, fetch the latest stats dynamically in the background
        fetchStats();
      } else if (message.type === 'connected') {
        // Re-subscribe if we reconnect
        wsService.send({ type: 'subscribe_stats' });
      }
    });

    // Give the WebSocket a chance to connect if it hasn't already
    const checkConnectionAndSubscribe = () => {
      if (wsService.isConnected()) {
        wsService.send({ type: 'subscribe_stats' });
      } else {
        setTimeout(checkConnectionAndSubscribe, 1000);
      }
    };
    checkConnectionAndSubscribe();

    set({ isInitialized: true });
  }
}));
