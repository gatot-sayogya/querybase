import { apiClient } from '@/lib/api-client';

export interface StatementPreview {
  sequence: number;
  query_text: string;
  operation_type: string;
  estimated_rows: number;
  preview_rows?: Record<string, unknown>[];
  columns?: { name: string; type: string }[];
  error?: string;
}

export interface MultiQueryPreviewResponse {
  statement_count: number;
  total_estimated_rows: number;
  statements: StatementPreview[];
  requires_approval: boolean;
}

export interface StatementResult {
  sequence: number;
  query_text: string;
  operation_type: string;
  status: string;
  affected_rows: number;
  row_count: number;
  columns?: { name: string; type: string }[];
  data?: Record<string, unknown>[];
  error_message?: string;
  execution_time_ms: number;
}

export interface MultiQueryResponse {
  query_id?: string;
  transaction_id?: string;
  status: string;
  is_multi_query: boolean;
  statement_count: number;
  total_affected_rows: number;
  execution_time_ms: number;
  statements: StatementResult[];
  error_message?: string;
  requires_approval?: boolean;
  approval_id?: string;
}

/**
 * Preview multiple queries before execution
 */
export async function previewMultiQuery(
  dataSourceId: string,
  queryTexts: string[]
): Promise<MultiQueryPreviewResponse> {
  return apiClient.previewMultiQuery(dataSourceId, queryTexts);
}

/**
 * Execute multiple queries in a transaction
 */
export async function executeMultiQuery(
  dataSourceId: string,
  queryTexts: string[],
  name?: string,
  description?: string
): Promise<MultiQueryResponse> {
  return apiClient.executeMultiQuery(dataSourceId, queryTexts, name, description);
}

/**
 * Get statement details for a multi-query transaction
 */
export async function getMultiQueryStatements(
  transactionId: string
): Promise<StatementResult[]> {
  return apiClient.getMultiQueryStatements(transactionId);
}

/**
 * Commit a multi-query transaction
 */
export async function commitMultiQuery(
  transactionId: string
): Promise<MultiQueryResponse> {
  return apiClient.commitMultiQuery(transactionId);
}

/**
 * Rollback a multi-query transaction
 */
export async function rollbackMultiQuery(
  transactionId: string
): Promise<void> {
  return apiClient.rollbackMultiQuery(transactionId);
}
