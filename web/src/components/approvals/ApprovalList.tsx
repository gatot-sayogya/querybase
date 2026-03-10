'use client';

import { useEffect, useState, useCallback } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { ApprovalRequest } from '@/types';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { staggerContainer, staggerItem, fadeIn, springConfig, duration, reducedMotionVariants } from '@/lib/animations';

interface ApprovalListProps {
  onSelectApproval: (approvalId: string) => void;
  selectedId: string | null;
  initialFilter?: 'all' | 'pending' | 'approved' | 'rejected';
}

export default function ApprovalList({
  onSelectApproval,
  selectedId,
  initialFilter = 'pending',
}: ApprovalListProps) {
  const [approvals, setApprovals] = useState<ApprovalRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'rejected'>(initialFilter);
  const [counts, setCounts] = useState<Record<string, number>>({
    all: 0,
    pending: 0,
    approved: 0,
    rejected: 0
  });

  const [formatter, setFormatter] = useState<Intl.DateTimeFormat | null>(null);
  const shouldReduceMotion = useReducedMotion();

  const fetchApprovals = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const countsData = await apiClient.getApprovalCounts();
      setCounts(countsData);

      const params = filter === 'all' ? {} : { status: filter };
      const data = await apiClient.getApprovals(params);
      setApprovals(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load approvals');
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    setFormatter(new Intl.DateTimeFormat(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }));
  }, []);

  useEffect(() => {
    fetchApprovals();
  }, [fetchApprovals]);

  const getStatusBadgeColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'badge-amber';
      case 'approved':
        return 'badge-green';
      case 'rejected':
        return 'badge-red text-red-700 bg-red-50';
      default:
        return 'badge-slate';
    }
  };

  const containerVariants = shouldReduceMotion ? reducedMotionVariants : staggerContainer;
  const itemVariants = shouldReduceMotion ? reducedMotionVariants : staggerItem;
  const filters = ['all', 'pending', 'approved', 'rejected'] as const;
  const activeFilterIndex = filters.indexOf(filter);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>
      <motion.div
        style={{ display: 'flex', gap: '8px', marginBottom: '16px', position: 'relative' }}
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: duration.normal }}
      >
        {filters.map((status, index) => {
          const count = counts[status] || 0;

          return (
            <motion.button
              key={status}
              onClick={() => setFilter(status)}
              className={`btn btn-sm flex items-center gap-1.5 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none relative z-10 ${
                filter === status
                  ? 'text-blue-600'
                  : 'btn-ghost'
              }`}
              whileHover={!shouldReduceMotion ? { scale: 1.02 } : {}}
              whileTap={!shouldReduceMotion ? { scale: 0.98 } : {}}
            >
              {status.charAt(0).toUpperCase() + status.slice(1)}
              <motion.span
                className={`inline-flex items-center justify-center px-1.5 min-w-[18px] h-[18px] text-[10px] font-bold rounded-full ${
                  filter === status
                    ? 'bg-blue-100 text-blue-600'
                    : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
                }`}
                key={count}
                initial={!shouldReduceMotion ? { scale: 1.2 } : {}}
                animate={{ scale: 1 }}
                transition={{ duration: 0.2 }}
              >
                {count}
              </motion.span>
            </motion.button>
          );
        })}

        {!shouldReduceMotion && (
          <motion.div
            className="absolute bottom-0 h-0.5 bg-blue-500 rounded-full"
            initial={false}
            animate={{
              x: activeFilterIndex * 90,
              width: 70,
            }}
            transition={{ duration: 0.25, ...springConfig.snappy }}
          />
        )}
      </motion.div>

      <AnimatePresence mode="wait">
        {loading ? (
          <motion.div
            key="loading"
            style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
          >
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </motion.div>
        ) : error ? (
          <motion.div
            key="error"
            className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 m-4"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
          >
            <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
            <button
              onClick={fetchApprovals}
              className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
            >
              Retry
            </button>
          </motion.div>
        ) : approvals.length === 0 ? (
          <motion.div
            key="empty"
            style={{ padding: '32px 0', textAlign: 'center' }}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
          >
            <motion.div
              style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '50%', background: 'var(--bg-hover)', color: 'var(--text-muted)', marginBottom: '16px' }}
              initial={!shouldReduceMotion ? { scale: 0.8, rotate: -10 } : {}}
              animate={{ scale: 1, rotate: 0 }}
              transition={{ delay: 0.2, ...springConfig.bouncy }}
            >
              <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </motion.div>
            <motion.h3
              style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text-primary)' }}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.3 }}
            >
              No approvals found
            </motion.h3>
            <motion.p
              style={{ marginTop: '4px', fontSize: '14px', color: 'var(--text-muted)' }}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.4 }}
            >
              {filter === 'all'
                ? 'No approval requests yet'
                : `No ${filter} approval requests`}
            </motion.p>
          </motion.div>
        ) : (
          <motion.div
            key="list"
            className="request-items"
            style={{ flex: 1 }}
            variants={containerVariants}
            initial="initial"
            animate="animate"
          >
            {approvals.map((approval) => (
              <motion.button
                key={approval.id}
                onClick={() => onSelectApproval(approval.id)}
                className={`request-item w-full text-left focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-blue-500 focus-visible:outline-none transition-colors ${selectedId === approval.id ? 'active' : ''}`}
                variants={itemVariants}
                whileHover={!shouldReduceMotion ? { x: 4 } : {}}
                whileTap={!shouldReduceMotion ? { scale: 0.99 } : {}}
              >
                <div className="req-top flex justify-between items-start mb-1 gap-2">
                  <span className="req-op truncate font-medium flex-1">
                    {approval.operation_type ? `${approval.operation_type.toUpperCase()} operation` : 'UNKNOWN operation'}
                  </span>
                  <span className={`badge shrink-0 ${getStatusBadgeColor(approval.status)}`}>
                    {approval.status}
                  </span>
                </div>
                <div className="req-user text-sm text-gray-500 dark:text-gray-400 truncate">
                  Requested by {approval.requester_name || approval.requester_id} &middot; {formatter ? formatter.format(new Date(approval.created_at)) : ''}
                </div>
              </motion.button>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
