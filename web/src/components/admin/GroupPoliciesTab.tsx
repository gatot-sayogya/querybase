'use client';

import { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group, GroupRolePolicy, DataSource } from '@/types';

interface GroupPoliciesTabProps {
  group: Group;
}

type PermField = 'allow_select' | 'allow_insert' | 'allow_update' | 'allow_delete';

const ROLES = ['viewer', 'member', 'analyst'] as const;
type Role = typeof ROLES[number];

const ROLE_DEFAULTS: Record<Role, Record<PermField, boolean>> = {
  viewer:  { allow_select: true,  allow_insert: false, allow_update: false, allow_delete: false },
  member:  { allow_select: true,  allow_insert: true,  allow_update: true,  allow_delete: false },
  analyst: { allow_select: true,  allow_insert: true,  allow_update: true,  allow_delete: true  },
};

const ROLE_BADGE_CLASS: Record<Role, string> = {
  viewer:  'badge badge-slate',
  member:  'badge badge-blue',
  analyst: 'badge badge-green',
};

export default function GroupPoliciesTab({ group }: GroupPoliciesTabProps) {
  const [policies, setPolicies] = useState<GroupRolePolicy[]>([]);
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  // Key: `${roleInGroup}|${dataSourceId ?? 'global'}|${field}` -> true when saving
  const [savingKeys, setSavingKeys] = useState<Set<string>>(new Set());

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [policiesData, dsData] = await Promise.all([
        apiClient.getGroupPolicies(group.id),
        apiClient.getDataSources(),
      ]);
      setPolicies(policiesData || []);
      setDataSources((dsData || []).filter((ds) => ds.is_active));
    } catch (err) {
      toast.error(`Failed to load policies: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [group.id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handlePolicyChange = async (
    roleInGroup: Role,
    dataSourceId: string | null,
    field: PermField,
    value: boolean
  ) => {
    const key = `${roleInGroup}|${dataSourceId ?? 'global'}|${field}`;
    setSavingKeys((prev) => new Set(prev).add(key));

    try {
      const existing = policies.find(
        (p) => p.role_in_group === roleInGroup && p.data_source_id === dataSourceId
      );

      const payload: Omit<GroupRolePolicy, 'id' | 'group_id'> = {
        role_in_group: roleInGroup,
        data_source_id: dataSourceId,
        allow_select: existing?.allow_select ?? ROLE_DEFAULTS[roleInGroup].allow_select,
        allow_insert: existing?.allow_insert ?? ROLE_DEFAULTS[roleInGroup].allow_insert,
        allow_update: existing?.allow_update ?? ROLE_DEFAULTS[roleInGroup].allow_update,
        allow_delete: existing?.allow_delete ?? ROLE_DEFAULTS[roleInGroup].allow_delete,
        [field]: value,
      };

      await apiClient.setGroupPolicy(group.id, payload);
      toast.success('Policy updated', { duration: 2000 });

      // Refresh from server to get generated IDs
      const newPoliciesData = await apiClient.getGroupPolicies(group.id);
      setPolicies(newPoliciesData || []);
    } catch (err) {
      toast.error(`Failed to update policy: ${err instanceof Error ? err.message : 'Unknown error'}`);
      await fetchData(); // revert
    } finally {
      setSavingKeys((prev) => {
        const next = new Set(prev);
        next.delete(key);
        return next;
      });
    }
  };

  const getPolicy = (role: Role, dataSourceId: string | null) => {
    const existing = policies.find(
      (p) => p.role_in_group === role && p.data_source_id === dataSourceId
    );
    return existing ?? { ...ROLE_DEFAULTS[role], data_source_id: dataSourceId, role_in_group: role };
  };

  if (loading) {
    return (
      <div className="space-y-4 mt-6 animate-pulse">
        <div className="h-10 bg-[var(--bg-hover)] rounded-lg" />
        <div className="h-40 bg-[var(--bg-hover)] rounded-lg" />
      </div>
    );
  }

  const renderPolicyMatrix = (dataSourceId: string | null, title: string, subtitle?: string) => (
    <div className="mb-8 last:mb-0" key={dataSourceId ?? 'global'}>
      <div className="mb-4">
        <h3 className="text-sm font-bold tracking-[0.1em] uppercase text-[var(--text-primary)]">{title}</h3>
        {subtitle && <p className="text-xs text-[var(--text-muted)] mt-1">{subtitle}</p>}
      </div>

      <div className="border border-[var(--border)] rounded-lg overflow-hidden bg-[var(--bg-page)]">
        <table className="w-full text-left text-sm">
          <thead className="bg-[var(--bg-hover)] bg-opacity-50 text-[var(--text-muted)] uppercase tracking-wider text-xs border-b border-[var(--border)]">
            <tr>
              <th className="px-4 py-3 font-medium w-1/5">Role</th>
              {(['allow_select', 'allow_insert', 'allow_update', 'allow_delete'] as PermField[]).map((f) => (
                <th key={f} className="px-4 py-3 font-medium text-center w-1/5">
                  {f.replace('allow_', '').toUpperCase()}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-[var(--border)]">
            {ROLES.map((role) => {
              const policy = getPolicy(role, dataSourceId);
              return (
                <tr key={role} className="hover:bg-[var(--bg-hover)] transition-colors">
                  <td className="px-4 py-3">
                    <span className={ROLE_BADGE_CLASS[role]}>{role}</span>
                  </td>
                  {(['allow_select', 'allow_insert', 'allow_update', 'allow_delete'] as PermField[]).map((field) => {
                    const saveKey = `${role}|${dataSourceId ?? 'global'}|${field}`;
                    const isSaving = savingKeys.has(saveKey);
                    return (
                      <td key={field} className="px-4 py-3 text-center">
                        {isSaving ? (
                          <span className="inline-block w-4 h-4 border-2 border-[var(--accent-blue)] border-t-transparent rounded-full animate-spin" />
                        ) : (
                          <input
                            type="checkbox"
                            checked={(policy as Record<string, unknown>)[field] as boolean}
                            onChange={(e) => handlePolicyChange(role, dataSourceId, field, e.target.checked)}
                            disabled={savingKeys.size > 0}
                            className="w-4 h-4 accent-[var(--accent-blue)] cursor-pointer disabled:opacity-50"
                          />
                        )}
                      </td>
                    );
                  })}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );

  return (
    <div className="space-y-6 mt-6">
      <div className="bg-[var(--bg-hover)] p-4 rounded-lg border border-[var(--border)]">
        <p className="text-sm text-[var(--text-muted)] leading-relaxed">
          Policies define what SQL operations each role can perform.{' '}
          <strong>Global Policies</strong> apply to all data sources by default.
          You can override with a per-data-source policy below.
        </p>
      </div>

      {renderPolicyMatrix(null, 'Global Policies', 'Default for any data source without a specific override')}

      {dataSources.length > 0 && (
        <div className="pt-6 mt-8 border-t border-[var(--border)]">
          <h2 className="text-base font-semibold text-[var(--text-primary)] mb-6">
            Data Source Overrides
          </h2>
          {dataSources.map((ds) =>
            renderPolicyMatrix(ds.id, ds.name, `Override for ${ds.type} — ${ds.database_name}`)
          )}
        </div>
      )}
    </div>
  );
}
