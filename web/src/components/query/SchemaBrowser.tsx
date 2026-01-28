'use client';

import { useEffect, useState } from 'react';
import { ChevronDownIcon, ChevronRightIcon } from '@heroicons/react/24/outline';
import { useSchemaStore } from '@/stores/schema-store';
import { useAuthStore } from '@/stores/auth-store';
import type { TableInfo, SchemaColumnInfo } from '@/types';
import DataSourceSelector from './DataSourceSelector';

export default function SchemaBrowser() {
  const { isAuthenticated } = useAuthStore();
  const [selectedDataSource, setSelectedDataSource] = useState<string | null>(null);
  const [expandedTables, setExpandedTables] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState('');

  const {
    schemas,
    loadSchema,
    isLoading,
    error,
    setCurrentDataSource,
    getTableNames,
    searchTables,
  } = useSchemaStore();

  useEffect(() => {
    if (selectedDataSource && isAuthenticated) {
      setCurrentDataSource(selectedDataSource);
      loadSchema(selectedDataSource);
    }
  }, [selectedDataSource, isAuthenticated]);

  const currentSchema = selectedDataSource ? schemas.get(selectedDataSource) : null;
  const filteredTables = currentSchema?.tables.filter((table) =>
    table.table_name.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  const toggleTable = (tableName: string) => {
    setExpandedTables((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(tableName)) {
        newSet.delete(tableName);
      } else {
        newSet.add(tableName);
      }
      return newSet;
    });
  };

  const expandAll = () => {
    if (currentSchema) {
      setExpandedTables(new Set(currentSchema.tables.map((t) => t.table_name)));
    }
  };

  const collapseAll = () => {
    setExpandedTables(new Set());
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-3">Database Schema</h3>
        <DataSourceSelector
          value={selectedDataSource}
          onChange={setSelectedDataSource}
          className="w-full"
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
            <div className="flex items-center justify-between mb-2">
              <div>
                <p className="text-xs font-medium text-gray-600 dark:text-gray-300">
                  {currentSchema.data_source_name}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  {currentSchema.database_type} • {currentSchema.tables.length} tables
                </p>
              </div>
              <div className="flex gap-1">
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

            {/* Search */}
            <input
              type="text"
              placeholder="Search tables..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Tables List */}
          <div className="flex-1 overflow-y-auto">
            {filteredTables.length === 0 ? (
              <div className="p-4 text-center text-sm text-gray-500 dark:text-gray-400">
                {searchTerm ? 'No tables found' : 'No tables available'}
              </div>
            ) : (
              <div className="divide-y divide-gray-200 dark:divide-gray-700">
                {filteredTables.map((table) => (
                  <TableItem
                    key={table.table_name}
                    table={table}
                    isExpanded={expandedTables.has(table.table_name)}
                    onToggle={() => toggleTable(table.table_name)}
                  />
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}

interface TableItemProps {
  table: TableInfo;
  isExpanded: boolean;
  onToggle: () => void;
}

function TableItem({ table, isExpanded, onToggle }: TableItemProps) {
  const primaryKeyCount = table.columns.filter((c) => c.is_primary_key).length;

  return (
    <div className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <button
        onClick={onToggle}
        className="w-full px-4 py-2 flex items-center justify-between text-left"
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
        <div className="px-4 pb-2 pl-10">
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
