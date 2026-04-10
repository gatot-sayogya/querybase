'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import Pagination from '@/components/ui/Pagination';
import Card from '@/components/ui/Card';
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
  const [search, setSearch] = useState('');
  
  // Pagination State
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    fetchUsers();
  }, []);

  // Reset to first page when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [filter, roleFilter, search]);

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
        return 'bg-purple-500/10 text-purple-600 border-purple-500/20';
      case 'user':
        return 'bg-blue-500/10 text-blue-600 border-blue-500/20';
      case 'viewer':
        return 'bg-slate-500/10 text-slate-500 border-slate-500/20';
      default:
        return 'bg-slate-500/10 text-slate-500 border-slate-500/20';
    }
  };

  const getStatusBadgeColor = (isActive: boolean) => {
    return isActive 
      ? 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20' 
      : 'bg-slate-500/10 text-slate-500 border-slate-500/20 opacity-60';
  };

  const getUserInitials = (name: string, email: string) => {
    if (name) return name.charAt(0).toUpperCase();
    if (email) return email.charAt(0).toUpperCase();
    return '?';
  };

  const getAvatarColor = (name: string) => {
    const colors = ['#0F766E', '#0EA5E9', '#10B981', '#F59E0B', '#14B8A6', '#EC4899'];
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
    if (search) {
      const qs = search.toLowerCase();
      if (!user.username?.toLowerCase().includes(qs) && 
          !user.full_name?.toLowerCase().includes(qs) && 
          !user.email?.toLowerCase().includes(qs)) {
        return false;
      }
    }
    return true;
  });

  const paginatedUsers = filteredUsers.slice((currentPage - 1) * pageSize, currentPage * pageSize);

  if (loading) {
    return (
      <div className="p-8 space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="animate-pulse">
            <div className="h-16 bg-[var(--input-bg)] rounded-lg"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-[var(--red-bg)] border border-[var(--red-border)] rounded-lg p-4 m-4">
        <p className="text-sm text-[var(--red-text)]">{error}</p>
        <button
          onClick={fetchUsers}
          className="mt-2 text-sm text-[var(--red-text)] underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6 flex flex-col h-full">
      {/* Search and Filters */}
      <div className="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between shrink-0">
        <div className="relative flex-1 max-w-md w-full">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-[var(--text-muted)]">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search users..."
            className="w-full pl-10 pr-4 py-2.5 bg-[var(--input-bg)] border border-[var(--border)] rounded-2xl focus:ring-2 focus:ring-[var(--accent-blue)] outline-none transition-all sleek-shadow placeholder-[var(--text-faint)] text-sm font-medium"
          />
        </div>

        <div className="flex items-center gap-2 p-1.5 glass rounded-2xl w-fit sleek-shadow">
          {(['all', 'active', 'inactive'] as const).map(f => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`px-6 py-2 text-xs font-bold rounded-xl transition-all duration-300 ${
                filter === f 
                  ? 'bg-[var(--card-bg)] text-[var(--accent-blue)] shadow-sm' 
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <Card variant="default" padding="none" className="border-none sleek-shadow flex flex-col flex-1 min-h-0 overflow-hidden">
        {filteredUsers.length === 0 ? (
          <div className="p-20 text-center flex-1 flex items-center justify-center">
            <div>
              <div className="w-16 h-16 bg-[var(--input-bg)] rounded-full flex items-center justify-center mx-auto mb-4 text-[var(--text-faint)]">
                <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              </div>
              <h3 className="text-lg font-bold text-[var(--text-primary)]">No Users Found</h3>
              <p className="text-[var(--text-muted)] text-sm mt-1">Adjust your search/filters or create new users to get started.</p>
            </div>
          </div>
        ) : (
          <>
            {/* Scrollable Container for items */}
            <div className="flex-1 overflow-y-auto sleek-scrollbar divide-y divide-slate-100 dark:divide-slate-800/50">
              {paginatedUsers.map((user) => (
                <div 
                  key={user.id}
                  className={`group p-5 transition-all duration-300 flex flex-col md:flex-row md:items-center justify-between gap-4 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 ${
                    selectedId === user.id ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''
                  }`}
                >
                  <div className="flex items-center gap-4">
                    <div 
                      className="w-14 h-14 rounded-2xl flex items-center justify-center text-xl font-black text-white shadow-lg"
                      style={{ background: `linear-gradient(135deg, ${getAvatarColor(user.full_name || user.email)}, #2563EB)` }}
                    >
                      {getUserInitials(user.full_name, user.email)}
                    </div>
                    <div>
                      <div className="flex items-center gap-3">
                        <span className="font-bold text-[var(--text-primary)] text-lg">
                          {user.full_name || user.username}
                        </span>
                        <span className={`text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border ${getRoleBadgeColor(user.role)}`}>
                          {user.role}
                        </span>
                        <span className={`text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border ${getStatusBadgeColor(user.is_active)}`}>
                          {user.is_active ? 'Active' : 'Inactive'}
                        </span>
                      </div>
                      <div className="text-sm text-[var(--text-muted)] font-medium flex items-center gap-2 mt-0.5">
                        <span>{user.email}</span>
                        {user.username !== user.email && (
                          <>
                            <span className="text-[var(--border)]">•</span>
                            <span className="opacity-60 italic">@{user.username}</span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300">
                    {onEditUser && (
                      <button
                        onClick={() => onEditUser(user)}
                        className="h-10 px-6 rounded-xl bg-[var(--accent-blue)]/10 text-[var(--accent-blue)] font-bold text-xs hover:bg-[var(--accent-blue)] hover:text-white transition-all shadow-sm uppercase tracking-wider"
                      >
                        Edit
                      </button>
                    )}
                    <button
                      onClick={() => handleDelete(user.id, user.username)}
                      className="h-10 px-6 rounded-xl bg-[var(--red-bg)] text-[var(--red-text)] font-bold text-xs hover:bg-[var(--red-bg)] hover:brightness-95 hover:text-[var(--red-text)] transition-all shadow-sm uppercase tracking-wider border border-[var(--red-border)]"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
            
            {/* Fixed Pagination Controls at bottom */}
            <div className="shrink-0 border-t border-slate-100 dark:border-slate-800/50 px-4">
              <Pagination
                currentPage={currentPage}
                totalItems={filteredUsers.length}
                pageSize={pageSize}
                onPageChange={setCurrentPage}
                onPageSizeChange={(size) => {
                  setPageSize(size);
                  setCurrentPage(1);
                }}
              />
            </div>
          </>
        )}
      </Card>
    </div>
  );
}
