'use client';

import { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group, GroupDataSourcePermission, DataSource } from '@/types';

interface GroupDataSourcesTabProps {
  group: Group;
}

export default function GroupDataSourcesTab({ group }: GroupDataSourcesTabProps) {
  const [permissions, setPermissions] = useState<GroupDataSourcePermission[]>([]);
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  
  // Track savings by data source ID
  const [savingKeys, setSavingKeys] = useState<Set<string>>(new Set());

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [permissionsData, dsData] = await Promise.all([
        apiClient.getGroupDataSourcePermissions(group.id),
        apiClient.getDataSources(),
      ]);
      setPermissions(permissionsData || []);
      setDataSources((dsData || []).filter((ds) => ds.is_active));
    } catch (err) {
      toast.error(`Failed to load data sources: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [group.id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handlePermissionChange = async (
    dataSourceId: string,
    field: 'can_read' | 'can_write' | 'can_approve',
    value: boolean
  ) => {
    setSavingKeys((prev) => new Set(prev).add(dataSourceId));

    try {
      const existing = permissions.find((p) => p.data_source_id === dataSourceId) || {
        data_source_id: dataSourceId,
        can_read: false,
        can_write: false,
        can_approve: false,
      };

      const payload = {
        data_source_id: dataSourceId,
        can_read: existing.can_read,
        can_write: existing.can_write,
        can_approve: existing.can_approve,
        [field]: value,
      };

      await apiClient.setGroupDataSourcePermission(group.id, payload);
      toast.success('Permission updated', { duration: 2000 });

      // Refresh to ensure we have the latest truth
      const newPermissionsData = await apiClient.getGroupDataSourcePermissions(group.id);
      setPermissions(newPermissionsData || []);
    } catch (err) {
      toast.error(`Failed to update permission: ${err instanceof Error ? err.message : 'Unknown error'}`);
      await fetchData(); // revert view on failure
    } finally {
      setSavingKeys((prev) => {
        const next = new Set(prev);
        next.delete(dataSourceId);
        return next;
      });
    }
  };

  const getPermission = (dataSourceId: string) => {
    return permissions.find((p) => p.data_source_id === dataSourceId) || {
      can_read: false,
      can_write: false,
      can_approve: false,
    };
  };

  if (loading) {
    return (
      <div className="space-y-4 mt-6 animate-pulse">
        <div className="h-10 bg-[var(--input-bg)] rounded-xl" />
        <div className="h-40 bg-[var(--input-bg)] rounded-xl" />
      </div>
    );
  }

  if (dataSources.length === 0) {
    return (
      <div className="mt-6 p-8 text-center bg-[var(--input-bg)] border border-transparent rounded-xl flex flex-col items-center justify-center min-h-[140px]">
        <p className="text-[var(--text-muted)] text-sm">No active data sources found.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 mt-6">
      <div className="bg-[var(--input-bg)] p-5 rounded-xl border border-transparent">
        <div className="text-sm text-[var(--text-muted)] leading-relaxed">
          Select which data sources this group is allowed to connect to. 
          <ul className="list-disc list-inside mt-3 space-y-1">
            <li><strong className="text-[var(--text-primary)] font-medium">Read Access:</strong> Members can see this data source in the editor and run SELECTs.</li>
            <li><strong className="text-[var(--text-primary)] font-medium">Write Access:</strong> Members can run INSERT/UPDATE/DELETE commands.</li>
            <li><strong className="text-[var(--text-primary)] font-medium">Approve Access:</strong> Members can review and approve queries.</li>
          </ul>
        </div>
      </div>

      <div className="bg-[var(--input-bg)] rounded-xl overflow-hidden border border-transparent">
        <table className="w-full text-left text-sm">
          <thead className="bg-[#00000005] dark:bg-[#ffffff05] text-[var(--text-muted)] uppercase tracking-[0.1em] text-[10px] font-bold border-b border-[var(--border)] border-opacity-30">
            <tr>
              <th className="px-5 py-4 w-2/5">Data Source</th>
              <th className="px-2 py-4 text-center w-1/5">Read Access</th>
              <th className="px-2 py-4 text-center w-1/5">Write Access</th>
              <th className="px-2 py-4 text-center w-1/5">Approve Access</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[var(--border)] divide-opacity-30">
            {dataSources.map((ds) => {
              const perm = getPermission(ds.id);
              const isSaving = savingKeys.has(ds.id);

              return (
                <tr key={ds.id} className="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                  <td className="px-5 py-4">
                    <div className="font-medium text-[var(--text-primary)] text-sm">{ds.name}</div>
                    <div className="text-xs text-[var(--text-muted)] mt-1">{ds.type} &mdash; {ds.database_name}</div>
                  </td>
                  
                  {/* Read Access Toggle */}
                  <td className="px-2 py-4 text-center align-middle">
                    <div className="flex justify-center items-center h-full">
                      {isSaving ? (
                        <span className="inline-block w-4 h-4 border-2 border-[var(--text-primary)] border-t-transparent rounded-full animate-spin" />
                      ) : (
                        <input
                          type="checkbox"
                          aria-label={`Allow read for ${ds.name}`}
                          checked={perm.can_read}
                          onChange={(e) => handlePermissionChange(ds.id, 'can_read', e.target.checked)}
                          disabled={savingKeys.size > 0}
                          className="w-4 h-4 rounded border-[var(--border)] text-[var(--text-primary)] focus:ring-[var(--text-primary)] cursor-pointer disabled:opacity-50"
                        />
                      )}
                    </div>
                  </td>

                  {/* Write Access Toggle */}
                  <td className="px-2 py-4 text-center align-middle">
                    <div className="flex justify-center items-center h-full">
                      {isSaving ? (
                        <span className="inline-block w-4 h-4 border-2 border-[var(--text-primary)] border-t-transparent rounded-full animate-spin" />
                      ) : (
                        <input
                          type="checkbox"
                          aria-label={`Allow write for ${ds.name}`}
                          checked={perm.can_write}
                          onChange={(e) => handlePermissionChange(ds.id, 'can_write', e.target.checked)}
                          disabled={savingKeys.size > 0}
                          className="w-4 h-4 rounded border-[var(--border)] text-[var(--text-primary)] focus:ring-[var(--text-primary)] cursor-pointer disabled:opacity-50"
                        />
                      )}
                    </div>
                  </td>
                  
                  {/* Approve Access Toggle */}
                  <td className="px-2 py-4 text-center align-middle">
                    <div className="flex justify-center items-center h-full">
                      {isSaving ? (
                        <span className="inline-block w-4 h-4 border-2 border-[var(--text-primary)] border-t-transparent rounded-full animate-spin" />
                      ) : (
                        <input
                          type="checkbox"
                          aria-label={`Allow approve for ${ds.name}`}
                          checked={perm.can_approve}
                          onChange={(e) => handlePermissionChange(ds.id, 'can_approve', e.target.checked)}
                          disabled={savingKeys.size > 0}
                          className="w-4 h-4 rounded border-[var(--border)] text-[var(--text-primary)] focus:ring-[var(--text-primary)] cursor-pointer disabled:opacity-50"
                        />
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
