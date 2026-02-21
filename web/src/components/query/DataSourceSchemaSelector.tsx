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

type HealthStatus = 'healthy' | 'unhealthy' | 'checking' | 'unknown';

const POLL_INTERVAL = 60000; // 60 seconds

function PgIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12.022 2.1c-5.467 0-9.878 4.41-9.878 9.88 0 5.467 4.411 9.878 9.878 9.878 5.467 0 9.878-4.411 9.878-9.878 0-5.47-4.411-9.88-9.878-9.88zm3.623 14.88c-.5.441-1.294.67-2.323.67H9.255v1.233h-1.97v-8.498h6.143c.97 0 1.705.235 2.176.64.441.41.676.97.676 1.734.025.794-.236 1.352-.647 1.734-.383.353-.941.529-1.646.529h-3.41v1.94h4.41v-3.41h1.56v3.438zM8.344 8.784h4.086c.764 0 1.352.176 1.735.47.382.264.588.675.588 1.146 0 .47-.206.882-.588 1.205-.383.294-.971.47-1.735.47H8.344v-3.29z"/>
    </svg>
  );
}

function MysqlIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12.002 1.357c-5.836 0-10.56 4.721-10.56 10.551 0 5.828 4.724 10.55 10.55 10.55 5.833 0 10.554-4.722 10.554-10.55 0-5.83-4.721-10.551-10.554-10.551zm1.264 6.646c1.373 0 1.956.892 1.956 1.94 0 1.135-.615 1.983-1.897 1.983-.756 0-1.284-.336-1.574-.75h-.06v3.593H9.863V8.11h1.616v.612h.063c.27-.406.804-.719 1.724-.719zm-1.028 1.48c-.68 0-1.042.47-1.042 1.055 0 .564.364 1.026 1.053 1.026.685 0 1.05-.44 1.05-1.042 0-.616-.365-1.038-1.06-1.038zM5.385 15.65h1.83V9.752H5.385v5.898zm.9-8c-.658 0-1.127.47-1.127 1.114 0 .647.469 1.116 1.127 1.116.634 0 1.077-.47 1.077-1.116 0-.645-.443-1.114-1.076-1.114zm11.332 7.822c-1.375 0-1.958-.893-1.958-1.94 0-1.133.616-1.984 1.898-1.984.757 0 1.285.337 1.575.75h.06V8.11h1.828v7.362h-1.616v-.613h-.063c-.27.406-.804.72-1.724.72zm1.028-1.48c.682 0 1.043-.466 1.043-1.053 0-.564-.361-1.027-1.05-1.027-.685 0-1.05.441-1.05 1.043 0 .615.365 1.037 1.058 1.037z"/>
    </svg>
  );
}

function DbIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <ellipse cx="12" cy="5" rx="9" ry="3"/>
      <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
      <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
    </svg>
  );
}

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
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [healthStatuses, setHealthStatuses] = useState<Record<string, HealthStatus>>({});
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
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

        // Run health checks first, then auto-select the first healthy datasource
        if (!value && activeSources.length > 0) {
          // Check health of all datasources in parallel
          const healthResults = await Promise.all(
            activeSources.map(async (ds) => {
              try {
                const h = await apiClient.getDataSourceHealth(ds.id);
                return { id: ds.id, healthy: h.status === 'healthy' || h.status === 'degraded' };
              } catch {
                return { id: ds.id, healthy: false };
              }
            })
          );

          // Update health statuses in state
          const statusMap: Record<string, HealthStatus> = {};
          healthResults.forEach(({ id, healthy }) => {
            statusMap[id] = healthy ? 'healthy' : 'unhealthy';
          });
          setHealthStatuses(statusMap);

          // Pick first healthy datasource, fall back to first if all are unhealthy
          const healthyFirst =
            activeSources.find(ds => statusMap[ds.id] === 'healthy') ?? activeSources[0];

          onChange(healthyFirst.id);
          setExpandedDataSources(new Set([healthyFirst.id]));

          if (!schemas.has(healthyFirst.id)) {
            await loadSchema(healthyFirst.id);
          }
        } else {
          // Health checks in background when a datasource is already selected
          fetchHealthStatuses(activeSources);
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

  // Fetch health status for all datasources in parallel
  const fetchHealthStatuses = async (sources: DataSource[]) => {
    const checking: Record<string, HealthStatus> = {};
    sources.forEach(ds => { checking[ds.id] = 'checking'; });
    setHealthStatuses(prev => ({ ...prev, ...checking }));

    await Promise.all(sources.map(async (ds) => {
      try {
        const result = await apiClient.getDataSourceHealth(ds.id);
        const status: HealthStatus =
          result?.status === 'healthy' ? 'healthy' :
          result?.status === 'degraded' ? 'checking' : // yellow = degraded/slow
          'unhealthy';
        setHealthStatuses(prev => ({ ...prev, [ds.id]: status }));
      } catch {
        setHealthStatuses(prev => ({ ...prev, [ds.id]: 'unhealthy' }));
      }
    }));
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

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
      // Also refresh health status on sync
      fetchHealthStatuses(dataSources);
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
      <div className="flex-shrink-0 mb-1" ref={dropdownRef}>
        <div className="flex items-center gap-1">
          {/* Custom Dropdown with Health Indicators */}
          <div className="relative flex-1">
          <button
            type="button"
            onClick={() => !disabled && setDropdownOpen(!dropdownOpen)}
            disabled={disabled}
            className="w-full flex items-center justify-between px-2 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded shadow-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="flex items-center gap-1.5 min-w-0 flex-1">
              {value && <HealthDot status={healthStatuses[value] || 'unknown'} />}
              <span className="truncate">
                {value
                  ? (dataSources.find(ds => ds.id === value)?.name ?? 'Select datasource')
                  : 'Select a data source'}
              </span>
            </span>
            <ChevronDownIcon className={`h-3 w-3 text-gray-400 flex-shrink-0 transition-transform ${dropdownOpen ? 'rotate-180' : ''}`} />
          </button>

          {dropdownOpen && (
            <div className="absolute z-30 mt-1 w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-lg">
              {dataSources.map((ds) => {
                const status = healthStatuses[ds.id] || 'unknown';
                const isSelected = ds.id === value;
                const isPg = ds.type === 'postgresql';
                const isMysql = ds.type === 'mysql';
                
                return (
                  <button
                    key={ds.id}
                    type="button"
                    onClick={() => {
                      handleDataSourceChange(ds.id);
                      setDropdownOpen(false);
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 text-left text-xs hover:bg-gray-50 dark:hover:bg-gray-700 ${
                      isSelected ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                    }`}
                  >
                    <HealthDot status={status} />
                    <span className="flex-1 truncate font-medium text-gray-800 dark:text-gray-100">{ds.name}</span>
                    <span className="flex-shrink-0 transition-opacity opacity-80" 
                          style={{
                             color: isPg ? '#1D4ED8' : isMysql ? '#166534' : 'inherit'
                          }}>
                      {isPg ? <PgIcon /> : isMysql ? <MysqlIcon /> : <DbIcon />}
                    </span>
                    <HealthBadge status={status} />
                  </button>
                );
              })}
            </div>
          )}
        </div>

        {/* Sync Now Button — also refreshes health */}
        <button
          onClick={handleSyncNow}
          disabled={!value || isSchemaLoading}
          className="flex-shrink-0 flex items-center justify-center p-1.5 text-blue-600 bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:hover:bg-blue-900/30 rounded border border-blue-100 dark:border-blue-900/30 disabled:opacity-50 disabled:cursor-not-allowed"
          title="Sync schema and re-check datasource health"
        >
          <ArrowPathIcon className={`h-4 w-4 ${isSchemaLoading ? 'animate-spin' : ''}`} />
        </button>
        </div>
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
                        <div key={func.function_name} className="px-6 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer group transition-colors">
                          <div className="flex items-center gap-2">
                            <ServerIcon className="h-4 w-4 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors" />
                            <span className="text-gray-700 dark:text-gray-200 group-hover:text-gray-900 dark:group-hover:text-white transition-colors">
                              {func.function_name}
                            </span>
                            <span className="text-xs text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-400 transition-colors">
                              ({func.return_type || 'void'})
                            </span>
                          </div>
                          {func.parameters && (
                            <div className="ml-6 text-xs text-gray-500 dark:text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors">
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
    <div className="hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer group transition-colors">
      <button
        onClick={() => onTableSelect && onTableSelect(table.table_name)}
        className="w-full px-3 py-1 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${table.table_name} LIMIT 100`}
      >
        <div className="flex items-center gap-1 flex-1 min-w-0">
          <TableCellsIcon className="h-3 w-3 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 flex-shrink-0 transition-colors" />
          <span className="text-xs text-gray-700 dark:text-gray-200 group-hover:text-gray-900 dark:group-hover:text-white truncate transition-colors">
            {table.table_name}
          </span>
          {table.columns && (
            <span className="text-[9px] text-gray-500 dark:text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 flex-shrink-0 transition-colors">
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
    <div className="hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer group transition-colors">
      <button
        onClick={() => onTableSelect && onTableSelect(viewName)}
        className="w-full px-3 py-1 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${viewName} LIMIT 100`}
      >
        <div className="flex items-center gap-1 flex-1 min-w-0">
          <EyeIcon className="h-3 w-3 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 flex-shrink-0 transition-colors" />
          <span className="text-xs text-gray-700 dark:text-gray-200 group-hover:text-gray-900 dark:group-hover:text-white truncate transition-colors">
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

// Health Dot — small colored circle
function HealthDot({ status }: { status: HealthStatus }) {
  const colors: Record<HealthStatus, string> = {
    healthy: 'bg-green-500',
    unhealthy: 'bg-red-500',
    checking: 'bg-yellow-400 animate-pulse',
    unknown: 'bg-gray-300',
  };
  return <span className={`inline-block w-2 h-2 rounded-full flex-shrink-0 ${colors[status]}`} />;
}

// Health Badge — small label pill
function HealthBadge({ status }: { status: HealthStatus }) {
  if (status === 'unknown') return null;
  const styles: Record<HealthStatus, string> = {
    healthy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    unhealthy: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    checking: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    unknown: '',
  };
  const labels: Record<HealthStatus, string> = {
    healthy: 'OK',
    unhealthy: 'Error',
    checking: '...',
    unknown: '',
  };
  return (
    <span className={`px-1 py-0 text-[9px] font-semibold rounded flex-shrink-0 ${styles[status]}`}>
      {labels[status]}
    </span>
  );
}
