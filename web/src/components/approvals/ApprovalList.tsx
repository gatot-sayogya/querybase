'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { ApprovalRequest } from '@/types';
import { formatDate } from '@/lib/utils';

interface ApprovalListProps {
  onSelectApproval: (approvalId: string) => void;
  selectedId: string | null;
}

export default function ApprovalList({
  onSelectApproval,
  selectedId,
}: ApprovalListProps) {
  const [approvals, setApprovals] = useState<ApprovalRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'rejected'>('pending');
  const [counts, setCounts] = useState<Record<string, number>>({
    all: 0,
    pending: 0,
    approved: 0,
    rejected: 0
  });

  useEffect(() => {
    fetchApprovals();
  }, [filter]);

  const fetchApprovals = async () => {
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
  };

  const getOperationBadgeColor = (operationType: string | null | undefined) => {
    if (!operationType) {
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
    switch (operationType.toUpperCase()) {
      case 'SELECT':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      case 'INSERT':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'UPDATE':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300';
      case 'DELETE':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
  };

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



  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>
      {/* Filters */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
        {(['all', 'pending', 'approved', 'rejected'] as const).map((status) => {
          const count = counts[status] || 0;
            
          return (
            <button
              key={status}
              onClick={() => setFilter(status)}
              className={`btn btn-sm flex items-center gap-1.5 ${
                filter === status
                  ? 'btn-primary'
                  : 'btn-ghost'
              }`}
            >
              {status.charAt(0).toUpperCase() + status.slice(1)}
              {/* Count badge - always show for consistency */}
              <span className={`inline-flex items-center justify-center px-1.5 min-w-[18px] h-[18px] text-[10px] font-bold rounded-full ${
                filter === status 
                  ? 'bg-white/20 text-white' 
                  : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
              }`}>
                {count}
              </span>
            </button>
          );
        })}
      </div>

      {/* Approval List */}
      {loading ? (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : error ? (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 m-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          <button
            onClick={fetchApprovals}
            className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
          >
            Retry
          </button>
        </div>
      ) : approvals.length === 0 ? (
        <div style={{ padding: '32px 0', textAlign: 'center' }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '50%', background: 'var(--bg-hover)', color: 'var(--text-muted)', marginBottom: '16px' }}>
            <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h3 style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text-primary)' }}>
            No approvals found
          </h3>
          <p style={{ marginTop: '4px', fontSize: '14px', color: 'var(--text-muted)' }}>
            {filter === 'all'
              ? 'No approval requests yet'
              : `No ${filter} approval requests`}
          </p>
        </div>
      ) : (
        <div className="request-items" style={{ flex: 1 }}>
          {approvals.map((approval) => (
            <div
              key={approval.id}
              onClick={() => onSelectApproval(approval.id)}
              className={`request-item ${selectedId === approval.id ? 'active' : ''}`}
            >
              <div className="req-top">
                <span className="req-op">
                  {approval.operation_type ? `${approval.operation_type.toUpperCase()} operation` : 'UNKNOWN operation'}
                </span>
                <span className={`badge ${getStatusBadgeColor(approval.status)}`}>
                  {approval.status}
                </span>
              </div>
              <div className="req-user">
                Requested by {approval.requester_name || approval.requester_id} &middot; {formatDate(approval.created_at)}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
