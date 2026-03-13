'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDownIcon, ChevronUpIcon, CheckCircleIcon, XCircleIcon, ClockIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline';
import Card from '@/components/ui/Card';
import { cn } from '@/lib/utils';
import type { StatementResult } from '@/lib/api/multi-query';

interface MultiQueryResultsProps {
  transactionId?: string;
  statements: StatementResult[];
  totalExecutionTime: number;
  totalAffectedRows: number;
  status: 'pending' | 'success' | 'failed';
  errorMessage?: string;
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

const statusIcons = {
  pending: ClockIcon,
  success: CheckCircleIcon,
  failed: XCircleIcon
};

const statusColors = {
  pending: 'text-yellow-600 dark:text-yellow-400',
  success: 'text-green-600 dark:text-green-400',
  failed: 'text-red-600 dark:text-red-400'
};

export function MultiQueryResults({
  transactionId,
  statements,
  totalExecutionTime,
  totalAffectedRows,
  status,
  errorMessage
}: MultiQueryResultsProps) {
  const [expandedStatements, setExpandedStatements] = useState<Set<number>>(
    new Set(statements.filter(s => s.status === 'failed').map(s => s.sequence))
  );

  const toggleExpand = (sequence: number) => {
    const newExpanded = new Set(expandedStatements);
    if (newExpanded.has(sequence)) {
      newExpanded.delete(sequence);
    } else {
      newExpanded.add(sequence);
    }
    setExpandedStatements(newExpanded);
  };

  const StatusIcon = statusIcons[status];

  return (
    <div className="space-y-6">
      {/* Summary Card */}
      <Card>
        <div className="p-6">
          <div className="flex items-center gap-2 text-lg font-bold mb-4">
            <StatusIcon className={cn('w-5 h-5', statusColors[status])} />
            Multi-Query Results
            {transactionId && (
              <span className="text-sm font-normal text-slate-500 dark:text-slate-400">
                (Transaction: {transactionId.slice(0, 8)}...)
              </span>
            )}
          </div>
          
          <div className="grid grid-cols-4 gap-4">
            <div className="text-center p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
              <div className={cn('text-2xl font-bold', statusColors[status])}>
                {status === 'success' ? 'Success' : status === 'failed' ? 'Failed' : 'Pending'}
              </div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Status</div>
            </div>
            <div className="text-center p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
              <div className="text-2xl font-bold">{statements.length}</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Statements</div>
            </div>
            <div className="text-center p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
              <div className="text-2xl font-bold">{totalAffectedRows.toLocaleString()}</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Total Affected</div>
            </div>
            <div className="text-center p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
              <div className="text-2xl font-bold">{(totalExecutionTime / 1000).toFixed(2)}s</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Execution Time</div>
            </div>
          </div>

          {errorMessage && (
            <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md flex items-start gap-2">
              <ExclamationCircleIcon className="w-5 h-5 text-red-600 dark:text-red-400 mt-0.5" />
              <div>
                <div className="text-sm font-medium text-red-800 dark:text-red-300">
                  Execution Failed
                </div>
                <div className="text-sm text-red-700 dark:text-red-400">
                  {errorMessage}
                </div>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Statement Results */}
      <div className="space-y-3">
        <h3 className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
          Statement Details
        </h3>
        
        <div className="max-h-[500px] overflow-y-auto pr-2 space-y-3">
          {statements.map((stmt) => {
            const isExpanded = expandedStatements.has(stmt.sequence);
            const StatementIcon = statusIcons[stmt.status as keyof typeof statusIcons] || ClockIcon;

            return (
              <motion.div
                key={stmt.sequence}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: stmt.sequence * 0.05 }}
                className={cn(
                  'border rounded-lg overflow-hidden',
                  stmt.status === 'failed' 
                    ? 'border-red-200 dark:border-red-800' 
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
                        operationTypeColors[stmt.operation_type] || 'bg-gray-100'
                      )}
                    >
                      {stmt.operation_type}
                    </span>
                    <span className="text-sm text-slate-500 dark:text-slate-400 truncate max-w-xs">
                      {stmt.query_text.substring(0, 60)}
                      {stmt.query_text.length > 60 ? '...' : ''}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <StatementIcon 
                      className={cn(
                        'w-4 h-4',
                        statusColors[stmt.status as keyof typeof statusColors] || 'text-slate-400'
                      )} 
                    />
                    {stmt.affected_rows > 0 && (
                      <span className="text-sm text-slate-500 dark:text-slate-400">
                        {stmt.affected_rows.toLocaleString()} rows
                      </span>
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
                        {/* SQL */}
                        <div>
                          <div className="text-sm font-medium mb-2">SQL</div>
                          <pre className="bg-slate-100 dark:bg-slate-800 p-3 rounded-md text-sm font-mono overflow-x-auto">
                            {stmt.query_text}
                          </pre>
                        </div>

                        {/* Error */}
                        {stmt.error_message && (
                          <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
                            <div className="text-sm font-medium text-red-800 dark:text-red-300 mb-1">
                              Error
                            </div>
                            <div className="text-sm text-red-700 dark:text-red-400">
                              {stmt.error_message}
                            </div>
                          </div>
                        )}

                        {/* Result Data (for SELECT) */}
                        {stmt.data && stmt.data.length > 0 && (
                          <div>
                            <div className="text-sm font-medium mb-2">
                              Results ({stmt.row_count} rows)
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
                                  {stmt.data.slice(0, 10).map((row, idx) => (
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
                              {stmt.data.length > 10 && (
                                <div className="text-center text-sm text-slate-500 dark:text-slate-400 py-2">
                                  +{stmt.data.length - 10} more rows
                                </div>
                              )}
                            </div>
                          </div>
                        )}

                        {/* Stats */}
                        <div className="flex flex-wrap gap-4 text-sm text-slate-500 dark:text-slate-400">
                          <span>Status: <span className={cn('font-medium', statusColors[stmt.status as keyof typeof statusColors])}>{stmt.status}</span></span>
                          <span>Affected rows: {stmt.affected_rows.toLocaleString()}</span>
                          <span>Execution time: {stmt.execution_time_ms}ms</span>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
