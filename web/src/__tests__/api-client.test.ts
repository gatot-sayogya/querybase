import { apiClient } from '@/lib/api-client';

// Mock axios
jest.mock('axios', () => ({
  create: () => ({
    interceptors: {
      request: { use: jest.fn() },
      response: { use: jest.fn() },
    },
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  }),
}));

describe('ApiClient', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should have login method', () => {
    expect(typeof apiClient.login).toBe('function');
  });

  it('should have logout method', () => {
    expect(typeof apiClient.logout).toBe('function');
  });

  it('should have getCurrentUser method', () => {
    expect(typeof apiClient.getCurrentUser).toBe('function');
  });

  it('should have getDataSources method', () => {
    expect(typeof apiClient.getDataSources).toBe('function');
  });

  it('should have executeQuery method', () => {
    expect(typeof apiClient.executeQuery).toBe('function');
  });

  it('should have getQuery method', () => {
    expect(typeof apiClient.getQuery).toBe('function');
  });
});
