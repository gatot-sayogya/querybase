import { useEffect } from 'react';
import { useStatsStore } from '@/stores/stats-store';
import { DashboardStats } from '@/types';

export function useDashboardStats() {
  const stats = useStatsStore(state => state.stats);
  const isLoading = useStatsStore(state => state.isLoading);
  const error = useStatsStore(state => state.error);
  const fetchStats = useStatsStore(state => state.fetchStats);
  const initializeWebsocket = useStatsStore(state => state.initializeWebsocket);

  useEffect(() => {
    // Fire up the websocket and fetch initial data once
    initializeWebsocket();
    fetchStats();
  }, [fetchStats, initializeWebsocket]);

  return { stats, isLoading, error, refetch: fetchStats };
}
