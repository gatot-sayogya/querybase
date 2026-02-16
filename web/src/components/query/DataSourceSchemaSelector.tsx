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
  MagnifyingGlassIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
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
  const [searchTerm, setSearchTerm] = useState('');
  const [showSearchModal, setShowSearchModal] = useState(false);
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
    <div className="space-y-2 flex flex-col flex-1 overflow-hidden">
      {/* Data Source Selector */}
      <div className="flex-shrink-0">
        <div className="flex items-center justify-between mb-1 px-1">
          <label
            htmlFor="datasource-select"
            className="block text-xs font-medium text-gray-600 dark:text-gray-400"
          >
            Data Source
          </label>

          {/* Sync Now Button */}
          <button
            onClick={handleSyncNow}
            disabled={!value || isSchemaLoading}
            className="flex items-center gap-0.5 px-1.5 py-0.5 text-[10px] font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:hover:bg-blue-900/30 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            title="Force schema refresh from database"
          >
            <ArrowPathIcon className={`h-3 w-3 ${isSchemaLoading ? 'animate-spin' : ''}`} />
            Sync
          </button>
        </div>
        <select
          id="datasource-select"
          value={value}
          onChange={(e) => handleDataSourceChange(e.target.value)}
          disabled={disabled}
          className="block w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-700 rounded shadow-sm focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
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
        <div className="border border-gray-200 dark:border-gray-700 rounded overflow-hidden flex flex-col flex-1">
          <div className="bg-gray-50 dark:bg-gray-900 px-2 py-1 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-semibold text-gray-700 dark:text-gray-300">
                Schema
              </h3>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => setShowSearchModal(true)}
                  className="p-0.5 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded"
                  title="Search tables, views, functions"
                >
                  <MagnifyingGlassIcon className="h-3 w-3" />
                </button>
                {value && lastSyncTime.get(value) && (
                  <span className="text-[9px] text-gray-500 dark:text-gray-400">
                    {lastSyncTime.get(value)!.toLocaleTimeString()}
                  </span>
                )}
                {isSchemaLoading && (
                  <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-blue-600"></div>
                )}
              </div>
            </div>

            {/* Active Search Filter Indicator */}
            {searchTerm && (
              <div className="mt-1 flex items-center gap-1 text-[9px] bg-blue-50 dark:bg-blue-900/20 px-1 py-0.5 rounded">
                <span className="text-blue-700 dark:text-blue-300">
                  Filter: <strong>{searchTerm}</strong>
                </span>
                <button
                  onClick={() => setSearchTerm('')}
                  className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200"
                >
                  <XMarkIcon className="h-2.5 w-2.5" />
                </button>
              </div>
            )}
          </div>

          {/* Search Modal */}
          {showSearchModal && (
            <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 p-4">
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full">
                <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                      Search Database Objects
                    </h3>
                    <button
                      onClick={() => setShowSearchModal(false)}
                      className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                    >
                      <XMarkIcon className="h-5 w-5" />
                    </button>
                  </div>
                </div>
                <div className="p-4">
                  <div className="relative">
                    <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                    <input
                      type="text"
                      placeholder="Search tables, views, functions..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="w-full pl-10 pr-4 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      autoFocus
                    />
                  </div>
                  {searchTerm && (() => {
                    const currentSchema = value ? schemas.get(value) : null;
                    if (!currentSchema) return null;

                    const tables = [...currentSchema.tables.filter(t => !t.table_type || t.table_type === 'table')]
                      .sort((a, b) => a.table_name.localeCompare(b.table_name));
                    const views = [...(currentSchema.views || currentSchema.tables.filter(t => t.table_type === 'view'))]
                      .sort((a, b) => {
                        const nameA = (a as any).view_name || (a as any).table_name || '';
                        const nameB = (b as any).view_name || (b as any).table_name || '';
                        return nameA.localeCompare(nameB);
                      });
                    const functions = [...(currentSchema.functions || [])]
                      .sort((a, b) => a.function_name.localeCompare(b.function_name));

                    const filteredTables = tables.filter(t =>
                      t.table_name.toLowerCase().includes(searchTerm.toLowerCase())
                    );
                    const filteredViews = views.filter(v => {
                      const name = (v as any).view_name || (v as any).table_name || '';
                      return name.toLowerCase().includes(searchTerm.toLowerCase());
                    });
                    const filteredFunctions = functions.filter(f =>
                      f.function_name.toLowerCase().includes(searchTerm.toLowerCase())
                    );

                    return (
                      <div className="mt-4">
                        <p className="text-xs text-gray-600 dark:text-gray-400 mb-2">
                          Found {filteredTables.length + filteredViews.length + filteredFunctions.length} results
                        </p>
                        <div className="max-h-48 overflow-y-auto space-y-1">
                          {filteredTables.length > 0 && (
                            <div>
                              <p className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">Tables</p>
                              {filteredTables.slice(0, 5).map((table) => (
                                <button
                                  key={table.table_name}
                                  onClick={() => {
                                    handleTableClick(table.table_name);
                                    setShowSearchModal(false);
                                  }}
                                  className="w-full text-left px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                                >
                                  📋 {table.table_name}
                                </button>
                              ))}
                              {filteredTables.length > 5 && (
                                <p className="text-xs text-gray-500 dark:text-gray-400 px-3">
                                  ...and {filteredTables.length - 5} more tables
                                </p>
                              )}
                            </div>
                          )}
                          {filteredViews.length > 0 && (
                            <div className="mt-2">
                              <p className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">Views</p>
                              {filteredViews.slice(0, 5).map((view) => {
                                const name = (view as any).view_name || (view as any).table_name;
                                return (
                                  <button
                                    key={name}
                                    onClick={() => {
                                      handleTableClick(name);
                                      setShowSearchModal(false);
                                    }}
                                    className="w-full text-left px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                                  >
                                    👁️ {name}
                                  </button>
                                );
                              })}
                              {filteredViews.length > 5 && (
                                <p className="text-xs text-gray-500 dark:text-gray-400 px-3">
                                  ...and {filteredViews.length - 5} more views
                                </p>
                              )}
                            </div>
                          )}
                          {filteredFunctions.length > 0 && (
                            <div className="mt-2">
                              <p className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">Functions</p>
                              {filteredFunctions.slice(0, 5).map((func) => (
                                <button
                                  key={func.function_name}
                                  onClick={() => {
                                    setShowSearchModal(false);
                                  }}
                                  className="w-full text-left px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                                >
                                  ⚙️ {func.function_name}
                                </button>
                              ))}
                              {filteredFunctions.length > 5 && (
                                <p className="text-xs text-gray-500 dark:text-gray-400 px-3">
                                  ...and {filteredFunctions.length - 5} more functions
                                </p>
                              )}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })()}
                </div>
                <div className="p-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2">
                  {searchTerm && (
                    <button
                      onClick={() => setSearchTerm('')}
                      className="px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded"
                    >
                      Clear Filter
                    </button>
                  )}
                  <button
                    onClick={() => setShowSearchModal(false)}
                    className="px-3 py-1.5 text-sm text-white bg-blue-600 hover:bg-blue-700 rounded"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          )}

          <div className="flex-1 overflow-y-auto">
            {(() => {
              const currentSchema = value ? schemas.get(value) : null;
              if (!currentSchema) return null;

              // Sort tables/views/functions A-Z by default
              const tables = [...(currentSchema?.tables.filter(t => !t.table_type || t.table_type === 'table') || [])]
                .sort((a, b) => a.table_name.localeCompare(b.table_name));
              const views = [...(currentSchema?.views || currentSchema?.tables.filter(t => t.table_type === 'view') || [])]
                .sort((a, b) => {
                  const nameA = (a as any).view_name || (a as any).table_name || '';
                  const nameB = (b as any).view_name || (b as any).table_name || '';
                  return nameA.localeCompare(nameB);
                });
              const functions = [...(currentSchema?.functions || [])]
                .sort((a, b) => a.function_name.localeCompare(b.function_name));

              // Filter based on search
              const filteredTables = tables.filter((table) =>
                table.table_name.toLowerCase().includes(searchTerm.toLowerCase())
              );
              const filteredViews = views.filter((view) => {
                const viewName = (view as any).view_name || (view as any).table_name || '';
                return viewName.toLowerCase().includes(searchTerm.toLowerCase());
              });
              const filteredFunctions = functions.filter((func) =>
                func.function_name.toLowerCase().includes(searchTerm.toLowerCase())
              );

              return (
                <>
                  {/* Tables Section */}
                  {filteredTables.length > 0 && (
                    <SchemaSection
                      title="Tables"
                      icon={<TableCellsIcon className="h-4 w-4" />}
                      count={filteredTables.length}
                      isExpanded={expandedSections.has('tables')}
                      onToggle={() => toggleSection('tables')}
                    >
                      {filteredTables.map((table: any) => (
                        <SchemaTableItem
                          key={table.table_name}
                          table={table}
                          onTableSelect={handleTableClick}
                        />
                      ))}
                    </SchemaSection>
                  )}

                  {/* Views Section */}
                  {filteredViews.length > 0 && (
                    <SchemaSection
                      title="Views"
                      icon={<EyeIcon className="h-4 w-4" />}
                      count={filteredViews.length}
                      isExpanded={expandedSections.has('views')}
                      onToggle={() => toggleSection('views')}
                    >
                      {filteredViews.map((view: any) => {
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
                  {filteredFunctions.length > 0 && (
                    <SchemaSection
                      title="Functions"
                      icon={<ChevronDoubleRightIcon className="h-4 w-4" />}
                      count={filteredFunctions.length}
                      isExpanded={expandedSections.has('functions')}
                      onToggle={() => toggleSection('functions')}
                    >
                      {filteredFunctions.map((func: any) => (
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
                  {filteredTables.length === 0 && filteredViews.length === 0 && filteredFunctions.length === 0 && (
                    <div className="p-4 text-center text-sm text-gray-500 dark:text-gray-400">
                      {searchTerm ? `No results found for "${searchTerm}"` : 'No schema information available'}
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
        className="w-full px-2 py-1 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-700/50"
      >
        <div className="flex items-center gap-1">
          {isExpanded ? (
            <ChevronDownIcon className="h-3 w-3 text-gray-500 dark:text-gray-400" />
          ) : (
            <ChevronRightIcon className="h-3 w-3 text-gray-500 dark:text-gray-400" />
          )}
          <div className="text-gray-600 dark:text-gray-300 scale-75">{icon}</div>
          <span className="text-xs font-medium text-gray-700 dark:text-gray-200">
            {title}
          </span>
          <span className="text-[10px] text-gray-500 dark:text-gray-400">
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
        className="w-full px-3 py-0.5 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${table.table_name} LIMIT 100`}
      >
        <div className="flex items-center gap-1 flex-1 min-w-0">
          <TableCellsIcon className="h-3 w-3 text-gray-400 flex-shrink-0" />
          <span className="text-xs text-gray-700 dark:text-gray-200 truncate">
            {table.table_name}
          </span>
          {table.columns && (
            <span className="text-[9px] text-gray-500 dark:text-gray-400 flex-shrink-0">
              ({table.columns.length})
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
        className="w-full px-3 py-0.5 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${viewName} LIMIT 100`}
      >
        <div className="flex items-center gap-1 flex-1 min-w-0">
          <EyeIcon className="h-3 w-3 text-gray-400 flex-shrink-0" />
          <span className="text-xs text-gray-700 dark:text-gray-200 truncate">
            {viewName}
          </span>
          <span className="px-1 py-0 text-[8px] bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded flex-shrink-0">
            VIEW
          </span>
        </div>
      </button>
    </div>
  );
}
