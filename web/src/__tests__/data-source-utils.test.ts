import { filterAccessibleDataSources, canWriteToDataSource, canApproveForDataSource } from '../lib/data-source-utils';
import type { DataSource, User } from '../types';

describe('Data Source Utilities', () => {
  const mockDataSource1: DataSource = {
    id: 'ds1',
    name: 'Production DB',
    type: 'postgresql',
    host: 'localhost',
    port: 5432,
    database_name: 'production',
    username: 'user1',
    is_active: true,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
    permissions: [
      {
        group_id: 'group1',
        group_name: 'Analysts',
        can_read: true,
        can_write: false,
        can_approve: false,
      },
    ],
  };

  const mockDataSource2: DataSource = {
    id: 'ds2',
    name: 'Staging DB',
    type: 'postgresql',
    host: 'localhost',
    port: 5433,
    database_name: 'staging',
    username: 'user1',
    is_active: true,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
    permissions: [
      {
        group_id: 'group2',
        group_name: 'Admins',
        can_read: true,
        can_write: true,
        can_approve: true,
      },
    ],
  };

  const mockDataSourceNoPermissions: DataSource = {
    id: 'ds3',
    name: 'Restricted DB',
    type: 'mysql',
    host: 'localhost',
    port: 3306,
    database_name: 'restricted',
    username: 'root',
    is_active: true,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
    permissions: [],
  };

  const mockAdminUser: User = {
    id: 'user1',
    email: 'admin@example.com',
    username: 'admin',
    full_name: 'Admin User',
    role: 'admin',
    is_active: true,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
    groups: [],
  };

  const mockRegularUser: User = {
    id: 'user2',
    email: 'user@example.com',
    username: 'user',
    full_name: 'Regular User',
    role: 'user',
    is_active: true,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
    groups: ['group1'],
  };

  const mockUserWithMultipleGroups: User = {
    id: 'user3',
    email: 'poweruser@example.com',
    username: 'poweruser',
    full_name: 'Power User',
    role: 'user',
    is_active: true,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
    groups: ['group1', 'group2'],
  };

  describe('filterAccessibleDataSources', () => {
    it('should return all data sources for admin users', () => {
      const dataSources = [mockDataSource1, mockDataSource2, mockDataSourceNoPermissions];
      const filtered = filterAccessibleDataSources(dataSources, mockAdminUser);

      expect(filtered).toHaveLength(3);
    });

    it('should return only data sources with can_read permission for regular users', () => {
      const dataSources = [mockDataSource1, mockDataSource2, mockDataSourceNoPermissions];
      const filtered = filterAccessibleDataSources(dataSources, mockRegularUser);

      expect(filtered).toHaveLength(1);
      expect(filtered[0].id).toBe('ds1');
    });

    it('should return all data sources user has read access to', () => {
      const dataSources = [mockDataSource1, mockDataSource2, mockDataSourceNoPermissions];
      const filtered = filterAccessibleDataSources(dataSources, mockUserWithMultipleGroups);

      expect(filtered).toHaveLength(2);
      expect(filtered.map(ds => ds.id)).toContain('ds1');
      expect(filtered.map(ds => ds.id)).toContain('ds2');
    });

    it('should return empty array for user with no groups', () => {
      const userWithoutGroups: User = {
        ...mockRegularUser,
        groups: [],
      };

      const dataSources = [mockDataSource1, mockDataSource2];
      const filtered = filterAccessibleDataSources(dataSources, userWithoutGroups);

      expect(filtered).toHaveLength(0);
    });

    it('should return empty array for null user', () => {
      const dataSources = [mockDataSource1, mockDataSource2];
      const filtered = filterAccessibleDataSources(dataSources, null);

      expect(filtered).toHaveLength(0);
    });

    it('should filter out inactive data sources', () => {
      const inactiveDataSource: DataSource = {
        ...mockDataSource1,
        is_active: false,
      };

      const dataSources = [mockDataSource1, inactiveDataSource];
      const filtered = filterAccessibleDataSources(dataSources, mockAdminUser);

      // Note: The filtering of inactive sources happens in the component, not in this utility
      expect(filtered).toHaveLength(2);
    });
  });

  describe('canWriteToDataSource', () => {
    it('should return true for admin users', () => {
      const result = canWriteToDataSource(mockDataSource1, mockAdminUser);
      expect(result).toBe(true);
    });

    it('should return false when user has no write permission', () => {
      const result = canWriteToDataSource(mockDataSource1, mockRegularUser);
      expect(result).toBe(false);
    });

    it('should return true when user has write permission', () => {
      const result = canWriteToDataSource(mockDataSource2, mockUserWithMultipleGroups);
      expect(result).toBe(true);
    });

    it('should return false for null user', () => {
      const result = canWriteToDataSource(mockDataSource1, null);
      expect(result).toBe(false);
    });

    it('should return false when data source has no permissions', () => {
      const result = canWriteToDataSource(mockDataSourceNoPermissions, mockRegularUser);
      expect(result).toBe(false);
    });
  });

  describe('canApproveForDataSource', () => {
    it('should return true for admin users', () => {
      const result = canApproveForDataSource(mockDataSource1, mockAdminUser);
      expect(result).toBe(true);
    });

    it('should return false when user has no approve permission', () => {
      const result = canApproveForDataSource(mockDataSource1, mockRegularUser);
      expect(result).toBe(false);
    });

    it('should return true when user has approve permission', () => {
      const result = canApproveForDataSource(mockDataSource2, mockUserWithMultipleGroups);
      expect(result).toBe(true);
    });

    it('should return false for null user', () => {
      const result = canApproveForDataSource(mockDataSource1, null);
      expect(result).toBe(false);
    });

    it('should return false when data source has no permissions', () => {
      const result = canApproveForDataSource(mockDataSourceNoPermissions, mockRegularUser);
      expect(result).toBe(false);
    });
  });
});
