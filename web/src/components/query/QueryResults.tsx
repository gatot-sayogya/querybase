'use client';

import { useState, useMemo } from 'react';
import type { QueryResult, ColumnInfo } from '@/types';
import { formatDate } from '@/lib/utils';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { QueryError } from '@/components/ui/Alert';
import { ArrowsUpDownIcon, FunnelIcon, ArrowDownIcon, ArrowUpIcon, ArrowsPointingOutIcon, ArrowsPointingInIcon } from '@heroicons/react/24/outline';

interface QueryResultsProps {
  queryId: string;
  results: QueryResult | null;
  loading: boolean;
  error: string | null;
  isFullscreen?: boolean;
  onToggleFullscreen?: () => void;
}
export default function QueryResults({
  queryId,
  results,
  loading,
  error,
  isFullscreen = false,
  onToggleFullscreen,
}: QueryResultsProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const [rowsPerPage] = useState(50);
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: 'asc' | 'desc' | null }>({
    key: '',
    direction: null,
  });
  const [filterTerms, setFilterTerms] = useState<Record<string, string>>({});
  const [isFilterVisible, setIsFilterVisible] = useState(false);

  // Sorting and Filtering logic
  const processedData = useMemo(() => {
    if (!results || results.data.length === 0) return [];
    
    let data = [...results.data];

    // Apply Filter
    Object.entries(filterTerms).forEach(([key, term]) => {
      if (term) {
        data = data.filter(row => 
          String(row[key] ?? '').toLowerCase().includes(term.toLowerCase())
        );
      }
    });

    // Apply Sort
    if (sortConfig.key && sortConfig.direction) {
      data.sort((a, b) => {
        const aVal = a[sortConfig.key];
        const bVal = b[sortConfig.key];
        
        if (aVal === bVal) return 0;
        if (aVal === undefined || aVal === null) return 1;
        if (bVal === undefined || bVal === null) return -1;
        
        const comparison = aVal < bVal ? -1 : 1;
        return sortConfig.direction === 'asc' ? comparison : -comparison;
      });
    }

    return data;
  }, [results, filterTerms, sortConfig]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="w-full max-w-4xl">
          <TableSkeleton rows={8} columns={results?.columns.length || 5} />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <QueryError
        error={error}
        onRetry={() => {
          // Parent component should handle retry by re-executing the query
          window.location.reload();
        }}
      />
    );
  }

  if (!results || results.data.length === 0) {
    return (
      <EmptyState
        illustration="no-results"
        title="No results found"
        description="Your query executed successfully but returned 0 rows. Try adjusting your query or filters."
      />
    );
  }

  const columns = results.columns;
  const rawData = results.data;

  const handleSort = (key: string) => {
    setSortConfig(current => ({
      key,
      direction: current.key === key && current.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  const handleFilterChange = (key: string, term: string) => {
    setFilterTerms(prev => ({ ...prev, [key]: term }));
    setCurrentPage(1); // Reset to first page on filter
  };

  // Client-side pagination
  const totalPages = Math.ceil(processedData.length / rowsPerPage);
  const startIndex = (currentPage - 1) * rowsPerPage;
  const endIndex = startIndex + rowsPerPage;
  const paginatedData = processedData.slice(startIndex, endIndex);

  const handlePrevious = () => {
    if (currentPage > 1) {
      setCurrentPage(currentPage - 1);
    }
  };

  const handleNext = () => {
    if (currentPage < totalPages) {
      setCurrentPage(currentPage + 1);
    }
  };

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 bg-transparent text-xs border-b border-slate-200/50 dark:border-white/10">
        <div className="flex items-center gap-4">
          <div className="text-slate-500 dark:text-slate-400 font-medium">
            Showing <strong className="text-slate-700 dark:text-slate-200">{processedData.length > 0 ? startIndex + 1 : 0}-{Math.min(endIndex, processedData.length)}</strong> of <strong className="text-slate-700 dark:text-slate-200">{processedData.length}</strong>
            {processedData.length !== rawData.length && (
              <span className="ml-2 text-blue-500 font-bold shrink-0">(Filtered)</span>
            )}
          </div>
          
          <div className="flex items-center gap-1">
            <button
              onClick={() => setIsFilterVisible(!isFilterVisible)}
              className={`p-1.5 rounded-lg transition-all ${isFilterVisible ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/30' : 'text-slate-500 hover:bg-slate-500/10'}`}
              title="Toggle column filters"
            >
              <FunnelIcon className="h-4 w-4" />
            </button>
            <button
              onClick={onToggleFullscreen}
              className="p-1.5 text-slate-500 hover:bg-slate-500/10 rounded-lg transition-all"
              title={isFullscreen ? "Exit fullscreen" : "Enter fullscreen"}
            >
              {isFullscreen ? <ArrowsPointingInIcon className="h-4 w-4" /> : <ArrowsPointingOutIcon className="h-4 w-4" />}
            </button>
          </div>
        </div>

        {processedData.length > rowsPerPage && (
          <div className="flex items-center gap-1">
            <button
              onClick={handlePrevious}
              disabled={currentPage === 1}
              className="px-2 py-1 text-xs font-bold text-slate-500 rounded-lg hover:bg-slate-500/10 dark:hover:bg-white/10 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              ← PREV
            </button>
            <span className="text-xs font-bold tracking-widest text-slate-400 uppercase px-2 text-center min-w-[60px]">
              {currentPage} / {totalPages}
            </span>
            <button
              onClick={handleNext}
              disabled={currentPage === totalPages}
              className="px-2 py-1 text-xs font-bold text-slate-500 rounded-lg hover:bg-slate-500/10 dark:hover:bg-white/10 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              NEXT →
            </button>
          </div>
        )}
      </div>

      {/* Results Table - Now with flexible height */}
      <div className="flex-1 overflow-auto min-h-0 custom-scrollbar">
        <table className="min-w-full divide-y divide-slate-200/50 dark:divide-white/5">
          <thead className="bg-white/50 dark:bg-slate-900/50 sticky top-0 z-10 backdrop-blur-md">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.name}
                  onClick={() => handleSort(column.name)}
                  className="px-4 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest whitespace-nowrap cursor-pointer hover:bg-slate-500/10 transition-colors group/th"
                >
                  <div className="flex flex-col gap-0.5">
                    <div className="flex items-center gap-2">
                      <span className="text-slate-700 dark:text-slate-300 group-hover/th:text-blue-500 transition-colors">{column.name}</span>
                      <span className="text-slate-400">
                        {sortConfig.key === column.name ? (
                          sortConfig.direction === 'asc' ? <ArrowUpIcon className="h-3 w-3" /> : <ArrowDownIcon className="h-3 w-3" />
                        ) : (
                          <ArrowsUpDownIcon className="h-3 w-3 opacity-0 group-hover/th:opacity-100 transition-opacity" />
                        )}
                      </span>
                    </div>
                    <span className="text-[9px] text-slate-400 dark:text-slate-500/80 font-medium lowercase">
                      {column.type}
                    </span>
                  </div>
                </th>
              ))}
            </tr>
            {isFilterVisible && (
              <tr className="bg-slate-50 dark:bg-slate-900/80 border-b border-slate-200 dark:border-white/5 animate-in slide-in-from-top-1 duration-200">
                {columns.map((column) => (
                  <th key={`filter-${column.name}`} className="px-2 py-2">
                    <input
                      type="text"
                      placeholder={`Filter...`}
                      value={filterTerms[column.name] || ''}
                      onChange={(e) => handleFilterChange(column.name, e.target.value)}
                      className="w-full bg-white dark:bg-slate-800 border border-slate-200 dark:border-white/10 rounded-lg px-2 py-1 text-[10px] font-bold text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-500/50 placeholder:text-slate-400/50"
                    />
                  </th>
                ))}
              </tr>
            )}
          </thead>
          <tbody className="bg-transparent divide-y divide-slate-100 dark:divide-white/5">
            {paginatedData.map((row, rowIndex) => (
              <tr key={rowIndex} className="hover:bg-blue-50/50 dark:hover:bg-blue-900/10 transition-colors group">
                {columns.map((column) => (
                  <td
                    key={column.name}
                    className="px-4 py-2 text-xs text-slate-700 dark:text-slate-300 max-w-xs truncate group-hover:text-slate-900 dark:group-hover:text-white transition-colors"
                    title={String(row[column.name])}
                  >
                    {formatCellValue(row[column.name])}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer */}
      {processedData.length > rowsPerPage && (
        <div className="flex-shrink-0 flex items-center justify-between px-4 py-2 bg-transparent text-xs border-t border-slate-200/50 dark:border-white/10">
          <div className="text-slate-500 dark:text-slate-400 font-medium">
            Total Records: <strong className="text-slate-700 dark:text-slate-200">{processedData.length}</strong>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={handlePrevious}
              disabled={currentPage === 1}
              className="px-2 py-1 text-xs font-bold text-slate-500 rounded-lg hover:bg-slate-500/10 dark:hover:bg-white/10 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              ← PREV
            </button>
            <span className="text-xs font-bold tracking-widest text-slate-400 uppercase px-2">
              {currentPage} / {totalPages}
            </span>
            <button
              onClick={handleNext}
              disabled={currentPage === totalPages}
              className="px-2 py-1 text-xs font-bold text-slate-500 rounded-lg hover:bg-slate-500/10 dark:hover:bg-white/10 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              NEXT →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function formatCellValue(value: unknown): React.ReactNode {
  if (value === null) {
    return <span className="text-gray-400 italic">NULL</span>;
  }
  if (value === undefined) {
    return '';
  }
  if (typeof value === 'boolean') {
    return value ? 'true' : 'false';
  }
  if (typeof value === 'object') {
    return JSON.stringify(value);
  }
  return String(value);
}
