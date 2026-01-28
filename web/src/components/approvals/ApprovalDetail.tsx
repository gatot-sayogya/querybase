'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { ApprovalRequest, ApprovalReview } from '@/types';
import { formatDate } from '@/lib/utils';

interface ApprovalDetailProps {
  approvalId: string | null;
  onRefresh?: () => void;
}

export default function ApprovalDetail({ approvalId, onRefresh }: ApprovalDetailProps) {
  const [approval, setApproval] = useState<ApprovalRequest | null>(null);
  const [reviews, setReviews] = useState<ApprovalReview[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (approvalId) {
      fetchApprovalDetails();
    }
  }, [approvalId]);

  const fetchApprovalDetails = async () => {
    if (!approvalId) return;

    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getApproval(approvalId);
      setApproval(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load approval details');
    } finally {
      setLoading(false);
    }
  };

  const handleReview = async (decision: 'approved' | 'rejected') => {
    if (!approvalId) return;

    setSubmitting(true);
    try {
      await apiClient.reviewApproval(approvalId, { decision, comments: comment });
      setComment('');
      await fetchApprovalDetails();
      if (onRefresh) onRefresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${decision} approval`);
    } finally {
      setSubmitting(false);
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
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading approval details...</p>
        </div>
      </div>
    );
  }

  if (!approvalId) {
    return (
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
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
          Select an approval
        </h3>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Choose an approval request from the list to view details
        </p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      </div>
    );
  }

  if (!approval) {
    return null;
  }

  const isPending = approval.status === 'pending';

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-700 pb-4">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
              Approval Request Details
            </h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Request ID: {approval.id.slice(0, 8)}...
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <span
              className={`px-3 py-1 text-sm font-medium rounded ${getOperationBadgeColor(
                approval.operation_type
              )}`}
            >
              {approval.operation_type ? approval.operation_type.toUpperCase() : 'UNKNOWN'}
            </span>
            <span
              className={`px-3 py-1 text-sm font-medium rounded ${getStatusBadgeColor(
                approval.status
              )}`}
            >
              {approval.status.charAt(0).toUpperCase() + approval.status.slice(1)}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500 dark:text-gray-400">Created:</span>
            <span className="ml-2 text-gray-900 dark:text-white">
              {formatDate(approval.created_at)}
            </span>
          </div>
          <div>
            <span className="text-gray-500 dark:text-gray-400">Updated:</span>
            <span className="ml-2 text-gray-900 dark:text-white">
              {formatDate(approval.updated_at)}
            </span>
          </div>
        </div>
      </div>

      {/* SQL Query */}
      <div>
        <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-2">
          SQL Query
        </h3>
        <div className="bg-gray-900 dark:bg-black rounded-lg p-4 overflow-x-auto">
          <pre className="text-sm text-gray-100 font-mono whitespace-pre-wrap">
            {approval.query_text}
          </pre>
        </div>
      </div>

      {/* Comment & Actions */}
      {isPending && (
        <div className="space-y-4">
          <div>
            <label
              htmlFor="comment"
              className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
            >
              Review Comment (Optional)
            </label>
            <textarea
              id="comment"
              rows={3}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              disabled={submitting}
              className="block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-800 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
              placeholder="Add a comment explaining your decision..."
            />
          </div>

          <div className="flex space-x-3">
            <button
              onClick={() => handleReview('approved')}
              disabled={submitting}
              className="flex-1 px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? 'Processing...' : 'Approve'}
            </button>
            <button
              onClick={() => handleReview('rejected')}
              disabled={submitting}
              className="flex-1 px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? 'Processing...' : 'Reject'}
            </button>
          </div>
        </div>
      )}

      {/* Info for non-pending */}
      {!isPending && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <p className="text-sm text-blue-800 dark:text-blue-300">
            This approval has been{' '}
            <strong>{approval.status}</strong>. No further actions can be taken.
          </p>
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}
    </div>
  );
}
