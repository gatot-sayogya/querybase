'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { ApprovalRequest, ApprovalReview, TransactionPreview, AuditMode, WriteQueryPreview } from '@/types';
import DataChangesPanel from './DataChangesPanel';
import { motion, AnimatePresence } from 'framer-motion';
import { isMultiQuery } from '@/lib/query-parser';
import { previewMultiQuery, executeMultiQuery } from '@/lib/api/multi-query';
import type { MultiQueryPreviewResponse } from '@/lib/api/multi-query';
import { MultiQueryPreviewModal } from '@/components/query/MultiQueryPreviewModal';
import InsertPreviewPanel from './InsertPreviewPanel';
import type { InsertPreviewResult } from '@/lib/api/insert-preview';

interface ApprovalDetailProps {
  approvalId: string | null;
  onRefresh?: () => void;
}

type PreviewPhase = 'idle' | 'loading_preview' | 'preview_ready' | 'loading_tx' | 'tx_ready';

export default function ApprovalDetail({ approvalId, onRefresh }: ApprovalDetailProps) {
  const [approval, setApproval] = useState<ApprovalRequest | null>(null);
  const [reviews, setReviews] = useState<ApprovalReview[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // 3-phase transaction state
  const [phase, setPhase] = useState<PreviewPhase>('idle');
  const [writePreview, setWritePreview] = useState<WriteQueryPreview | null>(null);
  const [multiQueryPreview, setMultiQueryPreview] = useState<MultiQueryPreviewResponse | null>(null);
  const [insertPreview, setInsertPreview] = useState<InsertPreviewResult | null>(null);
  const [transaction, setTransaction] = useState<TransactionPreview | null>(null);
  const [auditMode, setAuditMode] = useState<AuditMode>('full');
  
  // Hydration safe date formatter
  const [formatter, setFormatter] = useState<Intl.DateTimeFormat | null>(null);

  useEffect(() => {
    setFormatter(new Intl.DateTimeFormat(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }));
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
      setPhase('idle');
      setWritePreview(null);
      setTransaction(null);
      setComment('');
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
      toast.success(decision === 'approved' ? 'Approval granted' : 'Request rejected');
      await fetchApprovalDetails();
      if (onRefresh) onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : `Failed to ${decision} approval`);
      setError(err instanceof Error ? err.message : `Failed to ${decision} approval`);
    } finally {
      setSubmitting(false);
    }
  };

  // Phase 1: Dry-run preview (no changes to database)
  const handlePreviewQuery = async () => {
    if (!approval) return;
    setPhase('loading_preview');
    setError(null);
    try {
      const operationType = approval.operation_type?.toUpperCase();
      
      // Check if this is a multi-query
      if (isMultiQuery(approval.query_text)) {
        const preview = await previewMultiQuery(
          approval.data_source_id,
          [approval.query_text]
        );
        setMultiQueryPreview(preview);
        setPhase('preview_ready');
      } else if (operationType === 'INSERT') {
        // NEW: INSERT preview
        const preview = await apiClient.previewInsertQuery(
          approval.data_source_id,
          approval.query_text
        );
        setInsertPreview(preview);
        setPhase('preview_ready');
      } else if (operationType === 'UPDATE' || operationType === 'DELETE') {
        // Existing UPDATE/DELETE preview
        const preview = await apiClient.previewWriteQuery(
          approval.data_source_id,
          approval.query_text
        );
        setWritePreview(preview);
        setPhase('preview_ready');
      } else {
        throw new Error(`Unsupported operation type: ${operationType}`);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to fetch preview data');
      setError(err instanceof Error ? err.message : 'Failed to fetch preview data');
      setPhase('idle');
    }
  };

  // Phase 2: Start actual transaction (runs DELETE/UPDATE held in DB tx)
  const handleStartTransaction = async () => {
    if (!approvalId) return;
    setPhase('loading_tx');
    setSubmitting(true);
    setError(null);
    try {
      const txData = await apiClient.startApprovalTransaction(approvalId, { audit_mode: auditMode });
      setTransaction(txData);
      if (txData.preview?.audit_mode) {
        setAuditMode(txData.preview.audit_mode);
      }
      setPhase('tx_ready');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to start transaction');
      setError(err instanceof Error ? err.message : 'Failed to start transaction');
      setPhase('preview_ready'); // Fall back to preview step
    } finally {
      setSubmitting(false);
    }
  };

  const handleCommit = async () => {
    if (!transaction?.transaction_id) return;

    setSubmitting(true);
    setError(null);
    try {
      await apiClient.commitTransaction(transaction.transaction_id, { audit_mode: auditMode });

      // Post a review comment only if the approver typed one.
      // Commit already marks the approval as approved, so /review is optional here.
      if (comment.trim()) {
        try {
          await apiClient.reviewApproval(approvalId!, { decision: 'approved', comments: comment });
        } catch {
          // /review may fail if approval was already marked approved by commit — that's fine.
        }
      }

      toast.success('Transaction committed successfully');
      setTransaction(null);
      await fetchApprovalDetails();
      if (onRefresh) onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to commit transaction');
      setError(err instanceof Error ? err.message : 'Failed to commit transaction');
    } finally {
      setSubmitting(false);
    }
  };

  const handleRollback = async () => {
    if (!transaction?.transaction_id) return;

    setSubmitting(true);
    try {
      await apiClient.rollbackTransaction(transaction.transaction_id);
      toast.success('Transaction rolled back');
      setTransaction(null);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to rollback transaction');
      setError(err instanceof Error ? err.message : 'Failed to rollback transaction');
    } finally {
      setSubmitting(false);
    }
  };

  // Handle multi-query execution from preview modal
  const handleExecuteMultiQuery = async () => {
    if (!multiQueryPreview || !approval || !approvalId) return;
    
    setSubmitting(true);
    try {
      // Use the approval workflow - start transaction then commit
      // This ensures we're executing the approved query, not creating a new one
      const txData = await apiClient.startApprovalTransaction(approvalId, { audit_mode: auditMode });
      setTransaction(txData);
      
      if (txData.preview?.audit_mode) {
        setAuditMode(txData.preview.audit_mode);
      }
      
      // Close the modal and move to transaction ready phase
      setMultiQueryPreview(null);
      setPhase('tx_ready');
      
      toast.success('Transaction started. Review the execution preview and click Commit to apply changes.');
    } catch (err: any) {
      toast.error(err.response?.data?.error || err.message || 'Failed to start multi-query transaction');
      setError(err.response?.data?.error || err.message || 'Failed to start multi-query transaction');
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
      <div className="bg-gray-50 dark:bg-gray-800 p-8 text-center">
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
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-4">
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
          {formatter ? formatter.format(new Date(approval.created_at)) : ''}
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-4 mb-4">
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

      {/* Phase 1: Idle — initial actions before preview */}
      {isPending && phase === 'idle' && approval.can_approve && (
        <div className="mt-6 border-t border-gray-200 dark:border-gray-700 pt-6">
          <div className="detail-section-label">Comments (optional)</div>
          <textarea
            className="comment-area focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none w-full p-3 border"
            placeholder="Add a comment for the requester..."
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            disabled={submitting}
            rows={3}
          />
          <div className="detail-actions mt-4 flex justify-end gap-3">
            <button
              className="btn btn-danger focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center min-w-[100px]"
              disabled={submitting}
              onClick={() => handleReview('rejected')}
            >
              {submitting ? <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div> : '✕ Reject'}
            </button>
            <button
              className="btn btn-primary focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center min-w-[180px]"
              disabled={false}
              onClick={handlePreviewQuery}
            >
              ► Test & Preview Execution
            </button>
          </div>
        </div>
      )}

      {/* Phase 1 loading spinner */}
      {isPending && phase === 'loading_preview' && (
        <div className="mt-6 border-t border-gray-200 dark:border-gray-700 pt-6">
          <div className="flex items-center gap-3 text-blue-600 dark:text-blue-400">
            <div className="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            <span className="text-sm font-medium">Fetching preview data...</span>
          </div>
        </div>
      )}

      {/* Phase 2: Preview ready — show the dry-run data table */}
      {isPending && phase === 'preview_ready' && writePreview && approval.can_approve && (
          <div className="mt-6 border-t border-orange-200 dark:border-orange-900 pt-6 bg-orange-50 dark:bg-orange-900/10 -mx-6 px-6 pb-6">
          <h3 className="text-lg font-medium text-orange-900 dark:text-orange-300 mb-4 flex items-center gap-2">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
            Preview — Rows to be Affected
          </h3>

          {/* Impact Summary identical to WritePreviewModal */}
          <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 overflow-hidden mb-4">
            <div className="p-5 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-6">
                <div>
                  <span className="text-sm text-gray-500 dark:text-gray-400">Total Rows Affected</span>
                  <div className={`text-3xl font-bold font-mono ${approval.operation_type?.toLowerCase() === 'delete' ? 'text-red-600 dark:text-red-400' : 'text-yellow-600 dark:text-yellow-400'}`} style={{ fontVariantNumeric: 'tabular-nums' }}>
                    {writePreview.total_affected}
                  </div>
                </div>
                {writePreview.preview_rows?.length > 0 && (
                  <>
                    <div className="h-12 w-px bg-gray-200 dark:bg-gray-700" />
                    <div>
                      <span className="text-sm text-gray-500 dark:text-gray-400">Showing Preview</span>
                      <div className="text-lg font-medium text-gray-900 dark:text-gray-100">
                        {writePreview.preview_rows.length} of {writePreview.total_affected} rows
                      </div>
                    </div>
                  </>
                )}
                {writePreview.total_affected > (writePreview.preview_rows?.length || 0) && (
                  <>
                    <div className="h-12 w-px bg-gray-200 dark:bg-gray-700" />
                    <div className="flex items-center gap-2 p-2 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800">
                      <svg className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                      <span className="text-sm text-amber-800 dark:text-amber-300">
                        Large impact — sample only
                      </span>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Data table */}
            {!writePreview.preview_rows || writePreview.preview_rows.length === 0 ? (
              <div className="p-8 text-center text-gray-500 dark:text-gray-400">
                <svg className="mx-auto h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                </svg>
                No rows match the query conditions.
              </div>
            ) : (
              <div className="overflow-auto max-h-72">
                <table className="data-table w-full text-sm m-0 border-0">
                  <thead className="sticky top-0 z-10">
                    <tr>
                      <th className="bg-gray-50 dark:bg-gray-800 text-center text-gray-400 w-12 border-t-0 border-l-0">#</th>
                      {writePreview.columns.map((col, i) => (
                        <th key={i} className="bg-gray-50 dark:bg-gray-800 border-t-0">{col}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {writePreview.preview_rows.map((row, rowIdx) => (
                      <tr key={rowIdx} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                        <td className="text-center text-gray-400 font-mono text-xs border-l-0" style={{ background: 'var(--bg-hover)' }}>{rowIdx + 1}</td>
                        {writePreview.columns.map((col, colIdx) => {
                          const val = (row as Record<string, unknown>)[col];
                          return (
                            <td key={colIdx}>
                              {val === null || val === undefined ? (
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
              </div>
            )}

            {writePreview.total_affected > (writePreview.preview_rows?.length || 0) && (
              <div className="p-3 text-center text-sm text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
                Showing {writePreview.preview_rows?.length || 0} of {writePreview.total_affected.toLocaleString()} total affected rows
              </div>
            )}
          </div>

          {/* Actions for preview phase */}
          <div className="detail-actions flex justify-between gap-3 items-center">
            <button
              className="btn btn-danger focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-50 min-w-[100px]"
              disabled={submitting}
              onClick={() => handleReview('rejected')}
            >
              ✕ Reject
            </button>
            <div className="flex gap-3">
              <button
                className="btn btn-ghost text-gray-600 dark:text-gray-400 disabled:opacity-50"
                disabled={submitting}
                onClick={() => { setPhase('idle'); setWritePreview(null); }}
              >
                ← Back
              </button>
              <button
                className="btn btn-success focus-visible:ring-2 focus-visible:ring-green-500 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-50 flex items-center justify-center min-w-[220px]"
                disabled={submitting}
                onClick={async () => {
                  // First approve the request, then start transaction
                  await handleReview('approved');
                  // After approval is granted, start the transaction
                  handleStartTransaction();
                }}
              >
                {submitting
                  ? <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                  : '✔ Approve & Start Transaction'
                }
              </button>
            </div>
          </div>
        </div>
      )}

      {/* INSERT Preview Panel */}
      {isPending && phase === 'preview_ready' && insertPreview && approval.can_approve && (
        <div className="mt-6 border-t border-green-200 dark:border-green-900 pt-6 bg-green-50 dark:bg-green-900/10 -mx-6 px-6 pb-6">
          <InsertPreviewPanel
            preview={insertPreview}
            onProceed={async () => {
              // First approve the request, then start transaction
              await handleReview('approved');
              // After approval is granted, start the transaction
              handleStartTransaction();
            }}
            onCancel={() => { 
              setPhase('idle'); 
              setInsertPreview(null); 
            }}
          />
        </div>
      )}

      {/* Multi-Query Preview Modal */}
      {multiQueryPreview && (
        <MultiQueryPreviewModal
          isOpen={!!multiQueryPreview}
          onClose={() => { 
            setMultiQueryPreview(null); 
            setPhase('idle');
          }}
          statements={multiQueryPreview.statements}
          totalEstimatedRows={multiQueryPreview.total_estimated_rows}
          onApprove={handleExecuteMultiQuery}
          onReject={() => { 
            setMultiQueryPreview(null); 
            setPhase('idle');
          }}
          loading={submitting}
        />
      )}

      {/* Phase 3: Transaction open — Commit/Rollback */}
      {isPending && phase === 'tx_ready' && transaction && (
          <div className="mt-6 border-t border-blue-200 dark:border-blue-900 pt-6 bg-blue-50 dark:bg-blue-900/10 -mx-6 px-6 pb-6">
          <h3 className="text-lg font-medium text-blue-900 dark:text-blue-300 mb-4 flex items-center gap-2">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            Execution Preview Ready
          </h3>
          
          <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 overflow-hidden mb-4">
            
            {/* Impact Summary */}
            <div className="p-5 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-6">
                <div>
                  <span className="text-sm text-gray-500 dark:text-gray-400">Estimated Impact</span>
                  <div className={`text-3xl font-bold font-mono ${approval.operation_type?.toLowerCase() === 'delete' ? 'text-red-600 dark:text-red-400' : 'text-yellow-600 dark:text-yellow-400'}`} style={{ fontVariantNumeric: 'tabular-nums' }}>
                    {transaction.preview?.estimated_rows ?? 0}
                  </div>
                </div>
                {transaction.preview?.data && (
                  <>
                    <div className="h-12 w-px bg-gray-200 dark:bg-gray-700" />
                    <div>
                      <span className="text-sm text-gray-500 dark:text-gray-400">Showing Preview</span>
                      <div className="text-lg font-medium text-gray-900 dark:text-gray-100">
                        {transaction.preview.data.length} of {transaction.preview.estimated_rows ?? 0} rows
                      </div>
                    </div>
                  </>
                )}
                
                {transaction.preview?.estimated_rows > (transaction.preview?.data?.length || 0) && (
                  <>
                    <div className="h-12 w-px bg-gray-200 dark:bg-gray-700" />
                    <div className="flex items-center gap-2 p-2 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800">
                      <svg className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                      <span className="text-sm text-amber-800 dark:text-amber-300">
                        Large impact — showing sample rows only
                      </span>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Preview Data Table */}
            {transaction.preview?.data ? (
              transaction.preview.data.length === 0 ? (
                <div className="p-8 text-center text-gray-500 dark:text-gray-400">
                  <svg className="mx-auto h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                  </svg>
                  No rows match the query conditions.
                </div>
              ) : (
                <div className="overflow-auto max-h-64 bg-white dark:bg-gray-800">
                  <table className="data-table w-full text-sm m-0 border-0">
                    <thead className="sticky top-0 z-10">
                      <tr>
                        <th className="bg-gray-50 dark:bg-gray-800 text-center text-gray-400 w-12 border-t-0 border-l-0">#</th>
                        {transaction.preview.columns?.map((col: any, i: number) => (
                          <th key={i} className="bg-gray-50 dark:bg-gray-800 border-t-0">{typeof col === 'object' ? col.name : col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {transaction.preview.data.map((row: Record<string, any>, rowIdx: number) => (
                        <tr key={rowIdx} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                          <td className="text-center text-gray-400 font-mono text-xs border-l-0" style={{ background: 'var(--bg-hover)' }}>{rowIdx + 1}</td>
                          {(transaction.preview?.columns || []).map((col: any, colIdx: number) => {
                            const colName = typeof col === 'object' ? col.name : col;
                            const val = row[colName];
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
                            );
                          })}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )
            ) : (
              <div className="p-4 text-center text-sm text-gray-500 italic">
                Execution preview metadata only. Exact row records omitted.
              </div>
            )}
            
            {/* Caution Banner */}
            {transaction.preview?.caution && (
              <div className="p-4 bg-amber-50 dark:bg-amber-900/20 border-t border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-300 text-sm flex gap-2">
                <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <div className="flex-1">
                  <strong>Caution: Large impact detected.</strong> {transaction.preview.caution_message || 'This query modifies a large amount of data.'}
                </div>
              </div>
            )}
          </div>
          
          <div className="bg-white dark:bg-gray-800 p-4 mb-4 border border-gray-200 dark:border-gray-700">
            <div className="mb-4">
              <label htmlFor="audit-mode" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Audit Capture Mode
              </label>
              <select
                id="audit-mode"
                className="w-full sm:w-64 p-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none"
                value={auditMode}
                onChange={(e) => setAuditMode(e.target.value as AuditMode)}
                disabled={submitting}
              >
                {transaction.preview?.audit_mode === 'count_only' && (
                  <option value="count_only">Count Only (Target Database capability limit)</option>
                )}
                {transaction.preview?.audit_mode !== 'count_only' && (
                  <>
                    <option value="full">Full Context (Save all before/after row data)</option>
                    <option value="sample">Sample (Save first 50 rows only)</option>
                    <option value="count_only">Count Only (Fastest, no row data saved)</option>
                  </>
                )}
              </select>
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Determines how much historical data is saved to the audit log for rollback/review.
              </p>
            </div>
            
            <div className="detail-section-label">Comments (optional)</div>
            <textarea
            className="comment-area focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none w-full p-3 border"
              placeholder="Add final commit notes..."
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              disabled={submitting}
              rows={2}
            />
          </div>

          <div className="detail-actions mt-4 flex justify-between gap-3 items-center">
            <button
              className="btn btn-ghost text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:outline-none disabled:opacity-50 min-w-[100px]"
              disabled={submitting}
              onClick={handleRollback}
            >
              Abort & Run Rollback
            </button>
            <button
              className="btn btn-success focus-visible:ring-2 focus-visible:ring-green-500 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-50 flex items-center justify-center min-w-[200px]"
              disabled={submitting}
              onClick={handleCommit}
            >
               {submitting ? <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div> : '✔ Commit & Approve'}
            </button>
          </div>
        </div>
      )}

      {/* Info for pending but cannot approve */}
      {isPending && !approval.can_approve && (
        <div style={{ padding: '12px', background: 'var(--bg-hover)', borderRadius: 'var(--r-md)', marginTop: '20px' }}>
          <p style={{ fontSize: '13px', color: 'var(--text-primary)' }}>
            This approval request is waiting for review by an approver.
          </p>
        </div>
      )}

      {/* Execution Results (Approved) */}
      {!isPending && approval.status === 'approved' && approval.transaction && (
        <DataChangesPanel
          affectedRows={approval.transaction.affected_rows}
          auditMode={approval.transaction.audit_mode}
          beforeData={approval.transaction.before_data}
          afterData={approval.transaction.after_data}
          completedAt={approval.transaction.completed_at}
          reviewerName={approval.reviews && approval.reviews.length > 0 ? (approval.reviews[0].reviewer_name || 'Approver') : undefined}
        />
      )}

      {/* Info for non-pending (Rejected or missing transaction) */}
      {!isPending && (approval.status === 'rejected' || (!approval.transaction && approval.status === 'approved')) && (
        <div style={{ padding: '12px', background: 'var(--bg-hover)', borderRadius: 'var(--r-md)', marginTop: '20px' }}>
          <p style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
            This approval has been <strong>{approval.status}</strong>. No further actions can be taken.
          </p>
        </div>
      )}
    </div>
  );
}
