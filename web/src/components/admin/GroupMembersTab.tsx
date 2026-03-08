'use client';

import { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group, GroupMember, User } from '@/types';

interface GroupMembersTabProps {
  group: Group;
}

export default function GroupMembersTab({ group }: GroupMembersTabProps) {
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [availableUsers, setAvailableUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);

  // Form state
  const [selectedUserId, setSelectedUserId] = useState('');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [membersData, allUsersData] = await Promise.all([
        apiClient.getGroupMembers(group.id),
        apiClient.getUsers(),
      ]);
      setMembers(membersData || []);
      setAvailableUsers(allUsersData || []);
    } catch (err) {
      toast.error(`Failed to load members: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [group.id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleAddMember = async () => {
    if (!selectedUserId) return;
    if (members.some((m) => m.id === selectedUserId)) {
      toast.error('User is already in this group');
      return;
    }
    try {
      setSaving('add');
      await apiClient.addGroupMember(group.id, selectedUserId);
      toast.success('Member added successfully');
      setSelectedUserId('');
      await fetchData();
    } catch (err) {
      toast.error(`Failed to add member: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setSaving(null);
    }
  };

  const handleRemoveMember = async (userId: string) => {
    try {
      setSaving(`remove-${userId}`);
      await apiClient.removeGroupMember(group.id, userId);
      toast.success('Member removed');
      await fetchData();
    } catch (err) {
      toast.error(`Failed to remove member: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setSaving(null);
    }
  };



  // Filter out users already in the group AND any zero-UUID records (corrupted pre-fix)
  const usersToAdd = availableUsers.filter(
    (u) =>
      u.id !== '00000000-0000-0000-0000-000000000000' &&
      !members.some((m) => m.id === u.id)
  );

  if (loading) {
    return (
      <div className="space-y-4 mt-6 animate-pulse">
        {[1, 2, 3].map((n) => (
          <div key={n} className="h-14 bg-[var(--bg-hover)] rounded-lg" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-8 mt-6">
      {/* Add Member Form */}
      <div className="bg-[var(--bg-hover)] p-6 rounded-lg border border-[var(--border)]">
        <h3 className="text-sm font-bold tracking-[0.1em] uppercase text-[var(--text-primary)] mb-4">
          Add Member
        </h3>
        <div className="flex flex-col md:flex-row gap-4 items-end">
          <div className="flex-1 w-full">
            <label className="block text-xs font-medium text-[var(--text-muted)] mb-1">
              Select User
            </label>
            <select
              value={selectedUserId}
              onChange={(e) => setSelectedUserId(e.target.value)}
              className="w-full bg-[var(--bg-page)] border border-[var(--border)] px-3 py-2 text-sm text-[var(--text-primary)] rounded focus:outline-none focus:border-[var(--accent-blue)]"
              disabled={saving !== null}
            >
              <option value="">-- Choose a user --</option>
              {usersToAdd.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.full_name || u.username} ({u.email})
                </option>
              ))}
            </select>
            {usersToAdd.length === 0 && availableUsers.length > 0 && (
              <p className="text-xs text-[var(--text-muted)] mt-1">
                All users are already in this group.
              </p>
            )}
          </div>

          <button
            onClick={handleAddMember}
            disabled={!selectedUserId || saving !== null}
            className="w-full md:w-auto px-6 py-2 bg-[var(--accent-blue)] text-white text-sm font-medium rounded hover:bg-opacity-90 disabled:opacity-50 transition-colors"
          >
            {saving === 'add' ? 'Adding…' : 'Add Member'}
          </button>
        </div>
      </div>

      {/* Current Members List */}
      <div>
        <h3 className="text-sm font-bold tracking-[0.1em] uppercase text-[var(--text-primary)] mb-4">
          Current Members
          <span className="ml-2 text-xs font-normal text-[var(--text-muted)] normal-case tracking-normal">
            ({members.length})
          </span>
        </h3>

        {members.length === 0 ? (
          <div className="text-center py-10 text-sm text-[var(--text-muted)] border border-[var(--border)] border-dashed rounded-lg">
            No members in this group yet.
          </div>
        ) : (
          <div className="border border-[var(--border)] rounded-lg overflow-hidden">
            <table className="w-full text-left text-sm">
              <thead className="bg-[var(--bg-hover)] bg-opacity-50 text-[var(--text-muted)] uppercase tracking-wider text-xs border-b border-[var(--border)]">
                <tr>
                  <th className="px-4 py-3 font-medium">User</th>
                  <th className="px-4 py-3 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--border)] bg-[var(--bg-page)] text-[var(--text-primary)]">
                {members.map((m) => {
                  const isBusy = saving === `remove-${m.id}`;
                  return (
                    <tr
                      key={m.id}
                      className={`transition-colors ${isBusy ? 'opacity-60' : 'hover:bg-[var(--bg-hover)]'}`}
                    >
                      <td className="px-4 py-3">
                        <div className="font-medium">{m.full_name || m.username}</div>
                        <div className="text-xs text-[var(--text-muted)]">{m.email}</div>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button
                          onClick={() => handleRemoveMember(m.id)}
                          disabled={saving !== null}
                          className="text-[var(--accent-red)] hover:text-red-400 font-medium px-2 py-1 rounded hover:bg-[var(--accent-red)] hover:bg-opacity-10 transition-colors text-xs uppercase tracking-wide disabled:opacity-50"
                        >
                          {saving === `remove-${m.id}` ? '…' : 'Remove'}
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
