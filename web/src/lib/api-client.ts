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
  HealthStatus,
  CreateDataSourceRequest,
  DatabaseSchema,
  TableInfo,
} from '@/types';

class ApiClient {
  private client: AxiosInstance;
  private token: string | null = null;

  constructor() {
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Load token from localStorage on client side
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('token');
      if (this.token) {
        this.setAuthToken(this.token);
      }
    }

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
      (error) => {
        if (error.response?.status === 401) {
          // Don't redirect if:
          // 1. Already on login page
          // 2. The failed request was to /auth/login (legitimate failed login)
          const isLoginPage = typeof window !== 'undefined' && window.location.pathname === '/login';
          const isLoginRequest = error.config?.url?.includes('/auth/login');
          
          if (!isLoginPage && !isLoginRequest) {
            // Clear token and redirect to login
            this.clearToken();
            if (typeof window !== 'undefined') {
              // Also clear the auth store
              localStorage.removeItem('auth-storage');
              window.location.href = '/login';
            }
          } else if (!isLoginRequest) {
            // On login page but got 401 from other request, just clear token
            this.clearToken();
            if (typeof window !== 'undefined') {
              localStorage.removeItem('auth-storage');
            }
          }
        }
        return Promise.reject(error);
      }
    );
  }

  setAuthToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
    }
  }

  // Authentication
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.post<LoginResponse>('/api/v1/auth/login', credentials);
    this.setAuthToken(response.data.token);
    return response.data;
  }

  async logout(): Promise<void> {
    this.clearToken();
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

  async getQueryHistory(page = 1, limit = 20): Promise<{ queries: Query[]; total: number }> {
    const response = await this.client.get<{ history: Query[]; limit: number; page: number; total: number }>(
      `/api/v1/queries/history?page=${page}&limit=${limit}`
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

  // Approvals
  async getApprovals(params?: { status?: string; page?: number }): Promise<ApprovalRequest[]> {
    const queryParams = new URLSearchParams();
    if (params?.status) queryParams.append('status', params.status);
    if (params?.page) queryParams.append('page', params.page.toString());

    const response = await this.client.get<{ approvals: ApprovalRequest[] }>(
      `/api/v1/approvals?${queryParams.toString()}`
    );
    return response.data.approvals;
  }

  async getApproval(id: string): Promise<ApprovalRequest> {
    const response = await this.client.get<ApprovalRequest>(`/api/v1/approvals/${id}`);
    return response.data;
  }

  async reviewApproval(id: string, data: ReviewApprovalRequest): Promise<void> {
    await this.client.post(`/api/v1/approvals/${id}/review`, data);
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

  async addUserToGroup(groupId: string, userId: string): Promise<void> {
    await this.client.post(`/api/v1/groups/${groupId}/users`, { user_id: userId });
  }

  async removeUserFromGroup(groupId: string, userId: string): Promise<void> {
    await this.client.delete(`/api/v1/groups/${groupId}/users`, { data: { user_id: userId } });
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
}

// Export singleton instance
export const apiClient = new ApiClient();
