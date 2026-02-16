'use client';

import { useState } from 'react';
import type { QueryResult, ColumnInfo } from '@/types';
import { formatDate } from '@/lib/utils';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { QueryError } from '@/components/ui/Alert';

interface QueryResultsProps {
  queryId: string;
  results: QueryResult | null;
  loading: boolean;
  error: string | null;
}

export default function QueryResults({
  queryId,
  results,
  loading,
  error,
}: QueryResultsProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const [rowsPerPage] = useState(50);

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
  const data = results.data;

  // Client-side pagination
  const totalPages = Math.ceil(data.length / rowsPerPage);
  const startIndex = (currentPage - 1) * rowsPerPage;
  const endIndex = startIndex + rowsPerPage;
  const paginatedData = data.slice(startIndex, endIndex);

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
    <div>
      {/* Results Header */}
      <div className="flex items-center justify-between px-2 py-1 bg-gray-50 dark:bg-gray-800/30 text-xs border-b border-gray-200 dark:border-gray-700">
        <div className="text-gray-600 dark:text-gray-400">
          {startIndex + 1}-{Math.min(endIndex, data.length)} of {data.length}
        </div>
        {data.length > rowsPerPage && (
          <div className="flex items-center gap-1">
            <button
              onClick={handlePrevious}
              disabled={currentPage === 1}
              className="px-2 py-0.5 text-xs border border-gray-300 dark:border-gray-700 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed dark:text-white"
            >
              ←
            </button>
            <span className="text-gray-600 dark:text-gray-400 px-1">
              {currentPage}/{totalPages}
            </span>
            <button
              onClick={handleNext}
              disabled={currentPage === totalPages}
              className="px-2 py-0.5 text-xs border border-gray-300 dark:border-gray-700 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed dark:text-white"
            >
              →
            </button>
          </div>
        )}
      </div>

      {/* Results Table - Now with flexible max height */}
      <div className="overflow-auto border-b border-gray-200 dark:border-gray-700 max-h-[650px]">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.name}
                  className="px-2 py-1 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-tight"
                >
                  <div className="flex flex-col">
                    <span className="font-semibold">{column.name}</span>
                    <span className="text-[9px] text-gray-400 dark:text-gray-500 font-normal">
                      {column.type}
                    </span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
            {paginatedData.map((row, rowIndex) => (
              <tr key={rowIndex} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                {columns.map((column) => (
                  <td
                    key={column.name}
                    className="px-2 py-1 text-xs text-gray-900 dark:text-gray-100 max-w-xs truncate"
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
      {data.length > rowsPerPage && (
        <div className="flex items-center justify between px-2 py-1 bg-gray-50 dark:bg-gray-800/30 text-xs border-t border-gray-200 dark:border-gray-700">
          <div className="text-gray-600 dark:text-gray-400">
            {data.length} total
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={handlePrevious}
              disabled={currentPage === 1}
              className="px-2 py-0.5 text-xs border border-gray-300 dark:border-gray-700 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed dark:text-white"
            >
              ←
            </button>
            <span className="text-gray-600 dark:text-gray-400 px-1">
              {currentPage}/{totalPages}
            </span>
            <button
              onClick={handleNext}
              disabled={currentPage === totalPages}
              className="px-2 py-0.5 text-xs border border-gray-300 dark:border-gray-700 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed dark:text-white"
            >
              →
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
