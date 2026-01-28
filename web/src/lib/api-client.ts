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
          // Clear token and redirect to login
          this.clearToken();
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
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

  // Data Sources
  async getDataSources(): Promise<DataSource[]> {
    const response = await this.client.get<DataSource[]>('/api/v1/datasources');
    return response.data;
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
    const response = await this.client.get<{ queries: Query[]; total: number }>(
      `/api/v1/queries/history?page=${page}&limit=${limit}`
    );
    return response.data;
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

    const response = await this.client.get<ApprovalRequest[]>(
      `/api/v1/approvals?${queryParams.toString()}`
    );
    return response.data;
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
    const response = await this.client.get<Group[]>('/api/v1/groups');
    return response.data;
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
}

// Export singleton instance
export const apiClient = new ApiClient();
