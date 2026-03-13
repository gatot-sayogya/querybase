'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';

export type QueryExecutionStatus = 
  | 'idle'
  | 'running' 
  | 'completed' 
  | 'completed_empty'
  | 'failed' 
  | 'no_match' 
  | 'pending_approval';

interface QueryStatusIndicatorProps {
  status: QueryExecutionStatus;
  executionTime?: number;
  rowCount?: number;
  message?: string;
}

export default function QueryStatusIndicator({
  status,
  executionTime,
  rowCount,
  message,
}: QueryStatusIndicatorProps) {
  const getStatusConfig = () => {
    switch (status) {
      case 'running':
        return {
          icon: (
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
              className="w-5 h-5"
            >
              <svg className="w-5 h-5 text-blue-500" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            </motion.div>
          ),
          color: 'bg-blue-50 dark:bg-blue-500/10 border-blue-200 dark:border-blue-500/20',
          textColor: 'text-blue-800 dark:text-blue-300',
          title: 'Executing query...',
          subtitle: message || 'Please wait',
        };

      case 'completed':
        return {
          icon: (
            <svg className="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          ),
          color: 'bg-green-50 dark:bg-green-500/10 border-green-200 dark:border-green-500/20',
          textColor: 'text-green-800 dark:text-green-300',
          title: 'Query completed successfully',
          subtitle: executionTime !== undefined 
            ? `Duration: ${executionTime < 1000 ? `${executionTime}ms` : `${(executionTime / 1000).toFixed(2)}s`} • ${rowCount ?? 0} rows`
            : `${rowCount ?? 0} rows returned`,
        };

      case 'completed_empty':
        return {
          icon: (
            <svg className="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          ),
          color: 'bg-green-50/50 dark:bg-green-500/5 border-green-200/50 dark:border-green-500/10',
          textColor: 'text-green-700 dark:text-green-300',
          title: 'Query completed',
          subtitle: executionTime !== undefined 
            ? `Duration: ${executionTime < 1000 ? `${executionTime}ms` : `${(executionTime / 1000).toFixed(2)}s`} • 0 rows`
            : '0 rows returned',
        };

      case 'failed':
        return {
          icon: (
            <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          ),
          color: 'bg-red-50 dark:bg-red-500/10 border-red-200 dark:border-red-500/20',
          textColor: 'text-red-800 dark:text-red-300',
          title: 'Query failed',
          subtitle: message || 'Check the error details below',
        };

      case 'no_match':
        return {
          icon: (
            <svg className="w-5 h-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          ),
          color: 'bg-blue-50 dark:bg-blue-500/10 border-blue-200 dark:border-blue-500/20',
          textColor: 'text-blue-800 dark:text-blue-300',
          title: 'No rows match',
          subtitle: message || 'Your query would affect 0 rows',
        };

      case 'pending_approval':
        return {
          icon: (
            <svg className="w-5 h-5 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          ),
          color: 'bg-amber-50 dark:bg-amber-500/10 border-amber-200 dark:border-amber-500/20',
          textColor: 'text-amber-800 dark:text-amber-300',
          title: 'Pending approval',
          subtitle: message || 'Your query has been submitted for approval',
        };

      default:
        return {
          icon: null,
          color: 'bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700',
          textColor: 'text-gray-600 dark:text-gray-400',
          title: 'Ready',
          subtitle: 'Enter a query to execute',
        };
    }
  };

  const config = getStatusConfig();

  return (
    <motion.div
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      className={`flex items-center gap-3 px-4 py-2.5 rounded-xl border ${config.color} ${config.textColor}`}
    >
      {config.icon && <div className="flex-shrink-0">{config.icon}</div>}
      <div className="flex flex-col min-w-0">
        <span className="text-sm font-semibold truncate">{config.title}</span>
        {config.subtitle && (
          <span className="text-xs opacity-80 truncate">{config.subtitle}</span>
        )}
      </div>
    </motion.div>
  );
}

// Hook to manage query status
export function useQueryStatus() {
  const [status, setStatus] = useState<QueryExecutionStatus>('idle');
  const [executionTime, setExecutionTime] = useState<number>(0);
  const [rowCount, setRowCount] = useState<number>(0);
  const [message, setMessage] = useState<string>('');

  const reset = () => {
    setStatus('idle');
    setExecutionTime(0);
    setRowCount(0);
    setMessage('');
  };

  const setRunning = () => setStatus('running');
  
  const setCompleted = (time: number, rows: number) => {
    setExecutionTime(time);
    setRowCount(rows);
    setStatus(rows === 0 ? 'completed_empty' : 'completed');
  };

  const setFailed = (msg?: string) => {
    if (msg) setMessage(msg);
    setStatus('failed');
  };

  const setNoMatch = (msg?: string) => {
    if (msg) setMessage(msg);
    setStatus('no_match');
  };

  const setPendingApproval = (msg?: string) => {
    if (msg) setMessage(msg);
    setStatus('pending_approval');
  };

  return {
    status,
    executionTime,
    rowCount,
    message,
    reset,
    setRunning,
    setCompleted,
    setFailed,
    setNoMatch,
    setPendingApproval,
  };
}
