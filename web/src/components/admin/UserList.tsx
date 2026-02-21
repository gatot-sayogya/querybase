'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { User } from '@/types';

interface UserListProps {
  onEditUser?: (user: User) => void;
  selectedId: string | null;
}

export default function UserList({ onEditUser, selectedId }: UserListProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'active' | 'inactive'>('all');
  const [roleFilter, setRoleFilter] = useState<'all' | 'admin' | 'user' | 'viewer'>('all');

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      const usersData = await apiClient.getUsers();
      setUsers(usersData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string, username: string) => {
    if (!confirm(`Are you sure you want to delete user "${username}"?`)) {
      return;
    }

    try {
      await apiClient.deleteUser(id);
      setUsers(users.filter((u) => u.id !== id));
    } catch (err) {
      toast.error(`Failed to delete user: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'badge-purple';
      case 'user':
        return 'badge-blue';
      case 'viewer':
        return 'badge-slate';
      default:
        return 'badge-slate';
    }
  };

  const getStatusBadgeColor = (isActive: boolean) => {
    return isActive ? 'badge-green' : 'badge-slate text-red-700 bg-red-50';
  };

  const getUserInitials = (name: string, email: string) => {
    if (name) return name.charAt(0).toUpperCase();
    if (email) return email.charAt(0).toUpperCase();
    return '?';
  };

  const getAvatarColor = (name: string) => {
    const colors = ['#4F46E5', '#0EA5E9', '#10B981', '#F59E0B', '#8B5CF6', '#EC4899'];
    let num = 0;
    for (let i = 0; i < (name || '').length; i++) {
      num += name.charCodeAt(i);
    }
    return colors[num % colors.length];
  };

  const filteredUsers = users.filter((user) => {
    if (filter === 'active' && !user.is_active) return false;
    if (filter === 'inactive' && user.is_active) return false;
    if (roleFilter !== 'all' && user.role !== roleFilter) return false;
    return true;
  });

  if (loading) {
    return (
      <div className="p-8 space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="animate-pulse">
            <div className="h-16 bg-slate-100 rounded-lg"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 m-4">
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        <button
          onClick={fetchUsers}
          className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <>
      <style>{`
        .user-avatar-cell { display: flex; align-items: center; gap: 10px; }
        .user-avatar-sm {
          width: 34px; height: 34px; border-radius: 50%;
          display: flex; align-items: center; justify-content: center;
          font-size: 14px; font-weight: 700; color: #fff; flex-shrink: 0;
        }
      `}</style>
      
      {/* User List */}
      {filteredUsers.length === 0 ? (
        <div style={{ padding: '60px 20px', textAlign: 'center' }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '50%', background: 'var(--bg-hover)', color: 'var(--text-muted)', marginBottom: '16px' }}>
            <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          </div>
          <h3 style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text-primary)' }}>No users found</h3>
          <p style={{ marginTop: '4px', fontSize: '14px', color: 'var(--text-muted)' }}>
            {filter === 'all' && roleFilter === 'all'
              ? 'No users yet'
              : `No ${filter !== 'all' ? filter + ' ' : ''}${roleFilter !== 'all' ? roleFilter + ' ' : ''}users`}
          </p>
        </div>
      ) : (
        <table className="data-table compact">
          <thead>
            <tr>
              <th>USER</th>
              <th>EMAIL</th>
              <th>ROLE</th>
              <th>STATUS</th>
              <th style={{ textAlign: 'right' }}>ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.map((user) => (
              <tr key={user.id} className={selectedId === user.id ? 'active' : ''}>
                <td>
                  <div className="user-avatar-cell">
                    <div className="user-avatar-sm" style={{ background: getAvatarColor(user.full_name || user.email) }}>
                      {getUserInitials(user.full_name, user.email)}
                    </div>
                    <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>
                      {user.full_name || user.username}
                      {(!user.full_name && user.username !== user.email) && <span style={{ color: 'var(--text-muted)', fontWeight: 400, marginLeft: '6px' }}>@{user.username}</span>}
                    </span>
                  </div>
                </td>
                <td style={{ color: 'var(--text-muted)' }}>{user.email}</td>
                <td>
                  <span className={`badge ${getRoleBadgeColor(user.role)}`}>
                    {user.role}
                  </span>
                </td>
                <td>
                  <span className={`badge ${getStatusBadgeColor(user.is_active)}`}>
                    {user.is_active ? 'active' : 'inactive'}
                  </span>
                </td>
                <td style={{ textAlign: 'right' }}>
                  <div className="action-buttons" style={{ display: 'flex', justifyContent: 'flex-end', gap: '6px' }}>
                    {onEditUser && (
                      <button
                        onClick={() => onEditUser(user)}
                        className="btn btn-ghost btn-sm"
                        style={{ color: 'var(--accent-blue)' }}
                      >
                        Edit
                      </button>
                    )}
                    <button
                      onClick={() => handleDelete(user.id, user.username)}
                      className="btn btn-ghost btn-danger btn-sm"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  );
}
