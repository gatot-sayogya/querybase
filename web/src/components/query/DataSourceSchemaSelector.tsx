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
import Button from '@/components/ui/Button';
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
  onWritePermissionChange?: (canWrite: boolean) => void;
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
  onWritePermissionChange,
}: DataSourceSchemaSelectorProps) {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedDataSources, setExpandedDataSources] = useState<Set<string>>(new Set());
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['tables']));
  const [isPolling, setIsPolling] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [isSearchHovered, setIsSearchHovered] = useState(false);
  const searchBarRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
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

  // Compute write permission when value or dataSources change
  useEffect(() => {
    if (value && dataSources.length > 0 && onWritePermissionChange) {
      const selectedDs = dataSources.find((ds) => ds.id === value);
      if (selectedDs) {
        const { canWriteToDataSource } = require('@/lib/data-source-utils');
        onWritePermissionChange(canWriteToDataSource(selectedDs, user));
      }
    }
  }, [value, dataSources, user, onWritePermissionChange]);

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

  // Pagination state for schema items
  const [tableLimit, setTableLimit] = useState(50);
  const [viewLimit, setViewLimit] = useState(50);
  const [functionLimit, setFunctionLimit] = useState(50);

  const handleDataSourceChange = async (dataSourceId: string) => {
    onChange(dataSourceId);
    setExpandedDataSources(new Set([dataSourceId]));
    setExpandedSections(new Set(['tables']));
    setTableLimit(50);
    setViewLimit(50);
    setFunctionLimit(50);

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
      <div className="animate-pulse px-2">
        <div className="h-10 bg-slate-200 dark:bg-slate-700/50 rounded-2xl mb-4"></div>
        <div className="h-64 bg-slate-200 dark:bg-slate-700/50 rounded-2xl"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-2xl">
        <p className="text-sm text-red-600 dark:text-red-400 font-medium">{error}</p>
        <button
          onClick={() => {
            hasInitialized.current = false;
            window.location.reload();
          }}
          className="mt-2 text-xs text-red-600 dark:text-red-400 underline font-bold"
        >
          RETRY
        </button>
      </div>
    );
  }

  if (dataSources.length === 0) {
    return (
      <div className="p-4 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-2xl">
        <p className="text-sm text-amber-600 dark:text-amber-400 font-medium">
          No accessible data sources available. Please contact an administrator to get access.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4 flex flex-col flex-1 overflow-hidden">
      {/* Data Source Selector */}
      <div className="flex-shrink-0 mb-3 px-1" ref={dropdownRef}>
        <div className="flex items-center gap-2">
          {/* Custom Dropdown with Health Indicators */}
          <div className="relative flex-1">
            <button
              type="button"
              onClick={() => !disabled && setDropdownOpen(!dropdownOpen)}
              disabled={disabled}
              className="w-full flex items-center justify-between px-3 py-2 text-sm border-none rounded-2xl sleek-shadow-sm bg-white/50 dark:bg-slate-800/50 backdrop-blur-md text-slate-900 dark:text-white hover:bg-white/80 dark:hover:bg-slate-800/80 transition-all ease-spring disabled:opacity-50 disabled:cursor-not-allowed group"
            >
              <span className="flex items-center gap-2 min-w-0 flex-1">
                {value && <HealthDot status={healthStatuses[value] || 'unknown'} />}
                <span className="truncate font-bold tracking-tight">
                  {value
                    ? (dataSources.find(ds => ds.id === value)?.name ?? 'Select datasource')
                    : 'Select a data source'}
                </span>
              </span>
              <ChevronDownIcon className={`h-4 w-4 text-slate-400 group-hover:text-slate-600 transition-all ease-spring duration-300 ${dropdownOpen ? 'rotate-180' : ''}`} />
            </button>

            {dropdownOpen && (
              <div className="absolute z-30 mt-2 w-full glass sleek-shadow rounded-2xl overflow-hidden animate-in fade-in slide-in-from-top-2 duration-300 border border-white/20 dark:border-white/5">
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
                      className={`w-full flex items-center gap-3 px-3 py-2.5 text-left text-xs hover:bg-slate-500/10 dark:hover:bg-white/10 transition-colors ${
                        isSelected ? 'bg-blue-500/10 dark:bg-blue-500/20' : ''
                      }`}
                    >
                      <HealthDot status={status} />
                      <span className="flex-1 truncate font-bold text-slate-700 dark:text-slate-100">{ds.name}</span>
                      <span className="flex-shrink-0 opacity-40 group-hover:opacity-100 transition-opacity" 
                            style={{
                               color: isPg ? '#3B82F6' : isMysql ? '#10B981' : 'inherit'
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

          {/* Sync Now Button — moved inline */}
          <button
            onClick={handleSyncNow}
            disabled={!value || isSchemaLoading}
            className="flex-shrink-0 flex items-center justify-center w-10 h-[38px] text-slate-500 bg-white/50 dark:bg-slate-800/50 hover:bg-white/80 dark:hover:bg-slate-800/80 rounded-2xl sleek-shadow-sm transition-all ease-spring disabled:opacity-50 disabled:cursor-not-allowed border border-white/20 dark:border-white/5"
            title="Sync schema and re-check datasource health"
          >
            <ArrowPathIcon className={`h-4 w-4 ${isSchemaLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Schema Browser for Selected Data Source */}
      {value && (
        <div className="glass rounded-3xl sleek-shadow border border-white/20 dark:border-white/5 overflow-hidden flex flex-col flex-1 animate-in fade-in slide-in-from-bottom-2 duration-500">
          <div className="bg-white/30 dark:bg-slate-800/30 px-3 py-3 border-b border-slate-200/50 dark:border-white/10 flex-shrink-0 backdrop-blur-sm">
            <div className={`flex items-center justify-between group/search px-1 h-6 transition-all duration-500 ease-spring ${isSearchHovered || searchTerm ? 'gap-0' : 'gap-2'}`}
                 onMouseEnter={() => setIsSearchHovered(true)}
                 onMouseLeave={() => !searchTerm && !document.activeElement?.className.includes('search-input') && setIsSearchHovered(false)}>
              
              <h3 className={`text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 transition-all duration-500 whitespace-nowrap overflow-hidden ${isSearchHovered || searchTerm ? 'opacity-0 w-0' : 'opacity-100 w-auto'}`}>
                Schema Explorer
              </h3>

              <div className={`relative flex items-center transition-all duration-500 ease-spring ${isSearchHovered || searchTerm ? 'flex-1 translate-x-0' : 'w-6 translate-x-1'}`}>
                <MagnifyingGlassIcon className={`absolute left-2 h-3.5 w-3.5 text-slate-400 transition-opacity duration-300 ${isSearchHovered || searchTerm ? 'opacity-100' : 'opacity-0'}`} />
                
                <input
                  ref={inputRef}
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onFocus={() => setIsSearchHovered(true)}
                  onBlur={() => !searchTerm && setIsSearchHovered(false)}
                  placeholder="Search schema..."
                  className={`search-input w-full bg-slate-500/5 dark:bg-white/5 border-none focus:ring-0 rounded-xl py-1 pl-7 pr-8 text-[10px] font-bold text-slate-700 dark:text-white transition-all duration-500 placeholder:text-slate-400/50 ${isSearchHovered || searchTerm ? 'opacity-100 w-full' : 'opacity-0 w-0'}`}
                />

                {!isSearchHovered && !searchTerm && (
                  <button className="absolute right-0 p-1 text-slate-400 hover:text-blue-500 transition-colors">
                    <MagnifyingGlassIcon className="h-4 w-4" />
                  </button>
                )}

                {searchTerm && (
                  <button 
                    onClick={() => {
                      setSearchTerm('');
                      setIsSearchHovered(false);
                      inputRef.current?.blur();
                    }}
                    className="absolute right-2 p-0.5 text-slate-400 hover:text-red-500 transition-colors"
                  >
                    <XMarkIcon className="h-3 w-3" />
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto custom-scrollbar">
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
                      <div className="flex flex-col">
                        {filteredTables.slice(0, tableLimit).map((table: any) => (
                          <SchemaTableItem
                            key={table.table_name}
                            table={table}
                            onTableSelect={handleTableClick}
                          />
                        ))}
                        {filteredTables.length > tableLimit && (
                          <button
                            onClick={() => setTableLimit(prev => prev + 50)}
                            className="w-full py-2 px-8 text-left text-[10px] font-bold text-blue-500 hover:text-blue-600 transition-colors uppercase tracking-widest bg-slate-500/5"
                          >
                            + Show {Math.min(50, filteredTables.length - tableLimit)} More Tables
                          </button>
                        )}
                      </div>
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
                      <div className="flex flex-col">
                        {filteredViews.slice(0, viewLimit).map((view: any) => {
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
                        {filteredViews.length > viewLimit && (
                          <button
                            onClick={() => setViewLimit(prev => prev + 50)}
                            className="w-full py-2 px-8 text-left text-[10px] font-bold text-purple-500 hover:text-purple-600 transition-colors uppercase tracking-widest bg-slate-500/5"
                          >
                            + Show {Math.min(50, filteredViews.length - viewLimit)} More Views
                          </button>
                        )}
                      </div>
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
                      <div className="flex flex-col">
                        {filteredFunctions.slice(0, functionLimit).map((func: any) => (
                          <div key={func.function_name} className="px-6 py-2.5 text-xs hover:bg-slate-500/10 dark:hover:bg-white/10 cursor-pointer group transition-colors border-b border-white/5 last:border-none">
                            <div className="flex items-center gap-2">
                              <ServerIcon className="h-3.5 w-3.5 text-slate-400 group-hover:text-blue-500 transition-colors" />
                              <span className="text-slate-600 dark:text-slate-300 font-bold group-hover:text-slate-900 dark:group-hover:text-white transition-colors truncate">
                                {func.function_name}
                              </span>
                              <span className="text-[9px] font-bold text-slate-400 tracking-tighter">
                                {func.return_type || 'void'}
                              </span>
                            </div>
                            {func.parameters && (
                              <div className="ml-5 mt-0.5 text-[9px] leading-tight text-slate-400 group-hover:text-slate-500 transition-colors truncate">
                                {func.parameters}
                              </div>
                            )}
                          </div>
                        ))}
                        {filteredFunctions.length > functionLimit && (
                          <button
                            onClick={() => setFunctionLimit(prev => prev + 50)}
                            className="w-full py-2 px-8 text-left text-[10px] font-bold text-emerald-500 hover:text-emerald-600 transition-colors uppercase tracking-widest bg-slate-500/5"
                          >
                            + Show {Math.min(50, filteredFunctions.length - functionLimit)} More Functions
                          </button>
                        )}
                      </div>
                    </SchemaSection>
                  )}

                  {/* No Results */}
                  {filteredTables.length === 0 && filteredViews.length === 0 && filteredFunctions.length === 0 && (
                    <div className="p-8 text-center text-xs font-medium text-slate-400 italic">
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
    <div className="border-b border-slate-200/50 dark:border-white/5 last:border-b-0">
      <button
        onClick={onToggle}
        className="w-full px-3 py-2 flex items-center justify-between text-left hover:bg-slate-500/5 dark:hover:bg-white/5 transition-colors"
      >
        <div className="flex items-center gap-2">
          {isExpanded ? (
            <ChevronDownIcon className="h-3 w-3 text-slate-400" />
          ) : (
            <ChevronRightIcon className="h-3 w-3 text-slate-400" />
          )}
          <div className="text-slate-500 scale-90">{icon}</div>
          <span className="text-[11px] font-bold text-slate-700 dark:text-slate-300 tracking-wide">
            {title} <span className="text-slate-400 font-medium">({count})</span>
          </span>
        </div>
      </button>
      {isExpanded && <div className="divide-y divide-slate-100 dark:divide-white/5 bg-white/10 dark:bg-black/10">{children}</div>}
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
    <div className="hover:bg-slate-500/10 dark:hover:bg-white/10 cursor-pointer group transition-colors">
      <button
        onClick={() => onTableSelect && onTableSelect(table.table_name)}
        className="w-full pl-8 pr-3 py-2 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${table.table_name} LIMIT 100`}
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <TableCellsIcon className="h-3 w-3 text-slate-300 dark:text-slate-600 group-hover:text-blue-500 transition-colors" />
          <span className="text-[11px] font-medium text-slate-600 dark:text-slate-300 group-hover:text-slate-900 dark:group-hover:text-white truncate transition-colors">
            {table.table_name}
          </span>
          {table.columns && (
            <span className="text-[9px] text-slate-400 group-hover:text-slate-500 flex-shrink-0 transition-colors">
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
    <div className="hover:bg-slate-500/10 dark:hover:bg-white/10 cursor-pointer group transition-colors">
      <button
        onClick={() => onTableSelect && onTableSelect(viewName)}
        className="w-full pl-8 pr-3 py-2 flex items-center justify-between text-left"
        title={`Click to view data: SELECT * FROM ${viewName} LIMIT 100`}
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <EyeIcon className="h-3 w-3 text-slate-300 dark:text-slate-600 group-hover:text-purple-500 transition-colors" />
          <span className="text-xs text-gray-700 dark:text-gray-200 group-hover:text-gray-900 dark:group-hover:text-white truncate transition-colors">
            {viewName}
          </span>
        <span className="px-1 py-0 text-[8px] bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300 rounded flex-shrink-0">
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
