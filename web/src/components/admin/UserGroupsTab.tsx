'use client';

import { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import { useAuthStore } from '@/stores/auth-store';
import type { User, UserGroupDetail, Group } from '@/types';

interface UserGroupsTabProps {
  user: User;
}

export default function UserGroupsTab({ user }: UserGroupsTabProps) {
  const [userGroups, setUserGroups] = useState<UserGroupDetail[]>([]);
  const [availableGroups, setAvailableGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);
  
  const currentUser = useAuthStore((state) => state.user);
  const loadUser = useAuthStore((state) => state.loadUser);

  // Form state for adding new group
  const [selectedGroupId, setSelectedGroupId] = useState('');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [groupsData, allGroupsData] = await Promise.all([
        apiClient.getUserGroups(user.id),
        apiClient.getGroups(),
      ]);
      setUserGroups(groupsData || []);
      setAvailableGroups(allGroupsData || []);
    } catch (err) {
      toast.error(`Failed to load groups: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [user.id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleAddGroup = async () => {
    if (!selectedGroupId) return;
    if (userGroups.some((g) => g.group_id === selectedGroupId)) {
      toast.error('User is already in this group');
      return;
    }
    try {
      setSaving('add');
      const updatedGroups = [
        ...userGroups.map((g) => ({ group_id: g.group_id })),
        { group_id: selectedGroupId },
      ];
      await apiClient.assignUserGroups(user.id, updatedGroups);
      toast.success('Added to group successfully');
      setSelectedGroupId('');
      await fetchData();
      if (currentUser?.id === user.id) {
        await loadUser();
      }
    } catch (err) {
      toast.error(`Failed to add group: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setSaving(null);
    }
  };

  const handleRemoveGroup = async (groupId: string) => {
    try {
      setSaving(`remove-${groupId}`);
      const updatedGroups = userGroups
        .filter((g) => g.group_id !== groupId)
        .map((g) => ({ group_id: g.group_id }));
      await apiClient.assignUserGroups(user.id, updatedGroups);
      toast.success('Removed from group');
      await fetchData();
      if (currentUser?.id === user.id) {
        await loadUser();
      }
    } catch (err) {
      toast.error(`Failed to remove group: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setSaving(null);
    }
  };



  // Groups user is not yet a member of
  const groupsToAdd = availableGroups.filter(
    (ag) => !userGroups.some((ug) => ug.group_id === ag.id)
  );

  if (loading) {
    return (
      <div className="space-y-4 mt-6 animate-pulse">
        {[1, 2].map((n) => (
          <div key={n} className="h-14 bg-[var(--bg-hover)] rounded-lg" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-8 mt-6">
      {/* Add to Group Form */}
      <div className="bg-[var(--bg-hover)] p-6 rounded-lg border border-[var(--border)]">
        <h3 className="text-sm font-bold tracking-[0.1em] uppercase text-[var(--text-primary)] mb-4">
          Add to Group
        </h3>
        <div className="flex flex-col md:flex-row gap-4 items-end">
          <div className="flex-1 w-full">
            <label className="block text-xs font-medium text-[var(--text-muted)] mb-1">
              Select Group
            </label>
            <select
              value={selectedGroupId}
              onChange={(e) => setSelectedGroupId(e.target.value)}
              className="w-full bg-[var(--bg-page)] border border-[var(--border)] px-3 py-2 text-sm text-[var(--text-primary)] rounded focus:outline-none focus:border-[var(--accent-blue)]"
              disabled={saving !== null}
            >
              <option value="">-- Choose a group --</option>
              {groupsToAdd.map((g) => (
                <option key={g.id} value={g.id}>{g.name}</option>
              ))}
            </select>
            {groupsToAdd.length === 0 && availableGroups.length > 0 && (
              <p className="text-xs text-[var(--text-muted)] mt-1">
                User is already in all available groups.
              </p>
            )}
          </div>

          <button
            onClick={handleAddGroup}
            disabled={!selectedGroupId || saving !== null}
            className="w-full md:w-auto px-6 py-2 bg-[var(--accent-blue)] text-white text-sm font-medium rounded hover:bg-opacity-90 disabled:opacity-50 transition-colors"
          >
            {saving === 'add' ? 'Adding…' : 'Add to Group'}
          </button>
        </div>
      </div>

      {/* Current Memberships */}
      <div>
        <h3 className="text-sm font-bold tracking-[0.1em] uppercase text-[var(--text-primary)] mb-4">
          Current Memberships
          <span className="ml-2 text-xs font-normal text-[var(--text-muted)] normal-case tracking-normal">
            ({userGroups.length})
          </span>
        </h3>

        {userGroups.length === 0 ? (
          <div className="text-center py-10 text-sm text-[var(--text-muted)] border border-[var(--border)] border-dashed rounded-lg">
            This user is not a member of any groups yet.
          </div>
        ) : (
          <div className="border border-[var(--border)] rounded-lg overflow-hidden">
            <table className="w-full text-left text-sm">
              <thead className="bg-[var(--bg-hover)] bg-opacity-50 text-[var(--text-muted)] uppercase tracking-wider text-xs border-b border-[var(--border)]">
                <tr>
                  <th className="px-4 py-3 font-medium">Group</th>
                  <th className="px-4 py-3 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--border)] bg-[var(--bg-page)] text-[var(--text-primary)]">
                {userGroups.map((ug) => {
                  const isBusy = saving === `remove-${ug.group_id}`;
                  return (
                    <tr key={ug.group_id} className={`transition-colors ${isBusy ? 'opacity-60' : 'hover:bg-[var(--bg-hover)]'}`}>
                      <td className="px-4 py-3 font-medium">{ug.group_name}</td>
                      <td className="px-4 py-3 text-right">
                        <button
                          onClick={() => handleRemoveGroup(ug.group_id)}
                          disabled={saving !== null}
                          className="text-[var(--accent-red)] hover:text-red-400 font-medium px-2 py-1 rounded hover:bg-[var(--accent-red)] hover:bg-opacity-10 transition-colors text-xs uppercase tracking-wide disabled:opacity-50"
                        >
                          {saving === `remove-${ug.group_id}` ? '…' : 'Remove'}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
