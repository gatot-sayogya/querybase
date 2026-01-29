'use client';

import { useEffect, useState, useRef } from 'react';
import {
  ChevronDownIcon,
  ChevronRightIcon,
  TableCellsIcon,
  EyeIcon,
  ServerIcon,
  ChevronDoubleRightIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import { apiClient } from '@/lib/api-client';
import { useSchemaStore } from '@/stores/schema-store';
import { useAuthStore } from '@/stores/auth-store';
import { filterAccessibleDataSources } from '@/lib/data-source-utils';
import type { DataSource } from '@/types';

interface DataSourceSchemaSelectorProps {
  value: string;
  onChange: (dataSourceId: string) => void;
  onTableSelect?: (tableName: string) => void;
  disabled?: boolean;
}

const POLL_INTERVAL = 60000; // 60 seconds

export default function DataSourceSchemaSelector({
  value,
  onChange,
  onTableSelect,
  disabled = false,
}: DataSourceSchemaSelectorProps) {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedDataSources, setExpandedDataSources] = useState<Set<string>>(new Set());
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['tables']));
  const [isPolling, setIsPolling] = useState(true);
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const { user } = useAuthStore();
  const hasInitialized = useRef(false);

  const {
    schemas,
    loadSchema,
    syncSchema,
    lastSyncTime,
    isLoading: isSchemaLoading,
  } = useSchemaStore();

  // Fetch data sources on mount (only once)
  useEffect(() => {
    if (hasInitialized.current) return;

    const initializeDataSources = async () => {
      try {
        setLoading(true);
        setError(null);

        const sources = await apiClient.getDataSourcesWithPermissions();
        const accessibleSources = filterAccessibleDataSources(sources, user);
        const activeSources = accessibleSources.filter((ds) => ds.is_active);

        setDataSources(activeSources);

        // Auto-select first data source if none selected
        if (!value && activeSources.length > 0) {
          const firstSource = activeSources[0];
          onChange(firstSource.id);
          setExpandedDataSources(new Set([firstSource.id]));

          // Load schema for it (only if not already loaded)
          if (!schemas.has(firstSource.id)) {
            await loadSchema(firstSource.id);
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load data sources');
      } finally {
        setLoading(false);
        hasInitialized.current = true;
      }
    };

    initializeDataSources();
  }, []); // Empty dependency array - run once on mount

  // Setup polling for schema updates
  useEffect(() => {
    if (!value || !isPolling) return;

    // Initial load
    if (!schemas.has(value)) {
      loadSchema(value).catch(err => {
        console.warn('Initial schema load failed:', err);
      });
    }

    // Setup polling interval
    pollingIntervalRef.current = setInterval(() => {
      if (value && !document.hidden) { // Only poll if page is visible
        loadSchema(value).catch(err => {
          console.warn('Polling failed:', err);
        });
      }
    }, POLL_INTERVAL);

    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    };
  }, [value, isPolling]);

  const handleDataSourceChange = async (dataSourceId: string) => {
    onChange(dataSourceId);
    setExpandedDataSources(new Set([dataSourceId]));
    setExpandedSections(new Set(['tables']));

    // Load schema if not already loaded
    if (!schemas.has(dataSourceId)) {
      await loadSchema(dataSourceId);
    }
  };

  const handleSyncNow = async () => {
    if (!value) return;

    try {
      await syncSchema(value);
    } catch (error) {
      console.error('Failed to sync schema:', error);
    }
  };

  const toggleDataSource = (dataSourceId: string) => {
    setExpandedDataSources((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(dataSourceId)) {
        newSet.delete(dataSourceId);
      } else {
        newSet.add(dataSourceId);
      }
      return newSet;
    });
  };

  const toggleSection = (section: string) => {
    setExpandedSections((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(section)) {
        newSet.delete(section);
      } else {
        newSet.add(section);
      }
      return newSet;
    });
  };

  const handleTableClick = (tableName: string) => {
    if (onTableSelect) {
      onTableSelect(tableName);
    }
  };

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
          onClick={() => {
            hasInitialized.current = false;
            window.location.reload();
          }}
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
          No accessible data sources available. Please contact an administrator to get access.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Data Source Selector */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <label
            htmlFor="datasource-select"
            className="block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            Data Source
          </label>

          {/* Sync Now Button */}
          <button
            onClick={handleSyncNow}
            disabled={!value || isSchemaLoading}
            className="flex items-center gap-1 px-3 py-1 text-xs font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:hover:bg-blue-900/30 rounded disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Force schema refresh from database"
          >
            <ArrowPathIcon className={`h-3.5 w-3.5 ${isSchemaLoading ? 'animate-spin' : ''}`} />
            Sync Now
          </button>
        </div>
        <select
          id="datasource-select"
          value={value}
          onChange={(e) => handleDataSourceChange(e.target.value)}
          disabled={disabled}
          className="block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-800 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {dataSources.map((ds) => (
            <option key={ds.id} value={ds.id}>
              {ds.name} ({ds.type})
            </option>
          ))}
        </select>
      </div>

      {/* Schema Browser for Selected Data Source */}
      {value && (
        <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <div className="bg-gray-50 dark:bg-gray-900 px-4 py-2 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                Schema
              </h3>
              <div className="flex items-center gap-3">
                {value && lastSyncTime.get(value) && (
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    Last sync: {lastSyncTime.get(value)!.toLocaleTimeString()}
                  </span>
                )}
                {isSchemaLoading && (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
                )}
              </div>
            </div>
          </div>

          <div className="max-h-96 overflow-y-auto">
            {(() => {
              const currentSchema = value ? schemas.get(value) : null;
              if (!currentSchema) return null;

              const tables = currentSchema?.tables.filter(t => !t.table_type || t.table_type === 'table') || [];
              const views = currentSchema?.views || currentSchema?.tables.filter(t => t.table_type === 'view') || [];
              const functions = currentSchema?.functions || [];

              return (
                <>
                  {/* Tables Section */}
                  {tables.length > 0 && (
                    <SchemaSection
                      title="Tables"
                      icon={<TableCellsIcon className="h-4 w-4" />}
                      count={tables.length}
                      isExpanded={expandedSections.has('tables')}
                      onToggle={() => toggleSection('tables')}
                    >
                      {tables.map((table: any) => (
                        <SchemaTableItem
                          key={table.table_name}
                          table={table}
                          onTableSelect={handleTableClick}
                        />
                      ))}
                    </SchemaSection>
                  )}

                  {/* Views Section */}
                  {views.length > 0 && (
                    <SchemaSection
                      title="Views"
                      icon={<EyeIcon className="h-4 w-4" />}
                      count={views.length}
                      isExpanded={expandedSections.has('views')}
                      onToggle={() => toggleSection('views')}
                    >
                      {views.map((view: any) => {
                        const viewName = (view as any).view_name || (view as any).table_name || '';
                        return (
                          <SchemaViewItem
                            key={viewName}
                            view={view}
                            viewName={viewName}
                            onTableSelect={handleTableClick}
                          />
                        );
                      })}
                    </SchemaSection>
                  )}

                  {/* Functions Section */}
                  {functions.length > 0 && (
                    <SchemaSection
                      title="Functions"
                      icon={<ChevronDoubleRightIcon className="h-4 w-4" />}
                      count={functions.length}
                      isExpanded={expandedSections.has('functions')}
                      onToggle={() => toggleSection('functions')}
                    >
                      {functions.map((func: any) => (
                        <div key={func.function_name} className="px-6 py-2 text-sm hover:bg-gray-50 dark:hover:bg-gray-700/50">
                          <div className="flex items-center gap-2">
                            <ServerIcon className="h-4 w-4 text-gray-400" />
                            <span className="font-mono text-gray-700 dark:text-gray-200">
                              {func.function_name}
                            </span>
                            <span className="text-xs text-gray-500">
                              ({func.return_type || 'void'})
                            </span>
                          </div>
                          {func.parameters && (
                            <div className="ml-6 text-xs text-gray-500 dark:text-gray-400">
                              {func.parameters}
                            </div>
                          )}
                        </div>
                      ))}
                    </SchemaSection>
                  )}

                  {/* No Results */}
                  {tables.length === 0 && views.length === 0 && functions.length === 0 && (
                    <div className="p-4 text-center text-sm text-gray-500 dark:text-gray-400">
                      No schema information available
                    </div>
                  )}
                </>
              );
            })()}
          </div>
        </div>
      )}
    </div>
  );
}

// Schema Section Component
interface SchemaSectionProps {
  title: string;
  icon: React.ReactNode;
  count: number;
  isExpanded: boolean;
  onToggle: () => void;
  children: React.ReactNode;
}

function SchemaSection({ title, icon, count, isExpanded, onToggle, children }: SchemaSectionProps) {
  return (
    <div className="border-b border-gray-200 dark:border-gray-700 last:border-b-0">
      <button
        onClick={onToggle}
        className="w-full px-4 py-2 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-700/50"
      >
        <div className="flex items-center gap-2">
          {isExpanded ? (
            <ChevronDownIcon className="h-4 w-4 text-gray-500 dark:text-gray-400" />
          ) : (
            <ChevronRightIcon className="h-4 w-4 text-gray-500 dark:text-gray-400" />
          )}
          <div className="text-gray-600 dark:text-gray-300">{icon}</div>
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
            {title}
          </span>
          <span className="text-xs text-gray-500 dark:text-gray-400">
            ({count})
          </span>
        </div>
      </button>
      {isExpanded && <div className="divide-y divide-gray-200 dark:divide-gray-700">{children}</div>}
    </div>
  );
}

// Table Item Component
interface SchemaTableItemProps {
  table: any;
  onTableSelect?: (tableName: string) => void;
}

function SchemaTableItem({ table, onTableSelect }: SchemaTableItemProps) {
  return (
    <div className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <button
        onClick={() => onTableSelect && onTableSelect(table.table_name)}
        className="w-full px-6 py-2 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${table.table_name} LIMIT 100`}
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <TableCellsIcon className="h-4 w-4 text-gray-400" />
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">
            {table.table_name}
          </span>
          {table.columns && (
            <span className="text-xs text-gray-500 dark:text-gray-400">
              ({table.columns.length} cols)
            </span>
          )}
        </div>
      </button>
    </div>
  );
}

// View Item Component
interface SchemaViewItemProps {
  view: any;
  viewName: string;
  onTableSelect?: (tableName: string) => void;
}

function SchemaViewItem({ view, viewName, onTableSelect }: SchemaViewItemProps) {
  return (
    <div className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <button
        onClick={() => onTableSelect && onTableSelect(viewName)}
        className="w-full px-6 py-2 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${viewName} LIMIT 100`}
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <EyeIcon className="h-4 w-4 text-gray-400" />
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">
            {viewName}
          </span>
          <span className="px-1.5 py-0.5 text-xs bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded">
            VIEW
          </span>
        </div>
      </button>
    </div>
  );
}
