'use client';

import { useEffect, useState } from 'react';
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

  useEffect(() => {
    fetchApprovals();
  }, [filter]);

  const fetchApprovals = async () => {
    try {
      setLoading(true);
      setError(null);
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
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300';
      case 'approved':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'rejected':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="animate-pulse">
            <div className="h-24 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        <button
          onClick={fetchApprovals}
          className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex space-x-2">
        {(['all', 'pending', 'approved', 'rejected'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`px-3 py-1 text-sm rounded-md ${
              filter === status
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
            }`}
          >
            {status.charAt(0).toUpperCase() + status.slice(1)}
            {status !== 'all' && ` (${approvals.filter((a) => a.status === status).length})`}
          </button>
        ))}
      </div>

      {/* Approval List */}
      {approvals.length === 0 ? (
        <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
            No approvals found
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {filter === 'all'
              ? 'No approval requests yet'
              : `No ${filter} approval requests`}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {approvals.map((approval) => (
            <div
              key={approval.id}
              onClick={() => onSelectApproval(approval.id)}
              className={`p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                selectedId === approval.id
                  ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                  : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
              }`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-2">
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded ${getOperationBadgeColor(
                        approval.operation_type
                      )}`}
                    >
                      {approval.operation_type ? approval.operation_type.toUpperCase() : 'UNKNOWN'}
                    </span>
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded ${getStatusBadgeColor(
                        approval.status
                      )}`}
                    >
                      {approval.status.charAt(0).toUpperCase() + approval.status.slice(1)}
                    </span>
                  </div>
                  <p className="text-sm text-gray-900 dark:text-white font-mono mb-2">
                    {approval.query_text.length > 100
                      ? approval.query_text.substring(0, 100) + '...'
                      : approval.query_text}
                  </p>
                  <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
                    <span>
                      Requested: {formatDate(approval.created_at)}
                    </span>
                  </div>
                </div>
                <svg
                  className={`w-5 h-5 text-gray-400 transition-transform ${
                    selectedId === approval.id ? 'rotate-90' : ''
                  }`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
