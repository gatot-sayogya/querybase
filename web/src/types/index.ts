// User & Authentication Types
export interface User {
  id: string;
  email: string;
  username: string;
  full_name: string;
  role: 'admin' | 'user' | 'viewer';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  groups?: string[]; // List of group IDs the user belongs to
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

// Data Source Types
export interface DataSource {
  id: string;
  name: string;
  type: 'postgresql' | 'mysql';
  host: string;
  port: number;
  database_name: string;
  username: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  permissions?: DataSourcePermission[];
}

export interface DataSourcePermission {
  group_id: string;
  group_name: string;
  can_read: boolean;
  can_write: boolean;
  can_approve: boolean;
}

export interface CreateDataSourceRequest {
  name: string;
  type: 'postgresql' | 'mysql';
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
}

// Query Types
export interface Query {
  id: string;
  data_source_id: string;
  data_source_name?: string;
  query_text: string;
  name?: string;
  description?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'no_match' | 'pending_approval';
  created_at: string;
  executed_at?: string;
  row_count?: number;
  error_message?: string;
  // These fields are returned when executing a query
  data?: Record<string, unknown>[];
  columns?: ColumnInfo[];
  execution_time_ms?: number;
  requires_approval?: boolean;
  approval_id?: string;
}

export interface QueryResult {
  query_id: string;
  row_count: number;
  columns: ColumnInfo[];
  data: Record<string, unknown>[];
}

export interface ColumnInfo {
  name: string;
  type: string;
}

export interface ExecuteQueryRequest {
  data_source_id: string;
  query_text: string;
  name?: string;
  description?: string;
}

export interface PaginatedResults {
  query_id: string;
  row_count: number;
  columns: ColumnInfo[];
  data: Record<string, unknown>[];
  metadata: {
    page: number;
    per_page: number;
    total_pages: number;
    total_rows: number;
    has_next: boolean;
    has_prev: boolean;
  };
}

// Approval Types
export type AuditMode = 'full' | 'sample' | 'count_only';

export interface TransactionPreview {
  transaction_id: string;
  approval_id: string;
  data_source_id: string;
  query_text: string;
  started_by: string;
  status: 'active' | 'committed' | 'rolled_back' | 'failed';
  started_at: string;
  preview: {
    data: any;
    row_count: number;
    columns: any[];
    estimated_rows: number;
    caution: boolean;
    caution_message?: string;
    audit_mode: AuditMode;
  };
}

export interface ApprovalRequest {
  id: string;
  query_id?: string;
  requester_id: string;
  status: 'pending' | 'approved' | 'rejected';
  operation_type: string | null;
  query_text: string;
  data_source_id: string;
  data_source_name?: string;
  requester_name?: string;
  can_approve?: boolean;
  created_at: string;
  updated_at: string;
  reviews?: any[];
  transaction?: {
    transaction_id?: string;
    status?: 'active' | 'committed' | 'failed';
    affected_rows: number;
    audit_mode: AuditMode;
    preview?: any;
    before_data?: Record<string, unknown>[];
    after_data?: Record<string, unknown>[];
    completed_at?: string;
  };
}

export interface ApprovalReview {
  id: string;
  approval_request_id: string;
  reviewer_id: string;
  decision: 'approved' | 'rejected';
  comments?: string;
  created_at: string;
}

export interface ReviewApprovalRequest {
  decision: 'approved' | 'rejected';
  comments?: string;
}

export interface StartTransactionRequest {
  audit_mode?: AuditMode;
}

export interface WriteQueryPreview {
  total_affected: number;
  preview_rows: Record<string, unknown>[];
  columns: string[];
  preview_limit: number;
  select_query: string;
  operation_type: string;
}

export interface CommitTransactionRequest {
  audit_mode?: AuditMode;
}

// Group Types
export interface Group {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  users?: User[];
  data_sources?: DataSource[];
}

export interface UserGroupDetail {
  group_id: string;
  group_name: string;
}

export interface GroupMember {
  id: string;
  email: string;
  username: string;
  full_name: string;
}



export interface GroupDataSourcePermission {
  data_source_id: string;
  data_source_name: string;
  group_id: string;
  can_read: boolean;
  can_write: boolean;
  can_approve: boolean;
}

// API Response Types
export interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

// Health Check Types
export interface HealthStatus {
  data_source_id: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  latency_ms: number;
  last_error?: string;
  last_checked: string;
  message: string;
}

// Dashboard Stats Types
export interface DashboardStats {
  my_queries_today: number;
  pending_approvals?: number;
  total_queries?: number;
  db_access_count: number;
  total_users?: number;
}

// Schema Types
export interface DatabaseSchema {
  data_source_id: string;
  data_source_name: string;
  database_type: string;
  database_name: string;
  tables: TableInfo[];
  views?: ViewInfo[];
  functions?: FunctionInfo[];
  schemas?: string[];
}

export interface TableInfo {
  table_name: string;
  schema: string;
  table_type: 'table' | 'view';
  columns: SchemaColumnInfo[];
  indexes?: IndexInfo[];
}

export interface ViewInfo {
  view_name: string;
  schema: string;
  columns: SchemaColumnInfo[];
  definition?: string;
}

export interface FunctionInfo {
  function_name: string;
  schema: string;
  return_type?: string;
  parameters?: string;
  function_type: 'scalar' | 'aggregate' | 'window';
}

export interface SchemaColumnInfo {
  column_name: string;
  data_type: string;
  is_nullable: boolean;
  column_default?: string;
  is_primary_key: boolean;
  is_foreign_key: boolean;
}

export interface IndexInfo {
  index_name: string;
  columns: string[];
  is_unique: boolean;
  is_primary: boolean;
}

// WebSocket Types
export interface WebSocketMessage {
  type: 'connected' | 'schema' | 'schema_update' | 'subscribed' | 'error' | 'get_schema' | 'subscribe_schema' | 'subscribe_stats' | 'subscribed_stats' | 'stats_changed';
  payload?: any;
}

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export interface SchemaUpdatePayload {
  data_source_id: string;
  schema: DatabaseSchema;
}
