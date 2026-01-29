import type { DataSource, User } from '@/types';

/**
 * Filter data sources based on user's group permissions
 * @param dataSources - List of all data sources
 * @param user - Current user
 * @returns List of data sources the user has access to
 */
export function filterAccessibleDataSources(
  dataSources: DataSource[],
  user: User | null
): DataSource[] {
  // Admins can access all data sources
  if (user?.role === 'admin') {
    return dataSources;
  }

  // If no user or no groups, return empty array
  if (!user || !user.groups || user.groups.length === 0) {
    return [];
  }

  const userGroupIds = new Set(user.groups);

  return dataSources.filter((ds) => {
    // If no permissions defined, deny access
    if (!ds.permissions || ds.permissions.length === 0) {
      return false;
    }

    // Check if user has can_read permission for any of their groups
    return ds.permissions.some(
      (perm) => userGroupIds.has(perm.group_id) && perm.can_read
    );
  });
}

/**
 * Check if user can write to a specific data source
 * @param dataSource - Data source to check
 * @param user - Current user
 * @returns true if user can write to the data source
 */
export function canWriteToDataSource(
  dataSource: DataSource,
  user: User | null
): boolean {
  // Admins can write to all data sources
  if (user?.role === 'admin') {
    return true;
  }

  // If no user or no groups, deny access
  if (!user || !user.groups || user.groups.length === 0) {
    return false;
  }

  const userGroupIds = new Set(user.groups);

  // Check if user has can_write permission for any of their groups
  return dataSource.permissions?.some(
    (perm) => userGroupIds.has(perm.group_id) && perm.can_write
  ) || false;
}

/**
 * Check if user can approve queries for a specific data source
 * @param dataSource - Data source to check
 * @param user - Current user
 * @returns true if user can approve queries for the data source
 */
export function canApproveForDataSource(
  dataSource: DataSource,
  user: User | null
): boolean {
  // Admins can approve for all data sources
  if (user?.role === 'admin') {
    return true;
  }

  // If no user or no groups, deny access
  if (!user || !user.groups || user.groups.length === 0) {
    return false;
  }

  const userGroupIds = new Set(user.groups);

  // Check if user has can_approve permission for any of their groups
  return dataSource.permissions?.some(
    (perm) => userGroupIds.has(perm.group_id) && perm.can_approve
  ) || false;
}
