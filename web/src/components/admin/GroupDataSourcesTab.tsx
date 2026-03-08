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
        <div className="h-10 bg-[var(--bg-hover)] rounded-lg" />
        <div className="h-40 bg-[var(--bg-hover)] rounded-lg" />
      </div>
    );
  }

  if (dataSources.length === 0) {
    return (
      <div className="mt-6 p-8 text-center bg-[var(--bg-page)] border border-[var(--border)] rounded-lg">
        <p className="text-[var(--text-muted)]">No active data sources found.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 mt-6">
      <div className="bg-[var(--bg-hover)] p-4 rounded-lg border border-[var(--border)]">
        <div className="text-sm text-[var(--text-muted)] leading-relaxed">
          Select which data sources this group is allowed to connect to. 
          <ul className="list-disc list-inside mt-2 space-y-1">
            <li><strong>Read Access:</strong> Members can see this data source in the editor and run SELECTs.</li>
            <li><strong>Write Access:</strong> Members can run INSERT/UPDATE/DELETE commands (subject to Role Policies).</li>
            <li><strong>Approve Access:</strong> Members can review and approve queries for this data source.</li>
          </ul>
        </div>
      </div>

      <div className="border border-[var(--border)] rounded-lg overflow-hidden bg-[var(--bg-page)]">
        <table className="w-full text-left text-sm">
          <thead className="bg-[var(--bg-hover)] bg-opacity-50 text-[var(--text-muted)] uppercase tracking-wider text-xs border-b border-[var(--border)]">
            <tr>
              <th className="px-4 py-3 font-medium w-2/5">Data Source</th>
              <th className="px-4 py-3 font-medium text-center w-1/5">Read Access</th>
              <th className="px-4 py-3 font-medium text-center w-1/5">Write Access</th>
              <th className="px-4 py-3 font-medium text-center w-1/5">Approve Access</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[var(--border)]">
            {dataSources.map((ds) => {
              const perm = getPermission(ds.id);
              const isSaving = savingKeys.has(ds.id);

              return (
                <tr key={ds.id} className="hover:bg-[var(--bg-hover)] transition-colors">
                  <td className="px-4 py-3">
                    <div className="font-medium text-[var(--text-primary)]">{ds.name}</div>
                    <div className="text-xs text-[var(--text-muted)] mt-0.5">{ds.type} &mdash; {ds.database_name}</div>
                  </td>
                  
                  {/* Read Access Toggle */}
                  <td className="px-4 py-3 text-center">
                    {isSaving ? (
                      <span className="inline-block w-4 h-4 border-2 border-[var(--accent-blue)] border-t-transparent rounded-full animate-spin" />
                    ) : (
                      <input
                        type="checkbox"
                        aria-label={`Allow read for ${ds.name}`}
                        checked={perm.can_read}
                        onChange={(e) => handlePermissionChange(ds.id, 'can_read', e.target.checked)}
                        disabled={savingKeys.size > 0}
                        className="w-4 h-4 accent-[var(--accent-blue)] cursor-pointer disabled:opacity-50"
                      />
                    )}
                  </td>

                  {/* Write Access Toggle */}
                  <td className="px-4 py-3 text-center">
                    {isSaving ? (
                      <span className="inline-block w-4 h-4 border-2 border-[var(--accent-blue)] border-t-transparent rounded-full animate-spin" />
                    ) : (
                      <input
                        type="checkbox"
                        aria-label={`Allow write for ${ds.name}`}
                        checked={perm.can_write}
                        onChange={(e) => handlePermissionChange(ds.id, 'can_write', e.target.checked)}
                        disabled={savingKeys.size > 0}
                        className="w-4 h-4 accent-[var(--accent-blue)] cursor-pointer disabled:opacity-50"
                      />
                    )}
                  </td>
                  
                  {/* Approve Access Toggle */}
                  <td className="px-4 py-3 text-center">
                    {isSaving ? (
                      <span className="inline-block w-4 h-4 border-2 border-[var(--accent-blue)] border-t-transparent rounded-full animate-spin" />
                    ) : (
                      <input
                        type="checkbox"
                        aria-label={`Allow approve for ${ds.name}`}
                        checked={perm.can_approve}
                        onChange={(e) => handlePermissionChange(ds.id, 'can_approve', e.target.checked)}
                        disabled={savingKeys.size > 0}
                        className="w-4 h-4 accent-[var(--accent-blue)] cursor-pointer disabled:opacity-50"
                      />
                    )}
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
