'use client';

import { useEffect, useState, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';
import type { DataSource } from '@/types';

interface DataSourceSelectorProps {
  value: string;
  onChange: (dataSourceId: string) => void;
  disabled?: boolean;
}

export default function DataSourceSelector({
  value,
  onChange,
  disabled = false,
}: DataSourceSelectorProps) {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDataSources = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const sources = await apiClient.getDataSources();
      const activeSources = sources.filter((ds) => ds.is_active);
      setDataSources(activeSources);

      // Auto-select first data source if none selected
      if (!value && activeSources.length > 0) {
        onChange(activeSources[0].id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data sources');
    } finally {
      setLoading(false);
    }
  }, [value, onChange]);

  useEffect(() => {
    fetchDataSources();
  }, [fetchDataSources]);

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="h-10 bg-gray-200 dark:bg-gray-700 rounded"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        <button
          onClick={fetchDataSources}
          className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
        >
          Retry
        </button>
      </div>
    );
  }

  if (dataSources.length === 0) {
    return (
      <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
        <p className="text-sm text-yellow-600 dark:text-yellow-400">
          No active data sources available. Please contact an administrator.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <label
        htmlFor="datasource-select"
        className="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        Data Source
      </label>
      <select
        id="datasource-select"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-800 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {dataSources.map((ds) => (
          <option key={ds.id} value={ds.id}>
            {ds.name} ({ds.type}) - {ds.database_name}
          </option>
        ))}
      </select>
    </div>
  );
}
