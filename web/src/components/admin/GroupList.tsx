'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group } from '@/types';

interface GroupListProps {
  onEditGroup?: (group: Group) => void;
  selectedId: string | null;
}

export default function GroupList({ onEditGroup, selectedId }: GroupListProps) {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchGroups();
  }, []);

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
          onClick={fetchGroups}
          className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-3">
        {groups.length === 0 ? (
          <div className="p-20 text-center glass rounded-3xl border border-slate-100 dark:border-slate-800/50">
            <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4 text-slate-400">
              <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v2c0 .656-.126 1.283-.356 1.857m-7.5 10.5a3 3 0 11-5.997 3.019m-6.035 3.019A3 3 0 0110 21 12.979m3 4.5c0 1.412-.656 2.675-1.718 3.014M5 21h12a2 2 0 002-2V6a2 2 0 002-2V8a2 2 0 002-2H6a2 2 0 002-2v2a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2-2v-2a2 2 0 00-2-2h12a2 2 0 002 2v2a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-bold text-slate-800 dark:text-white">No Groups Detected</h3>
            <p className="text-slate-500 text-sm mt-1">Start by defining access clusters for your team.</p>
          </div>
        ) : (
          groups.map((group) => {
            const iconStyle = getIconColor(group.name);
            const dataSources = group.data_sources || [];
            const users = group.users || [];
            
            return (
              <div 
                key={group.id}
                className="group p-5 glass rounded-3xl border border-white/50 dark:border-slate-800/50 hover:border-blue-500/30 hover:bg-white dark:hover:bg-slate-800/50 transition-all duration-300 sleek-shadow flex flex-col md:flex-row md:items-center justify-between gap-4"
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
                      <span className="font-bold text-slate-900 dark:text-white text-lg">
                        {group.name}
                      </span>
                      <span className="text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border bg-blue-500/10 text-blue-600 border-blue-500/20">
                        {users.length} Nodes
                      </span>
                    </div>
                    <div className="text-sm text-slate-500 dark:text-slate-400 font-medium flex items-center gap-2 mt-0.5">
                      <span>{group.description || 'No description assigned'}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                    <div className="hidden lg:flex items-center gap-2">
                        {dataSources.length > 0 ? (
                            <div className="flex -space-x-2">
                                {dataSources.slice(0, 3).map((ds, idx) => (
                                    <div key={ds.id} className="w-8 h-8 rounded-lg bg-white dark:bg-slate-700 border-2 border-slate-50 dark:border-slate-900 flex items-center justify-center text-[10px] font-black shadow-sm tooltip" title={ds.name}>
                                        {ds.name.charAt(0).toUpperCase()}
                                    </div>
                                ))}
                                {dataSources.length > 3 && (
                                    <div className="w-8 h-8 rounded-lg bg-slate-100 dark:bg-slate-800 border-2 border-slate-50 dark:border-slate-900 flex items-center justify-center text-[10px] font-black text-slate-400 shadow-sm">
                                        +{dataSources.length - 3}
                                    </div>
                                )}
                            </div>
                        ) : (
                            <span className="text-[10px] font-bold text-slate-300 uppercase italic">Unlinked</span>
                        )}
                    </div>

                    <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300">
                    {onEditGroup && (
                        <button
                        onClick={() => onEditGroup(group)}
                        className="h-10 px-6 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 font-bold text-xs hover:bg-blue-500 hover:text-white transition-all shadow-sm"
                        >
                        Configure
                        </button>
                    )}
                    <button
                        onClick={() => handleDelete(group.id, group.name)}
                        className="h-10 px-6 rounded-xl bg-rose-500/10 text-rose-600 dark:text-rose-400 font-bold text-xs hover:bg-rose-500 hover:text-white transition-all shadow-sm"
                    >
                        Purge
                    </button>
                    </div>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
