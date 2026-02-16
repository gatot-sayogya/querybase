'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { DataSource } from '@/types';

interface DataSourceListProps {
  onSelectDataSource?: (dataSourceId: string) => void;
  onEditDataSource?: (dataSource: DataSource) => void;
  selectedId: string | null;
}

export default function DataSourceList({
  onSelectDataSource,
  onEditDataSource,
  selectedId,
}: DataSourceListProps) {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

  useEffect(() => {
    fetchDataSources();
  }, []);

  const fetchDataSources = async () => {
    try {
      setLoading(true);
      setError(null);
      const sources = await apiClient.getDataSources();
      setDataSources(sources);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data sources');
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async (id: string) => {
    try {
      setTestingId(id);
      await apiClient.testDataSourceConnection(id, {});
      toast.success('Connection successful! ✓', { duration: 5000 });
    } catch (err) {
      toast.error(`Connection failed: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    } finally {
      setTestingId(null);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete data source "${name}"?`)) {
      return;
    }

    try {
      await apiClient.deleteDataSource(id);
      setDataSources(dataSources.filter((ds) => ds.id !== id));
    } catch (err) {
      toast.error(`Failed to delete: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const getTypeBadgeColor = (type: string) => {
    return type === 'postgresql'
      ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      : 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300';
  };

  const getHealthStatusColor = (ds: DataSource) => {
    if (!ds.is_active) {
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
    return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="animate-pulse">
            <div className="h-32 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
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

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
          Data Sources
        </h2>
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {dataSources.length} {dataSources.length === 1 ? 'source' : 'sources'}
        </span>
      </div>

      {dataSources.length === 0 ? (
        <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4M0 12h18M0 12h18"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
            No data sources
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Get started by adding your first data source
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {dataSources.map((dataSource) => (
            <div
              key={dataSource.id}
              className={`p-4 rounded-lg border-2 transition-all ${
                selectedId === dataSource.id
                  ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                  : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
              }`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-2">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                      {dataSource.name}
                    </h3>
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded ${getTypeBadgeColor(
                        dataSource.type
                      )}`}
                    >
                      {dataSource.type.toUpperCase()}
                    </span>
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded ${getHealthStatusColor(
                        dataSource
                      )}`}
                    >
                      {dataSource.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  <div className="space-y-1 text-sm text-gray-600 dark:text-gray-400">
                    <p>
                      <span className="font-medium">Host:</span> {dataSource.host}:{dataSource.port}
                    </p>
                    <p>
                      <span className="font-medium">Database:</span> {dataSource.database_name}
                    </p>
                    <p>
                      <span className="font-medium">User:</span> {dataSource.username}
                    </p>
                  </div>
                </div>
                <div className="flex flex-col space-y-2 ml-4">
                  {onSelectDataSource && (
                    <button
                      onClick={() => onSelectDataSource(dataSource.id)}
                      className="px-3 py-1 text-xs font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                    >
                      Select
                    </button>
                  )}
                  {onEditDataSource && (
                    <button
                      onClick={() => onEditDataSource(dataSource)}
                      className="px-3 py-1 text-xs font-medium text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-300"
                    >
                      Edit
                    </button>
                  )}
                  <button
                    onClick={() => handleTestConnection(dataSource.id)}
                    disabled={testingId === dataSource.id}
                    className="px-3 py-1 text-xs font-medium text-green-600 hover:text-green-800 dark:text-green-400 dark:hover:text-green-300 disabled:opacity-50"
                  >
                    {testingId === dataSource.id ? 'Testing...' : 'Test'}
                  </button>
                  <button
                    onClick={() => handleDelete(dataSource.id, dataSource.name)}
                    className="px-3 py-1 text-xs font-medium text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
