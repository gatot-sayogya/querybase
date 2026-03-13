'use client';

import { useState } from 'react';
import type { AuditMode } from '@/types';
import { formatDate } from '@/lib/utils';

interface DataChangesPanelProps {
  affectedRows: number;
  auditMode: AuditMode;
  beforeData?: Record<string, unknown>[];
  afterData?: Record<string, unknown>[];
  completedAt?: string;
  reviewerName?: string;
}

export default function DataChangesPanel({
  affectedRows,
  auditMode,
  beforeData,
  afterData,
  completedAt,
  reviewerName
}: DataChangesPanelProps) {
  const [activeTab, setActiveTab] = useState<'after' | 'before'>('after');
  const hasData = (beforeData && beforeData.length > 0) || (afterData && afterData.length > 0);
  const [isExpanded, setIsExpanded] = useState(hasData);

  // Use either before/after columns or empty if none exist
  const getColumns = (data?: Record<string, unknown>[]) => {
    if (!data || data.length === 0) return [];
    return Object.keys(data[0]);
  };

  const currentData = activeTab === 'after' ? afterData : beforeData;
  const columns = getColumns(currentData);
  const dataExists = (beforeData && beforeData.length > 0) || (afterData && afterData.length > 0);

  return (
    <div className="mt-6 border border-gray-200 dark:border-gray-700 overflow-hidden flex flex-col">
      <div 
        className="bg-gray-50 dark:bg-gray-800/50 p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center bg-green-100 dark:bg-green-900/30">
              <svg className="h-4 w-4 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </span>
            Execution Result
          </h3>
          <div className="mt-1 flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
            <span>
              <strong className="text-gray-900 dark:text-gray-200 font-mono">{affectedRows}</strong> rows affected
            </span>
            <span>&bull;</span>
            <span className="flex items-center gap-1">
              Audit mode: <span className="px-1.5 py-0.5 text-xs font-mono bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300">{auditMode}</span>
            </span>
            {completedAt && (
              <>
                <span>&bull;</span>
                <span>Committed {formatDate(completedAt)}</span>
              </>
            )}
            {reviewerName && (
              <>
                <span>&bull;</span>
                <span>by {reviewerName}</span>
              </>
            )}
          </div>
        </div>
        <button className="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-200 dark:hover:text-gray-300 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500">
          <svg
            className={`h-5 w-5 transform transition-transform ${isExpanded ? 'rotate-180' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      {isExpanded && (
        <div className="p-0 bg-white dark:bg-gray-900">
          {auditMode === 'count_only' || !dataExists ? (
            <div className="p-8 text-center bg-gray-50/50 dark:bg-gray-800/20">
              <svg className="mx-auto h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
              </svg>
              <h4 className="mt-2 text-sm font-medium text-gray-900 dark:text-gray-300">No row data captured</h4>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                This transaction was committed in <span className="font-mono text-xs">count_only</span> mode, or the query did not return any row modifications.
              </p>
            </div>
          ) : (
            <div className="flex flex-col h-full border-t border-gray-200 dark:border-gray-800">
              <div className="flex bg-gray-50 dark:bg-gray-800/50 px-4 pt-2 border-b border-gray-200 dark:border-gray-800">
                <button
                  onClick={() => setActiveTab('after')}
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors duration-200 ${
                    activeTab === 'after'
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                  }`}
                >
                  After Operation Data
                </button>
                <button
                  onClick={() => setActiveTab('before')}
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors duration-200 ${
                    activeTab === 'before'
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                  }`}
                >
                  Before Operation Data
                </button>
              </div>
              
              <div className="overflow-auto max-h-[400px]">
                {!currentData || currentData.length === 0 ? (
                  <div className="p-8 text-center text-sm text-gray-500 dark:text-gray-400 italic">
                    No {activeTab} data available.
                  </div>
                ) : (
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th className="w-12 text-center text-gray-400">#</th>
                        {columns.map((col, idx) => (
                          <th key={idx}>{col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {currentData.map((row, rowIdx) => (
                        <tr key={rowIdx}>
                          <td className="text-center text-gray-400 font-mono text-xs" style={{ background: 'var(--bg-hover)' }}>{rowIdx + 1}</td>
                          {columns.map((col, colIdx) => {
                            const val = row[col];
                            return (
                              <td key={colIdx}>
                                {val === null ? (
                                  <span className="text-gray-400 italic font-mono text-xs">null</span>
                                ) : typeof val === 'object' ? (
                                  JSON.stringify(val)
                                ) : (
                                  String(val)
                                )}
                              </td>
                            )
                          })}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
