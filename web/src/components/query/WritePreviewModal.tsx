'use client';

import { useState } from 'react';
import type { WriteQueryPreview } from '@/types';

interface WritePreviewModalProps {
  preview: WriteQueryPreview;
  queryText: string;
  onConfirm: () => void;
  onCancel: () => void;
  loading?: boolean;
}

export default function WritePreviewModal({
  preview,
  queryText,
  onConfirm,
  onCancel,
  loading = false,
}: WritePreviewModalProps) {
  const [showQuery, setShowQuery] = useState(false);
  const isDelete = preview.operation_type?.toUpperCase() === 'DELETE';
  const isTruncated = (preview.preview_rows?.length || 0) < preview.total_affected;

  const opColor = isDelete
    ? 'text-red-600 dark:text-red-400'
    : 'text-yellow-600 dark:text-yellow-400';
  const opBg = isDelete
    ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'
    : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
  const confirmBg = isDelete
    ? 'bg-red-600 hover:bg-red-700 focus:ring-red-500'
    : 'bg-yellow-600 hover:bg-yellow-700 focus:ring-yellow-500';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 max-w-4xl w-full mx-4 max-h-[85vh] flex flex-col animate-slide-up">
        {/* Header */}
        <div className={`p-5 border-b ${opBg} rounded-t-xl`}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className={`flex h-10 w-10 items-center justify-center rounded-full ${isDelete ? 'bg-red-100 dark:bg-red-900/40' : 'bg-yellow-100 dark:bg-yellow-900/40'}`}>
                {isDelete ? (
                  <svg className="h-5 w-5 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                ) : (
                  <svg className="h-5 w-5 text-yellow-600 dark:text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                )}
              </span>
              <div>
                <h2 className={`text-lg font-semibold ${opColor}`}>
                  {isDelete ? 'Delete' : 'Update'} Preview
                </h2>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Review affected rows before submitting for approval
                </p>
              </div>
            </div>
            <button
              onClick={onCancel}
              className="p-2 rounded-lg text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* Impact Summary */}
        <div className="p-5 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-6">
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">Total Rows Affected</span>
              <div className={`text-3xl font-bold font-mono ${opColor}`} style={{ fontVariantNumeric: 'tabular-nums' }}>
                {preview.total_affected.toLocaleString()}
              </div>
            </div>
            <div className="h-12 w-px bg-gray-200 dark:bg-gray-700" />
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">Showing Preview</span>
              <div className="text-lg font-medium text-gray-900 dark:text-gray-100">
                {preview.preview_rows?.length || 0} of {preview.total_affected.toLocaleString()} rows
              </div>
            </div>
            {preview.total_affected > 100 && (
              <>
                <div className="h-12 w-px bg-gray-200 dark:bg-gray-700" />
                <div className="flex items-center gap-2 p-2 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg">
                  <svg className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                  <span className="text-sm text-amber-800 dark:text-amber-300">
                    Large impact — only first {preview.preview_limit} rows shown
                  </span>
                </div>
              </>
            )}
          </div>

          {/* Toggle original query */}
          <button
            onClick={() => setShowQuery(!showQuery)}
            className="mt-3 text-xs text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1"
          >
            <svg className={`h-3 w-3 transition-transform ${showQuery ? 'rotate-90' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
            {showQuery ? 'Hide' : 'Show'} query
          </button>
          {showQuery && (
            <pre className="mt-2 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg text-sm font-mono text-gray-800 dark:text-gray-200 overflow-x-auto border border-gray-200 dark:border-gray-700">
              {queryText}
            </pre>
          )}
        </div>

        {/* Data Table */}
        <div className="flex-1 overflow-auto min-h-0">
          {!preview.preview_rows || preview.preview_rows.length === 0 ? (
            <div className="p-8 text-center">
              <svg className="mx-auto h-10 w-10 mb-3 text-blue-400 dark:text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1">No matching rows found</p>
              <p className="text-xs text-gray-500 dark:text-gray-400">This query would affect 0 rows. Check your WHERE clause and try again.</p>
            </div>
          ) : (
            <table className="data-table w-full">
              <thead className="sticky top-0 z-10">
                <tr>
                  <th className="w-12 px-4 py-2 text-center text-gray-400 bg-gray-50 dark:bg-gray-800 whitespace-nowrap">#</th>
                  {(preview.columns || []).map((col, idx) => (
                    <th key={idx} className="bg-gray-50 dark:bg-gray-800 whitespace-nowrap px-4 py-2 text-left font-medium">{col}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(preview.preview_rows || []).map((row, rowIdx) => (
                  <tr key={rowIdx} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                    <td className="text-center text-gray-400 font-mono text-xs whitespace-nowrap px-4 py-2" style={{ background: 'var(--bg-hover)' }}>{rowIdx + 1}</td>
                    {(preview.columns || []).map((col, colIdx) => {
                      const val = row[col];
                      return (
                        <td key={colIdx} className="whitespace-nowrap px-4 py-2">
                          {val === null ? (
                            <span className="text-gray-400 italic font-mono text-xs">null</span>
                          ) : typeof val === 'object' ? (
                            JSON.stringify(val)
                          ) : (
                            String(val)
                          )}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          {isTruncated && (
            <div className="p-3 text-center text-sm text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
              Showing {preview.preview_rows?.length || 0} of {preview.total_affected.toLocaleString()} total affected rows
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="p-5 border-t border-gray-200 dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 rounded-b-xl">
          <button
            onClick={onCancel}
            disabled={loading}
            className="px-5 py-2.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors disabled:opacity-50"
          >
            {preview.total_affected === 0 ? 'Edit Query' : 'Cancel'}
          </button>
          {preview.total_affected > 0 && (
            <button
              onClick={onConfirm}
              disabled={loading}
              className={`px-5 py-2.5 text-sm font-medium text-white rounded-lg ${confirmBg} transition-colors disabled:opacity-50 flex items-center gap-2 focus:outline-none focus:ring-2 focus:ring-offset-2`}
            >
              {loading ? (
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              )}
              Confirm & Submit for Approval
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
