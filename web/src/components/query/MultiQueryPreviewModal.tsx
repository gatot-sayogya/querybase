'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDownIcon, ChevronUpIcon, ExclamationTriangleIcon, CheckCircleIcon, CircleStackIcon } from '@heroicons/react/24/outline';
import Button from '@/components/ui/Button';
import Modal from '@/components/ui/Modal';
import { cn } from '@/lib/utils';
import type { StatementPreview } from '@/lib/api/multi-query';

interface MultiQueryPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  statements: StatementPreview[];
  totalEstimatedRows: number;
  onApprove: () => void;
  onReject: () => void;
  loading?: boolean;
}

const operationTypeColors: Record<string, string> = {
  SELECT: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  INSERT: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  UPDATE: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  DELETE: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  CREATE_TABLE: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300',
  DROP_TABLE: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
  ALTER_TABLE: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
};

export function MultiQueryPreviewModal({
  isOpen,
  onClose,
  statements,
  totalEstimatedRows,
  onApprove,
  onReject,
  loading = false
}: MultiQueryPreviewModalProps) {
  const [expandedStatements, setExpandedStatements] = useState<Set<number>>(new Set());

  const toggleExpand = (sequence: number) => {
    const newExpanded = new Set(expandedStatements);
    if (newExpanded.has(sequence)) {
      newExpanded.delete(sequence);
    } else {
      newExpanded.add(sequence);
    }
    setExpandedStatements(newExpanded);
  };

  const WRITE_OPERATION_TYPES = ['INSERT', 'UPDATE', 'DELETE', 'CREATE_TABLE', 'DROP_TABLE', 'ALTER_TABLE'];
  const writeOperations = statements.filter(
    s => WRITE_OPERATION_TYPES.includes(s.operation_type.toUpperCase())
  );
  const requiresApproval = writeOperations.length > 0;

  const hasErrors = statements.some(s => s.error);

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Multi-Query Preview"
      size="full"
      closeOnBackdrop={!loading}
    >
      <div className="flex flex-col max-h-[80vh]">
        {/* Summary Section */}
        <div className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4 mb-4">
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold">{statements.length}</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Statements</div>
            </div>
            <div className="text-center">
              <div className={cn(
                "text-2xl font-bold",
                totalEstimatedRows === 0 && requiresApproval && "text-red-600 dark:text-red-400"
              )}>
                {totalEstimatedRows.toLocaleString()}
              </div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Est. Affected Rows</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{writeOperations.length}</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Write Operations</div>
            </div>
          </div>

          {hasErrors && (
            <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md flex items-center gap-2">
              <ExclamationTriangleIcon className="w-5 h-5 text-red-600 dark:text-red-400" />
              <span className="text-sm text-red-800 dark:text-red-300">
                Some statements have errors. Please review before proceeding.
              </span>
            </div>
          )}

          {totalEstimatedRows === 0 && requiresApproval && (
            <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md flex items-center gap-2">
              <ExclamationTriangleIcon className="w-5 h-5 text-red-600 dark:text-red-400" />
              <div className="text-sm text-red-800 dark:text-red-300">
                <div className="font-medium">No rows would be affected</div>
                <div className="text-red-700 dark:text-red-400">
                  The WHERE clause does not match any existing rows. Please review your query conditions.
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Statements List */}
        <div className="flex-1 overflow-y-auto min-h-[300px] pr-2 space-y-3">
          {statements.map((stmt) => {
            const isExpanded = expandedStatements.has(stmt.sequence);
            const hasError = !!stmt.error;

            return (
              <motion.div
                key={stmt.sequence}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: stmt.sequence * 0.05 }}
                className={cn(
                  'border rounded-lg overflow-hidden',
                  hasError 
                    ? 'border-red-200 dark:border-red-800 bg-red-50/50 dark:bg-red-900/10' 
                    : 'border-slate-200 dark:border-slate-700'
                )}
              >
                {/* Header */}
                <button
                  onClick={() => toggleExpand(stmt.sequence)}
                  className="w-full p-4 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-mono text-slate-500 dark:text-slate-400">
                      #{stmt.sequence + 1}
                    </span>
                    <span 
                      className={cn(
                        'px-2 py-1 rounded text-xs font-medium',
                        operationTypeColors[stmt.operation_type.toUpperCase()] || 'bg-gray-100'
                      )}
                    >
                      {stmt.operation_type.toUpperCase()}
                    </span>
                    <span className="text-sm text-slate-500 dark:text-slate-400 truncate max-w-md">
                      {stmt.query_text.substring(0, 80)}
                      {stmt.query_text.length > 80 ? '...' : ''}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    {stmt.estimated_rows > 0 && (
                      <span className="text-sm text-slate-500 dark:text-slate-400">
                        ~{stmt.estimated_rows.toLocaleString()} rows
                      </span>
                    )}
                    {hasError && (
                      <ExclamationTriangleIcon className="w-4 h-4 text-red-600 dark:text-red-400" />
                    )}
                    {isExpanded ? (
                      <ChevronUpIcon className="w-4 h-4 text-slate-400" />
                    ) : (
                      <ChevronDownIcon className="w-4 h-4 text-slate-400" />
                    )}
                  </div>
                </button>

                {/* Expanded Content */}
                <AnimatePresence>
                  {isExpanded && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="border-t border-slate-200 dark:border-slate-700"
                    >
                      <div className="p-4 space-y-4">
                        {/* SQL Preview */}
                        <div>
                          <div className="text-sm font-medium mb-2">SQL</div>
                          <pre className="bg-slate-100 dark:bg-slate-800 p-3 rounded-md text-sm font-mono overflow-x-auto">
                            {stmt.query_text}
                          </pre>
                        </div>

                        {/* Error Message */}
                        {hasError && (
                          <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
                            <div className="text-sm font-medium text-red-800 dark:text-red-300 mb-1">
                              Error
                            </div>
                            <div className="text-sm text-red-700 dark:text-red-400">
                              {stmt.error}
                            </div>
                          </div>
                        )}

                        {/* Preview Data */}
                        {stmt.preview_rows && stmt.preview_rows.length > 0 && (
                          <div>
                            <div className="text-sm font-medium mb-2">
                              Preview ({stmt.preview_rows.length} rows)
                            </div>
                            <div className="overflow-x-auto">
                              <table className="w-full text-sm">
                                <thead>
                                  <tr className="border-b border-slate-200 dark:border-slate-700">
                                    {stmt.columns?.map((col) => (
                                      <th key={col.name} className="text-left p-2 font-medium">
                                        {col.name}
                                      </th>
                                    ))}
                                  </tr>
                                </thead>
                                <tbody>
                                  {stmt.preview_rows.slice(0, 5).map((row, idx) => (
                                    <tr key={idx} className="border-b border-slate-100 dark:border-slate-800 last:border-0">
                                      {stmt.columns?.map((col) => (
                                        <td key={col.name} className="p-2 text-slate-600 dark:text-slate-400">
                                          {row[col.name] === null ? (
                                            <span className="italic">null</span>
                                          ) : (
                                            String(row[col.name]).substring(0, 50)
                                          )}
                                        </td>
                                      ))}
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                              {stmt.preview_rows.length > 5 && (
                                <div className="text-center text-sm text-slate-500 dark:text-slate-400 py-2">
                                  +{stmt.preview_rows.length - 5} more rows
                                </div>
                              )}
                            </div>
                          </div>
                        )}

                        {/* Stats */}
                        <div className="flex gap-4 text-sm text-slate-500 dark:text-slate-400">
                          <span>Estimated rows: {stmt.estimated_rows.toLocaleString()}</span>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            );
          })}
        </div>

        {/* Footer */}
        <div className="flex justify-between items-center gap-4 pt-4 border-t border-slate-200 dark:border-slate-700 mt-4">
          <div className="text-sm text-slate-500 dark:text-slate-400">
            {totalEstimatedRows === 0 && requiresApproval ? (
              <span className="text-red-600 dark:text-red-400 font-medium">
                ⚠ No rows would be affected — execution is disabled
              </span>
            ) : requiresApproval ? (
              <span className="text-amber-600 dark:text-amber-400 font-medium">
                ⚠ Contains write operations — requires approval before execution
              </span>
            ) : (
              'All statements will execute atomically'
            )}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onReject} disabled={loading}>
              Cancel
            </Button>
            <Button
              onClick={onApprove}
              disabled={loading || hasErrors || (totalEstimatedRows === 0 && requiresApproval)}
            >
              {loading ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                  Processing...
                </>
              ) : totalEstimatedRows === 0 && requiresApproval ? (
                <>
                  <ExclamationTriangleIcon className="w-4 h-4 mr-2" />
                  No Rows Affected
                </>
              ) : requiresApproval ? (
                <>
                  <CheckCircleIcon className="w-4 h-4 mr-2" />
                  Submit for Approval
                </>
              ) : (
                <>
                  <CheckCircleIcon className="w-4 h-4 mr-2" />
                  Execute All
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  );
}
