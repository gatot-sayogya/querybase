'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
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
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Detail Header */}
      <div className="detail-header">
        <div>
          <div className="detail-title">
            {approval.operation_type ? approval.operation_type.toUpperCase() : 'UNKNOWN'} Statement Review
          </div>
          <div style={{ marginTop: '6px' }}>
            <span className={`badge ${getStatusBadgeColor(approval.status)}`}>
              {approval.status}
            </span>
          </div>
        </div>
        <div style={{ fontSize: '12px', color: 'var(--text-muted)', textAlign: 'right' }}>
          Requested by <strong>{approval.requester_name || approval.requester_id}</strong><br/>
          {formatDate(approval.created_at)}
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}

      {/* SQL Statement */}
      <div className="detail-section-label">SQL Statement</div>
      <div className="code-block" style={{ marginBottom: '20px' }}>
        <pre style={{ margin: 0, fontFamily: 'inherit', whiteSpace: 'pre-wrap' }}>
          {approval.query_text}
        </pre>
      </div>

      {/* Data Source */}
      <div className="detail-section-label">Data Source</div>
      <div style={{ fontSize: '14px', color: 'var(--text-primary)', marginBottom: '20px' }}>
        {approval.data_source_name || approval.data_source_id || 'Unknown Data Source'}
      </div>

      {/* Comment & Actions */}
      {isPending && (
        <>
          <div className="detail-section-label">Comments (optional)</div>
          <textarea
            className="comment-area"
            placeholder="Add a comment for the requester..."
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            disabled={submitting}
          />
          <div className="detail-actions">
            <button
              className="btn btn-danger"
              disabled={submitting}
              onClick={() => handleReview('rejected')}
              style={submitting ? { opacity: 0.5, cursor: 'not-allowed' } : {}}
            >
              ✕ Reject
            </button>
            <button
              className="btn btn-success"
              disabled={submitting}
              onClick={() => handleReview('approved')}
              style={submitting ? { opacity: 0.5, cursor: 'not-allowed' } : {}}
            >
              ✔ Approve
            </button>
          </div>
        </>
      )}

      {/* Info for non-pending */}
      {!isPending && (
        <div style={{ padding: '12px', background: 'var(--bg-hover)', borderRadius: 'var(--r-md)', marginTop: '20px' }}>
          <p style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
            This approval has been <strong>{approval.status}</strong>. No further actions can be taken.
          </p>
        </div>
      )}
    </div>
  );
}
