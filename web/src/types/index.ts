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
  query_text: string;
  name?: string;
  description?: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  created_at: string;
  executed_at?: string;
  row_count?: number;
  error_message?: string;
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
export interface ApprovalRequest {
  id: string;
  query_id: string;
  requester_id: string;
  status: 'pending' | 'approved' | 'rejected';
  operation_type: string;
  query_text: string;
  data_source_id: string;
  created_at: string;
  updated_at: string;
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

// Group Types
export interface Group {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  users?: User[];
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
