'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import Pagination from '@/components/ui/Pagination';
import Card from '@/components/ui/Card';
import type { Group } from '@/types';

interface GroupListProps {
  onEditGroup?: (group: Group) => void;
  selectedId: string | null;
}

export default function GroupList({ onEditGroup, selectedId }: GroupListProps) {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');

  // Pagination State
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    fetchGroups();
  }, []);

  // Reset to first page when search changes
  useEffect(() => {
    setCurrentPage(1);
  }, [search]);

  const fetchGroups = async () => {
    try {
      setLoading(true);
      setError(null);
      const groupsData = await apiClient.getGroups();
      setGroups(groupsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load groups');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete group "${name}"?`)) {
      return;
    }

    try {
      await apiClient.deleteGroup(id);
      setGroups(groups.filter((g) => g.id !== id));
    } catch (err) {
      toast.error(`Failed to delete group: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const colors = [
    { bg: 'rgba(16, 185, 129, 0.1)', color: '#10B981' },
    { bg: 'rgba(59, 130, 246, 0.1)', color: '#3B82F6' },
    { bg: 'rgba(139, 92, 246, 0.1)', color: '#8B5CF6' },
    { bg: 'rgba(244, 63, 94, 0.1)', color: '#F43F5E' },
    { bg: 'rgba(245, 158, 11, 0.1)', color: '#F59E0B' },
  ];
  
  const getIconColor = (name: string) => {
    let num = 0;
    for (let i = 0; i < (name || '').length; i++) {
      num += name.charCodeAt(i);
    }
    return colors[num % colors.length];
  };

  const filteredGroups = groups.filter((group) => {
    if (search) {
      const qs = search.toLowerCase();
      if (!group.name?.toLowerCase().includes(qs) && 
          !group.description?.toLowerCase().includes(qs)) {
        return false;
      }
    }
    return true;
  });

  const paginatedGroups = filteredGroups.slice((currentPage - 1) * pageSize, currentPage * pageSize);

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
          onClick={fetchGroups}
          className="mt-2 text-sm text-[var(--red-text)] underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6 flex flex-col h-full">
      {/* Search Bar */}
      <div className="relative max-w-md w-full shrink-0">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-[var(--text-muted)]">
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search groups..."
          className="w-full pl-10 pr-4 py-2.5 bg-[var(--input-bg)] border border-[var(--border)] rounded-2xl focus:ring-2 focus:ring-[var(--accent-blue)] outline-none transition-all sleek-shadow placeholder-[var(--text-faint)] text-sm font-medium"
        />
      </div>

      <Card variant="default" padding="none" className="border-none sleek-shadow flex flex-col flex-1 min-h-0 overflow-hidden">
        {filteredGroups.length === 0 ? (
          <div className="p-20 text-center flex-1 flex items-center justify-center">
            <div>
              <div className="w-16 h-16 bg-[var(--input-bg)] rounded-full flex items-center justify-center mx-auto mb-4 text-[var(--text-faint)]">
                <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v2c0 .656-.126 1.283-.356 1.857m-7.5 10.5a3 3 0 11-5.997 3.019m-6.035 3.019A3 3 0 0110 21 12.979m3 4.5c0 1.412-.656 2.675-1.718 3.014M5 21h12a2 2 0 002-2V6a2 2 0 002-2V8a2 2 0 002-2H6a2 2 0 002-2v2a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2-2v-2a2 2 0 00-2-2h12a2 2 0 002 2v2a2 2 0 002 2z" />
                </svg>
              </div>
              <h3 className="text-lg font-bold text-[var(--text-primary)]">No Groups Found</h3>
              <p className="text-[var(--text-muted)] text-sm mt-1">Adjust your search or create a new group.</p>
            </div>
          </div>
        ) : (
          <>
            {/* Scrollable Container for items */}
            <div className="flex-1 overflow-y-auto sleek-scrollbar divide-y divide-slate-100 dark:divide-slate-800/50">
              {paginatedGroups.map((group) => {
                const iconStyle = getIconColor(group.name);
                const dataSources = group.data_sources || [];
                const users = group.users || [];
                
                return (
                  <div 
                    key={group.id}
                    className={`group p-5 transition-all duration-300 flex flex-col md:flex-row md:items-center justify-between gap-4 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 ${
                      selectedId === group.id ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''
                    }`}
                  >
                    <div className="flex items-center gap-4">
                      <div 
                        className="w-14 h-14 rounded-2xl flex items-center justify-center text-xl font-bold border shadow-sm transition-transform group-hover:scale-105"
                        style={{ background: iconStyle.bg, color: iconStyle.color, borderColor: `${iconStyle.color}20` }}
                      >
                        ⛾
                      </div>
                      <div>
                        <div className="flex items-center gap-3">
                          <span className="font-bold text-[var(--text-primary)] text-lg">
                            {group.name}
                          </span>
                          <span className="text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border bg-[var(--accent-blue)]/10 text-[var(--accent-blue)] border-[var(--accent-blue)]/20">
                            {users.length} Users
                          </span>
                        </div>
                        <div className="text-sm text-[var(--text-muted)] font-medium flex items-center gap-2 mt-0.5">
                          <span>{group.description || 'No description assigned'}</span>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                        <div className="hidden lg:flex items-center gap-2">
                            {dataSources.length > 0 ? (
                                <div className="flex -space-x-2">
                                    {dataSources.slice(0, 3).map((ds, idx) => (
                                        <div key={ds.id} className="w-8 h-8 rounded-lg bg-[var(--card-bg)] border-2 border-[var(--bg-page)] flex items-center justify-center text-[10px] font-black shadow-sm tooltip text-[var(--text-primary)]" title={ds.name}>
                                            {ds.name.charAt(0).toUpperCase()}
                                        </div>
                                    ))}
                                    {dataSources.length > 3 && (
                                        <div className="w-8 h-8 rounded-lg bg-[var(--input-bg)] border-2 border-[var(--bg-page)] flex items-center justify-center text-[10px] font-black text-[var(--text-muted)] shadow-sm">
                                            +{dataSources.length - 3}
                                        </div>
                                    )}
                                </div>
                            ) : (
                                <span className="text-[10px] font-bold text-[var(--text-faint)] uppercase italic">Unlinked</span>
                            )}
                        </div>

                        <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300">
                        {onEditGroup && (
                            <button
                            onClick={() => onEditGroup(group)}
                            className="h-10 px-6 rounded-xl bg-[var(--accent-blue)]/10 text-[var(--accent-blue)] font-bold text-xs hover:bg-[var(--accent-blue)] hover:text-white transition-all shadow-sm uppercase tracking-wider"
                            >
                            Edit
                            </button>
                        )}
                        <button
                            onClick={() => handleDelete(group.id, group.name)}
                            className="h-10 px-6 rounded-xl bg-[var(--red-bg)] text-[var(--red-text)] font-bold text-xs hover:bg-[var(--red-bg)] hover:brightness-95 hover:text-[var(--red-text)] transition-all shadow-sm uppercase tracking-wider border border-[var(--red-border)]"
                        >
                            Delete
                        </button>
                        </div>
                    </div>
                  </div>
                );
              })}
            </div>
            
            {/* Fixed Pagination Controls at bottom */}
            <div className="shrink-0 border-t border-slate-100 dark:border-slate-800/50 px-4">
              <Pagination
                currentPage={currentPage}
                totalItems={filteredGroups.length}
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
