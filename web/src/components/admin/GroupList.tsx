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
            { bg: '#F0FDF4', color: '#16A34A' },
            { bg: '#EFF6FF', color: '#1D4ED8' },
            { bg: '#FDF4FF', color: '#7C3AED' },
            { bg: '#FEF2F2', color: '#DC2626' },
            { bg: '#FFFBEB', color: '#D97706' },
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
    <>
      <style>{`
        .group-icon-cell { display: flex; align-items: center; gap: 10px; }
        .group-icon {
          width: 34px; height: 34px; border-radius: 9px;
          display: flex; align-items: center; justify-content: center;
          font-size: 16px; flex-shrink: 0;
        }
      `}</style>
      
      {groups.length === 0 ? (
        <div style={{ padding: '60px 20px', textAlign: 'center' }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '50%', background: 'var(--bg-hover)', color: 'var(--text-muted)', marginBottom: '16px' }}>
            <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v2c0 .656-.126 1.283-.356 1.857m-7.5 10.5a3 3 0 11-5.997 3.019m-6.035 3.019A3 3 0 0110 21 12.979m3 4.5c0 1.412-.656 2.675-1.718 3.014M5 21h12a2 2 0 002-2V6a2 2 0 00-2-2V8a2 2 0 00-2-2H6a2 2 0 00-2-2v2a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2-2v-2a2 2 0 00-2-2h12a2 2 0 002 2v2a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text-primary)' }}>No groups found</h3>
          <p style={{ marginTop: '4px', fontSize: '14px', color: 'var(--text-muted)' }}>
            Get started by creating your first group
          </p>
        </div>
      ) : (
        <table className="data-table compact">
          <thead>
            <tr>
              <th>GROUP NAME</th>
              <th>DESCRIPTION</th>
              <th>MEMBERS</th>
              <th>DATA SOURCES</th>
              <th style={{ textAlign: 'right' }}>ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((group) => {
              const iconStyle = getIconColor(group.name);
              const dataSources = group.data_sources || [];
              const users = group.users || [];
              
              return (
                <tr key={group.id} className={selectedId === group.id ? 'active' : ''}>
                  <td>
                    <div className="group-icon-cell">
                      <div className="group-icon" style={{ background: iconStyle.bg, color: iconStyle.color }}>⛾</div>
                      <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{group.name}</span>
                    </div>
                  </td>
                  <td style={{ color: 'var(--text-muted)' }}>
                    {group.description || '-'}
                  </td>
                  <td>
                    <span className="badge badge-slate">{users.length} members</span>
                  </td>
                  <td>
                    {dataSources.length === 0 && <span className="badge badge-slate" style={{ opacity: 0.5 }}>None</span>}
                    {dataSources.length > 0 && (
                      <span className="badge badge-blue">{dataSources[0].name}</span>
                    )}
                    {dataSources.length > 1 && (
                      <span className="badge badge-slate" style={{ marginLeft: '4px' }}>+{dataSources.length - 1} more</span>
                    )}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <div className="action-buttons" style={{ display: 'flex', justifyContent: 'flex-end', gap: '6px' }}>
                      {onEditGroup && (
                        <button
                          onClick={() => onEditGroup(group)}
                          className="btn btn-ghost btn-sm"
                          style={{ color: 'var(--accent-blue)' }}
                        >
                          Edit
                        </button>
                      )}
                      <button
                        onClick={() => handleDelete(group.id, group.name)}
                        className="btn btn-ghost btn-danger btn-sm"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </>
  );
}
