import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';
import { wsService } from '@/lib/websocket';
import { DashboardStats, WebSocketMessage } from '@/types';

export function useDashboardStats() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      const data = await apiClient.getDashboardStats();
      setStats(data);
      setError(null);
    } catch (err: any) {
      console.error('Failed to fetch dashboard stats:', err);
      // Don't overwrite existing stats on background refresh error
      if (!stats) {
        setError(err.message || 'Failed to fetch dashboard stats');
      }
    } finally {
      setIsLoading(false);
    }
  }, [stats]);

  useEffect(() => {
    // Initial fetch
    fetchStats();

    // Give the WebSocket a chance to connect if it hasn't already
    const checkConnectionAndSubscribe = () => {
      if (wsService.isConnected()) {
        wsService.send({ type: 'subscribe_stats' });
      } else {
        // Try again in a bit if not connected yet
        setTimeout(checkConnectionAndSubscribe, 1000);
      }
    };
    
    checkConnectionAndSubscribe();

    // Listen for incoming messages
    const removeListener = wsService.addListener((message: WebSocketMessage) => {
      if (message.type === 'stats_changed') {
        // When stats change, fetch the latest stats from the REST API
        // This is simpler than sending the full payload through WS for now,
        // and ensures we get the correctly permissioned stats.
        fetchStats();
      } else if (message.type === 'connected') {
        // Re-subscribe if we reconnect
        wsService.send({ type: 'subscribe_stats' });
      }
    });

    return () => {
      removeListener();
    };
  }, [fetchStats]);

  return { stats, isLoading, error, refetch: fetchStats };
}
