'use client';

import { useEffect, useState } from 'react';
import {
  ChevronDownIcon,
  ChevronRightIcon,
  TableCellsIcon,
  EyeIcon,
  CodeBracketIcon,
  MagnifyingGlassIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { useSchemaStore } from '@/stores/schema-store';
import { useAuthStore } from '@/stores/auth-store';
import type { TableInfo, SchemaColumnInfo, ViewInfo, FunctionInfo } from '@/types';
import DataSourceSelector from './DataSourceSelector';

interface SchemaBrowserProps {
  onTableSelect?: (tableName: string) => void;
}

export default function SchemaBrowser({ onTableSelect }: SchemaBrowserProps) {
  const { isAuthenticated } = useAuthStore();
  const [selectedDataSource, setSelectedDataSource] = useState<string | null>(null);
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState('');
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['tables']));
  const [showSearchModal, setShowSearchModal] = useState(false);

  const {
    schemas,
    loadSchema,
    isLoading,
    error,
    setCurrentDataSource,
  } = useSchemaStore();

  useEffect(() => {
    if (selectedDataSource && isAuthenticated) {
      setCurrentDataSource(selectedDataSource);
      loadSchema(selectedDataSource);
    }
  }, [selectedDataSource, isAuthenticated]);

  const currentSchema = selectedDataSource ? schemas.get(selectedDataSource) : null;

  // Separate tables from views if table_type is available, SORT A-Z by default
  const tables = [...(currentSchema?.tables.filter(t => !t.table_type || t.table_type === 'table') || [])]
    .sort((a, b) => a.table_name.localeCompare(b.table_name));
  const views = [...(currentSchema?.views || currentSchema?.tables.filter(t => t.table_type === 'view') || [])]
    .sort((a, b) => {
      const nameA = (a as ViewInfo).view_name || (a as TableInfo).table_name || '';
      const nameB = (b as ViewInfo).view_name || (b as TableInfo).table_name || '';
      return nameA.localeCompare(nameB);
    });
  const functions = [...(currentSchema?.functions || [])]
    .sort((a, b) => a.function_name.localeCompare(b.function_name));

  // Filter items based on search
  const filteredTables = tables.filter((table) =>
    table.table_name.toLowerCase().includes(searchTerm.toLowerCase())
  );
  const filteredViews = views.filter((view) => {
    const viewName = (view as ViewInfo).view_name || (view as TableInfo).table_name || '';
    return viewName.toLowerCase().includes(searchTerm.toLowerCase());
  });
  const filteredFunctions = functions.filter((func) =>
    func.function_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const toggleItem = (itemName: string) => {
    setExpandedItems((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(itemName)) {
        newSet.delete(itemName);
      } else {
        newSet.add(itemName);
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

  const expandAll = () => {
    if (currentSchema) {
      const allItems = [
        ...tables.map(t => t.table_name),
        ...views.map(v => (v as ViewInfo).view_name || (v as TableInfo).table_name || ''),
        ...functions.map(f => f.function_name)
      ].filter(Boolean);
      setExpandedItems(new Set(allItems));
      setExpandedSections(new Set(['tables', 'views', 'functions']));
    }
  };

  const collapseAll = () => {
    setExpandedItems(new Set());
    setExpandedSections(new Set(['tables']));
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-800">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-3">Database Schema</h3>
        <DataSourceSelector
          value={selectedDataSource || ''}
          onChange={setSelectedDataSource}
        />
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="p-4 m-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}

      {/* No Data Source Selected */}
      {!selectedDataSource && !isLoading && (
        <div className="flex items-center justify-center p-8 text-center">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Select a data source to view its schema
          </p>
        </div>
      )}

      {/* Schema Browser */}
      {selectedDataSource && currentSchema && !isLoading && (
        <>
          {/* Schema Info & Controls */}
          <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <p className="text-xs font-medium text-gray-600 dark:text-gray-300">
                  {currentSchema.data_source_name}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  {currentSchema.database_type} • {tables.length} tables, {views.length} views, {functions.length} functions
                </p>
              </div>
              <div className="flex gap-1">
                <button
                  onClick={() => setShowSearchModal(true)}
                  className="p-1.5 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded"
                  title="Search tables, views, functions"
                >
                  <MagnifyingGlassIcon className="h-4 w-4" />
                </button>
                <button
                  onClick={expandAll}
                  className="px-2 py-1 text-xs text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded"
                  title="Expand all"
                >
                  Expand
                </button>
                <button
                  onClick={collapseAll}
                  className="px-2 py-1 text-xs text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded"
                  title="Collapse all"
                >
                  Collapse
                </button>
              </div>
            </div>

            {/* Active Search Filter Indicator */}
            {searchTerm && (
              <div className="mt-2 flex items-center gap-2 text-xs bg-blue-50 dark:bg-blue-900/20 px-2 py-1 rounded">
                <span className="text-blue-700 dark:text-blue-300">
                  Filter: <strong>{searchTerm}</strong>
                </span>
                <button
                  onClick={() => setSearchTerm('')}
                  className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200"
                >
                  <XMarkIcon className="h-3 w-3" />
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
                  {searchTerm && (
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
                                  toggleItem(table.table_name);
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
                              const viewName = (view as ViewInfo).view_name || (view as TableInfo).table_name || '';
                              return (
                                <button
                                  key={viewName}
                                  onClick={() => {
                                    toggleItem(viewName);
                                    setShowSearchModal(false);
                                  }}
                                  className="w-full text-left px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                                >
                                  👁️ {viewName}
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
                                  toggleItem(func.function_name);
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
                  )}
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

          {/* Content */}
          <div className="flex-1 overflow-y-auto">
            {/* Tables Section */}
            {filteredTables.length > 0 && (
              <SchemaSection
                title="Tables"
                icon={<TableCellsIcon className="h-4 w-4" />}
                count={filteredTables.length}
                isExpanded={expandedSections.has('tables')}
                onToggle={() => toggleSection('tables')}
              >
                {filteredTables.map((table) => (
                  <TableItem
                    key={table.table_name}
                    table={table}
                    isExpanded={expandedItems.has(table.table_name)}
                    onToggle={() => toggleItem(table.table_name)}
                    onSelectTable={onTableSelect}
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
                {filteredViews.map((view) => {
                  const viewName = (view as ViewInfo).view_name || (view as TableInfo).table_name || '';
                  return (
                    <ViewItem
                      key={viewName}
                      view={view}
                      isExpanded={expandedItems.has(viewName)}
                      onToggle={() => toggleItem(viewName)}
                      onSelectTable={onTableSelect}
                    />
                  );
                })}
              </SchemaSection>
            )}

            {/* Functions Section */}
            {filteredFunctions.length > 0 && (
              <SchemaSection
                title="Functions"
                icon={<CodeBracketIcon className="h-4 w-4" />}
                count={filteredFunctions.length}
                isExpanded={expandedSections.has('functions')}
                onToggle={() => toggleSection('functions')}
              >
                {filteredFunctions.map((func) => (
                  <FunctionItem
                    key={func.function_name}
                    function={func}
                    isExpanded={expandedItems.has(func.function_name)}
                    onToggle={() => toggleItem(func.function_name)}
                  />
                ))}
              </SchemaSection>
            )}

            {/* No Results */}
            {filteredTables.length === 0 && filteredViews.length === 0 && filteredFunctions.length === 0 && searchTerm && (
              <div className="p-4 text-center text-sm text-gray-500 dark:text-gray-400">
                No results found for &quot;{searchTerm}&quot;
              </div>
            )}
          </div>
        </>
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
    <div className="border-b border-gray-200 dark:border-gray-700">
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

// Table Item Component (existing)
interface TableItemProps {
  table: TableInfo;
  isExpanded: boolean;
  onToggle: () => void;
  onSelectTable?: (tableName: string) => void;
}

function TableItem({ table, isExpanded, onToggle, onSelectTable }: TableItemProps) {
  const primaryKeyCount = table.columns.filter((c) => c.is_primary_key).length;

  const handleClick = () => {
    onToggle();
    if (onSelectTable) {
      onSelectTable(table.table_name);
    }
  };

  return (
    <div className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <button
        onClick={handleClick}
        className="w-full px-6 py-2 flex items-center justify-between text-left"
        title="Click to view table data (LIMIT 100)"
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {isExpanded ? (
            <ChevronDownIcon className="h-4 w-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
          ) : (
            <ChevronRightIcon className="h-4 w-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
          )}
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">
            {table.table_name}
          </span>
          {primaryKeyCount > 0 && (
            <span
              className="px-1.5 py-0.5 text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-300 rounded"
              title={`${primaryKeyCount} primary key${primaryKeyCount > 1 ? 's' : ''}`}
            >
              PK
            </span>
          )}
        </div>
        <span className="text-xs text-gray-500 dark:text-gray-400">
          {table.columns.length} cols
        </span>
      </button>

      {isExpanded && (
        <div className="px-6 pb-2 pl-14">
          <div className="space-y-1">
            {table.columns.map((column) => (
              <ColumnItem key={column.column_name} column={column} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// View Item Component
interface ViewItemProps {
  view: ViewInfo | TableInfo;
  isExpanded: boolean;
  onToggle: () => void;
  onSelectTable?: (tableName: string) => void;
}

function ViewItem({ view, isExpanded, onToggle, onSelectTable }: ViewItemProps) {
  const viewName = (view as ViewInfo).view_name || (view as TableInfo).table_name;
  const columns = (view as ViewInfo).columns || (view as TableInfo).columns;

  const handleClick = () => {
    onToggle();
    if (onSelectTable) {
      onSelectTable(viewName);
    }
  };

  return (
    <div className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <button
        onClick={handleClick}
        className="w-full px-6 py-2 flex items-center justify-between text-left"
        title="Click to view view data (LIMIT 100)"
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {isExpanded ? (
            <ChevronDownIcon className="h-4 w-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
          ) : (
            <ChevronRightIcon className="h-4 w-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
          )}
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">
            {viewName}
          </span>
          <span
            className="px-1.5 py-0.5 text-xs bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded"
            title="View"
          >
            VIEW
          </span>
        </div>
        <span className="text-xs text-gray-500 dark:text-gray-400">
          {columns.length} cols
        </span>
      </button>

      {isExpanded && columns && (
        <div className="px-6 pb-2 pl-14">
          <div className="space-y-1">
            {columns.map((column) => (
              <ColumnItem key={column.column_name} column={column} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Function Item Component
interface FunctionItemProps {
  function: FunctionInfo;
  isExpanded: boolean;
  onToggle: () => void;
}

function FunctionItem({ function: func, isExpanded, onToggle }: FunctionItemProps) {
  return (
    <div className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <button
        onClick={onToggle}
        className="w-full px-6 py-2 flex items-center justify-between text-left"
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {isExpanded ? (
            <ChevronDownIcon className="h-4 w-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
          ) : (
            <ChevronRightIcon className="h-4 w-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
          )}
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate font-mono">
            {func.function_name}
          </span>
          <span
            className={`px-1.5 py-0.5 text-xs rounded ${
              func.function_type === 'aggregate'
                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300'
                : func.function_type === 'window'
                ? 'bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300'
                : 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
            }`}
          >
            {func.function_type?.toUpperCase() || 'FUNCTION'}
          </span>
        </div>
      </button>

      {isExpanded && (
        <div className="px-6 pb-2 pl-14">
          <div className="text-xs text-gray-600 dark:text-gray-400 space-y-1">
            {func.return_type && (
              <div>
                <span className="font-medium">Returns:</span> {func.return_type}
              </div>
            )}
            {func.parameters && (
              <div>
                <span className="font-medium">Parameters:</span> {func.parameters}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// Column Item Component (existing)
interface ColumnItemProps {
  column: SchemaColumnInfo;
}

function ColumnItem({ column }: ColumnItemProps) {
  return (
    <div className="flex items-center gap-2 px-2 py-1 text-sm rounded hover:bg-gray-100 dark:hover:bg-gray-700">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-gray-700 dark:text-gray-200 truncate">
            {column.column_name}
          </span>
          <div className="flex items-center gap-1">
            {column.is_primary_key && (
              <span className="px-1 py-0.5 text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-300 rounded">
                PK
              </span>
            )}
            {column.is_foreign_key && (
              <span className="px-1 py-0.5 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded">
                FK
              </span>
            )}
            {column.is_nullable && (
              <span className="px-1 py-0.5 text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded">
                NULL
              </span>
            )}
          </div>
        </div>
        <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          {column.data_type}
          {column.column_default && (
            <span className="ml-2 text-gray-400 dark:text-gray-500">
              = {column.column_default}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
