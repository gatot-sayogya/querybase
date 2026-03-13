import axios, { AxiosInstance, AxiosError } from 'axios';
import type {
  LoginRequest,
  LoginResponse,
  User,
  DataSource,
  Query,
  QueryResult,
  ExecuteQueryRequest,
  PaginatedResults,
  ApprovalRequest,
  ReviewApprovalRequest,
  Group,
  ChangePasswordRequest,
  CreateDataSourceRequest,
  DatabaseSchema,
  TableInfo,
  DashboardStats,
  HealthStatus,
  UserGroupDetail,
  GroupMember,
  GroupDataSourcePermission,
  WriteQueryPreview,
} from '@/types';

type AuthErrorHandler = () => void;

class ApiClient {
  private client: AxiosInstance;
  private token: string | null = null;
  private isRefreshing = false;
  private refreshSubscribers: ((token: string) => void)[] = [];
  private onAuthErrorCallback: AuthErrorHandler | null = null;
  private isRedirecting = false;

  constructor() {
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    this.client = axios.create({
      baseURL,
      withCredentials: true, // Crucial for HttpOnly cookies
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        if (this.token) {
          config.headers.Authorization = `Bearer ${this.token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          // If already on login page or it was a login/refresh request, give up
          const isAuthEndpoint = originalRequest.url?.includes('/auth/login') || originalRequest.url?.includes('/auth/refresh');
          const isLoginPage = typeof window !== 'undefined' && window.location.pathname === '/login';
          
          if (isAuthEndpoint || isLoginPage) {
            this.clearToken();
            return Promise.reject(error);
          }

          // Prevent multiple redirects
          if (this.isRedirecting) {
            return Promise.reject(error);
          }

          if (this.isRefreshing) {
            return new Promise((resolve) => {
              this.refreshSubscribers.push((token: string) => {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                resolve(this.client(originalRequest));
              });
            });
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            const res = await this.client.post<{ token: string }>('/api/v1/auth/refresh');
            const newToken = res.data.token;
            this.setAuthToken(newToken);
            this.isRefreshing = false;
            this.onRefreshed(newToken);
            
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            this.isRefreshing = false;
            this.clearToken();
            
            // Call the auth error handler if set (for toast + proper state clearing)
            if (this.onAuthErrorCallback && !this.isRedirecting) {
              this.isRedirecting = true;
              this.onAuthErrorCallback();
            } else if (typeof window !== 'undefined' && !this.isRedirecting) {
              // Fallback: clear storage and redirect
              this.isRedirecting = true;
              localStorage.removeItem('auth-storage');
              window.location.href = '/login?session=expired';
            }
            
            // Create a clean error without axios details for auth failures
            // This prevents console noise while still rejecting properly
            const authError = new Error('Session expired');
            (authError as any).__authHandled = true;
            return Promise.reject(authError);
          }
        }
        return Promise.reject(error);
      }
    );
  }

  setOnAuthErrorHandler(callback: AuthErrorHandler) {
    this.onAuthErrorCallback = callback;
  }

  private onRefreshed(token: string) {
    this.refreshSubscribers.forEach((cb) => cb(token));
    this.refreshSubscribers = [];
  }


  setAuthToken(token: string | null) {
    this.token = token;
  }

  clearToken() {
    this.token = null;
  }


  // Dashboard Stats
  async getDashboardStats(): Promise<DashboardStats> {
    const response = await this.client.get<DashboardStats>('/api/v1/dashboard/stats');
    return response.data;
  }

  // Authentication
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.post<LoginResponse>('/api/v1/auth/login', credentials);
    this.setAuthToken(response.data.token);
    return response.data;
  }

  async logout(): Promise<void> {
    try {
      await this.client.post('/api/v1/auth/logout');
    } catch (error) {
      console.error('Logout request failed', error);
    } finally {
      this.clearToken();
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth-storage');
      }
    }
  }


  async getCurrentUser(): Promise<User> {
    const response = await this.client.get<User>('/api/v1/auth/me');
    return response.data;
  }

  async changePassword(data: ChangePasswordRequest): Promise<void> {
    await this.client.post('/api/v1/auth/change-password', data);
  }

  // User Management (Admin Only)
  async getUsers(): Promise<User[]> {
    const response = await this.client.get<User[]>('/api/v1/auth/users');
    return response.data;
  }

  async getUser(id: string): Promise<User> {
    const response = await this.client.get<User>(`/api/v1/auth/users/${id}`);
    return response.data;
  }

  async createUser(data: {
    email: string;
    username: string;
    password: string;
    full_name: string;
    role: 'admin' | 'user' | 'viewer';
  }): Promise<User> {
    const response = await this.client.post<User>('/api/v1/auth/users', data);
    return response.data;
  }

  async updateUser(
    id: string,
    data: {
      email?: string;
      username?: string;
      full_name?: string;
      role?: 'admin' | 'user' | 'viewer';
      is_active?: boolean;
    }
  ): Promise<User> {
    const response = await this.client.put<User>(`/api/v1/auth/users/${id}`, data);
    return response.data;
  }

  async deleteUser(id: string): Promise<void> {
    await this.client.delete(`/api/v1/auth/users/${id}`);
  }

  // Data Sources
  async getDataSources(): Promise<DataSource[]> {
    const response = await this.client.get<{ data_sources: DataSource[] }>('/api/v1/datasources');
    return response.data.data_sources;
  }

  async getDataSourcesWithPermissions(): Promise<DataSource[]> {
    // Backend now includes permissions in the list response
    // No need for N+1 queries anymore
    const response = await this.client.get<{ data_sources: DataSource[] }>('/api/v1/datasources');
    return response.data.data_sources;
  }

  async getDataSource(id: string): Promise<DataSource> {
    const response = await this.client.get<DataSource>(`/api/v1/datasources/${id}`);
    return response.data;
  }

  async createDataSource(data: CreateDataSourceRequest): Promise<DataSource> {
    const response = await this.client.post<DataSource>('/api/v1/datasources', data);
    return response.data;
  }

  async updateDataSource(id: string, data: Partial<CreateDataSourceRequest>): Promise<DataSource> {
    const response = await this.client.put<DataSource>(`/api/v1/datasources/${id}`, data);
    return response.data;
  }

  async deleteDataSource(id: string): Promise<void> {
    await this.client.delete(`/api/v1/datasources/${id}`);
  }

  async testDataSourceConnection(id: string, data: Partial<CreateDataSourceRequest>): Promise<void> {
    await this.client.post(`/api/v1/datasources/${id}/test`, data);
  }

  async testNewDataSourceConnection(data: CreateDataSourceRequest): Promise<void> {
    await this.client.post('/api/v1/datasources/test', data);
  }

  async getDataSourceHealth(id: string): Promise<HealthStatus> {
    const response = await this.client.get<HealthStatus>(`/api/v1/datasources/${id}/health`);
    return response.data;
  }

  // Queries
  async executeQuery(data: ExecuteQueryRequest): Promise<Query> {
    const response = await this.client.post<Query>('/api/v1/queries', data);
    return response.data;
  }

  async getQuery(id: string): Promise<Query & { results?: QueryResult }> {
    const response = await this.client.get<Query & { results?: QueryResult }>(`/api/v1/queries/${id}`);
    return response.data;
  }

  async getQueryResults(
    id: string,
    page = 1,
    perPage = 50,
    sortColumn?: string,
    sortDirection?: 'asc' | 'desc'
  ): Promise<PaginatedResults> {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('per_page', perPage.toString());
    if (sortColumn) params.append('sort_column', sortColumn);
    if (sortDirection) params.append('sort_direction', sortDirection);

    const response = await this.client.get<PaginatedResults>(
      `/api/v1/queries/${id}/results?${params.toString()}`
    );
    return response.data;
  }

  async getQueryHistory(page = 1, limit = 20, search = ''): Promise<{ queries: Query[]; total: number }> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });
    if (search) {
      params.append('search', search);
    }
    const response = await this.client.get<{ history: Query[]; limit: number; page: number; total: number }>(
      `/api/v1/queries/history?${params.toString()}`
    );
    return {
      queries: response.data.history,
      total: response.data.total || response.data.history.length,
    };
  }

  async saveQuery(data: ExecuteQueryRequest): Promise<Query> {
    const response = await this.client.post<Query>('/api/v1/queries/save', data);
    return response.data;
  }

  async deleteQuery(id: string): Promise<void> {
    await this.client.delete(`/api/v1/queries/${id}`);
  }

  async exportQuery(queryId: string, format: 'csv' | 'json'): Promise<Blob> {
    const response = await this.client.post(
      '/api/v1/queries/export',
      { query_id: queryId, format },
      { responseType: 'blob' }
    );
    return response.data;
  }

  // Write Query Preview
  async previewWriteQuery(dataSourceId: string, queryText: string): Promise<WriteQueryPreview> {
    const response = await this.client.post<WriteQueryPreview>('/api/v1/queries/preview', {
      data_source_id: dataSourceId,
      query_text: queryText,
    });
    return response.data;
  }

  // Approvals
  async getApprovalCounts(): Promise<Record<string, number>> {
    const response = await this.client.get<Record<string, number>>('/api/v1/approvals/counts');
    return response.data;
  }

  async getApprovals(params?: { status?: string; page?: number }): Promise<ApprovalRequest[]> {
    const queryParams = new URLSearchParams();
    if (params?.status) queryParams.append('status', params.status);
    if (params?.page) queryParams.append('page', params.page.toString());

    const response = await this.client.get<{ approvals: ApprovalRequest[] }>(
      `/api/v1/approvals?${queryParams.toString()}`
    );
    return response.data.approvals;
  }

  async getApprovalHistory(page = 1, limit = 20, search = ''): Promise<{ approvals: ApprovalRequest[]; total: number }> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });
    if (search) {
      params.append('search', search); // if backend supports it, otherwise it's just ignored
    }
    const response = await this.client.get<{ approvals: ApprovalRequest[]; total: number }>(
      `/api/v1/approvals?${params.toString()}`
    );
    return {
      approvals: response.data.approvals || [],
      total: response.data.total || (response.data.approvals ? response.data.approvals.length : 0),
    };
  }

  async getApproval(id: string): Promise<ApprovalRequest> {
    const response = await this.client.get<ApprovalRequest>(`/api/v1/approvals/${id}`);
    return response.data;
  }

  async reviewApproval(id: string, data: ReviewApprovalRequest): Promise<void> {
    await this.client.post(`/api/v1/approvals/${id}/review`, data);
  }

  async startApprovalTransaction(id: string, data?: { audit_mode?: string }): Promise<any> {
    const response = await this.client.post(`/api/v1/approvals/${id}/transaction-start`, data);
    return response.data;
  }

  async commitTransaction(transactionId: string, data?: { audit_mode?: string }): Promise<any> {
    const response = await this.client.post(`/api/v1/transactions/${transactionId}/commit`, data);
    return response.data;
  }

  async rollbackTransaction(transactionId: string): Promise<any> {
    const response = await this.client.post(`/api/v1/transactions/${transactionId}/rollback`);
    return response.data;
  }

  // Groups
  async getGroups(): Promise<Group[]> {
    const response = await this.client.get<{ groups: Group[] }>('/api/v1/groups');
    return response.data.groups;
  }

  async getGroup(id: string): Promise<Group> {
    const response = await this.client.get<Group>(`/api/v1/groups/${id}`);
    return response.data;
  }

  async createGroup(data: { name: string; description?: string }): Promise<Group> {
    const response = await this.client.post<Group>('/api/v1/groups', data);
    return response.data;
  }

  async updateGroup(id: string, data: { name?: string; description?: string }): Promise<Group> {
    const response = await this.client.put<Group>(`/api/v1/groups/${id}`, data);
    return response.data;
  }

  async deleteGroup(id: string): Promise<void> {
    await this.client.delete(`/api/v1/groups/${id}`);
  }

  // --- Group Members ---
  async getGroupMembers(groupId: string): Promise<GroupMember[]> {
    const response = await this.client.get<{ users: GroupMember[] }>(`/api/v1/groups/${groupId}/members`);
    return response.data.users;
  }

  async addGroupMember(groupId: string, userId: string): Promise<void> {
    await this.client.post(`/api/v1/groups/${groupId}/members`, { user_id: userId });
  }

  async removeGroupMember(groupId: string, userId: string): Promise<void> {
    await this.client.delete(`/api/v1/groups/${groupId}/members/${userId}`);
  }

  // --- Group Data Source Permissions ---
  async getGroupDataSourcePermissions(groupId: string): Promise<GroupDataSourcePermission[]> {
    const response = await this.client.get<{ permissions: GroupDataSourcePermission[] }>(`/api/v1/groups/${groupId}/datasource_permissions`);
    return response.data.permissions;
  }

  async setGroupDataSourcePermission(groupId: string, permission: Pick<GroupDataSourcePermission, 'data_source_id' | 'can_read' | 'can_write' | 'can_approve'>): Promise<void> {
    await this.client.put(`/api/v1/groups/${groupId}/datasource_permissions`, permission);
  }

  // --- User Group Memberships ---
  async getUserGroups(userId: string): Promise<UserGroupDetail[]> {
    const response = await this.client.get<{ groups: UserGroupDetail[] }>(`/api/v1/auth/users/${userId}/groups`);
    return response.data.groups;
  }

  async assignUserGroups(userId: string, groups: { group_id: string }[]): Promise<void> {
    await this.client.put(`/api/v1/auth/users/${userId}/groups`, { groups });
  }

  // Schema Inspection
  async getDatabaseSchema(dataSourceId: string): Promise<DatabaseSchema> {
    const response = await this.client.get<{
      schema: DatabaseSchema;
      last_sync: Date | null;
      is_cached: boolean;
      is_healthy: boolean;
      data_source: {
        id: string;
        name: string;
        type: string;
      };
    }>(`/api/v1/datasources/${dataSourceId}/schema`);
    return response.data.schema;
  }

  async syncSchema(dataSourceId: string): Promise<{
    message: string;
    task_id: string;
    schema: DatabaseSchema;
    data_source?: {
      id: string;
      name: string;
      type: string;
    };
  }> {
    const response = await this.client.post<{
      message: string;
      task_id: string;
      schema: DatabaseSchema;
      data_source?: {
        id: string;
        name: string;
        type: string;
      };
    }>(`/api/v1/datasources/${dataSourceId}/sync`);
    return response.data;
  }

  async getTables(dataSourceId: string): Promise<{ tables: TableInfo[]; total: number }> {
    const response = await this.client.get<{ tables: TableInfo[]; total: number }>(
      `/api/v1/datasources/${dataSourceId}/tables`
    );
    return response.data;
  }

  async getTableDetails(dataSourceId: string, tableName: string): Promise<TableInfo> {
    const response = await this.client.get<TableInfo>(
      `/api/v1/datasources/${dataSourceId}/table?table=${tableName}`
    );
    return response.data;
  }

  async searchTables(dataSourceId: string, searchTerm: string): Promise<{ tables: TableInfo[]; total: number }> {
    const response = await this.client.get<{ tables: TableInfo[]; total: number }>(
      `/api/v1/datasources/${dataSourceId}/search?q=${encodeURIComponent(searchTerm)}`
    );
    return response.data;
  }

  // Multi-Query Operations
  async previewMultiQuery(dataSourceId: string, queryTexts: string[]): Promise<{
    statement_count: number;
    total_estimated_rows: number;
    statements: Array<{
      sequence: number;
      query_text: string;
      operation_type: string;
      estimated_rows: number;
      preview_rows?: Record<string, unknown>[];
      columns?: { name: string; type: string }[];
      error?: string;
    }>;
    requires_approval: boolean;
  }> {
    const response = await this.client.post('/api/v1/queries/multi/preview', {
      data_source_id: dataSourceId,
      query_texts: queryTexts,
    });
    return response.data;
  }

  async executeMultiQuery(
    dataSourceId: string,
    queryTexts: string[],
    name?: string,
    description?: string
  ): Promise<{
    query_id?: string;
    transaction_id?: string;
    status: string;
    is_multi_query: boolean;
    statement_count: number;
    total_affected_rows: number;
    execution_time_ms: number;
    statements: Array<{
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
    }>;
    error_message?: string;
    requires_approval: boolean;
    approval_id?: string;
  }> {
    const response = await this.client.post('/api/v1/queries/multi/execute', {
      data_source_id: dataSourceId,
      query_texts: queryTexts,
      name,
      description,
    });
    return response.data;
  }

  async getMultiQueryStatements(transactionId: string): Promise<Array<{
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
  }>> {
    const response = await this.client.get(`/api/v1/queries/multi/${transactionId}/statements`);
    return response.data;
  }

  async commitMultiQuery(transactionId: string): Promise<{
    query_id?: string;
    transaction_id?: string;
    status: string;
    is_multi_query: boolean;
    statement_count: number;
    total_affected_rows: number;
    execution_time_ms: number;
    statements: Array<{
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
    }>;
    error_message?: string;
  }> {
    const response = await this.client.post(`/api/v1/queries/multi/${transactionId}/commit`);
    return response.data;
  }

  async rollbackMultiQuery(transactionId: string): Promise<void> {
    await this.client.post(`/api/v1/queries/multi/${transactionId}/rollback`);
  }
}

// Export singleton instance
export const apiClient = new ApiClient();
