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
          <div key={n} className="h-14 bg-[var(--input-bg)] rounded-xl" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-8 mt-6">
      {/* Add Member Form */}
      <div className="flex flex-col gap-3">
        <label className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Add Member
        </label>
        <div className="flex flex-col sm:flex-row gap-3 items-end">
          <div className="relative flex-1 w-full relative">
            <select
              value={selectedUserId}
              onChange={(e) => setSelectedUserId(e.target.value)}
              className="w-full bg-[var(--input-bg)] px-4 py-3 pr-10 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl cursor-pointer border border-transparent appearance-none"
              disabled={saving !== null}
            >
              <option value="" className="bg-[var(--card-bg)] text-[var(--text-primary)]">-- Choose a user --</option>
              {usersToAdd.map((u) => (
                <option key={u.id} value={u.id} className="bg-[var(--card-bg)] text-[var(--text-primary)]">
                  {u.full_name || u.username} ({u.email})
                </option>
              ))}
            </select>
            <div className="pointer-events-none absolute inset-y-0 right-4 flex items-center text-[var(--text-muted)]">
              <svg className="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"/></svg>
            </div>
            {usersToAdd.length === 0 && availableUsers.length > 0 && (
              <p className="text-xs text-[var(--text-muted)] mt-2 pl-1">
                All users are already in this group.
              </p>
            )}
          </div>

          <button
            onClick={handleAddMember}
            disabled={!selectedUserId || saving !== null}
            className="w-full sm:w-auto h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50 rounded-xl whitespace-nowrap flex-shrink-0"
          >
            {saving === 'add' ? 'Adding…' : 'Add'}
          </button>
        </div>
      </div>

      {/* Current Members List */}
      <div className="flex flex-col gap-3">
        <h3 className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1 flex items-center">
          Current Members
          <span className="ml-2 bg-[var(--input-bg)] text-[var(--text-primary)] px-2 py-0.5 rounded-full text-[10px]">
            {members.length}
          </span>
        </h3>

        {members.length === 0 ? (
          <div className="text-center py-10 text-sm text-[var(--text-muted)] border border-transparent bg-[var(--input-bg)] rounded-xl">
            No members in this group yet.
          </div>
        ) : (
          <div className="bg-[var(--input-bg)] rounded-xl overflow-hidden">
            <div className="divide-y divide-[var(--border)] divide-opacity-30">
              {members.map((m) => {
                const isBusy = saving === `remove-${m.id}`;
                return (
                  <div
                    key={m.id}
                    className={`flex items-center justify-between p-4 transition-colors ${isBusy ? 'opacity-60' : 'hover:bg-black/5 dark:hover:bg-white/5'}`}
                  >
                    <div className="flex flex-col gap-0.5">
                      <div className="font-medium text-[var(--text-primary)] text-sm">{m.full_name || m.username}</div>
                      <div className="text-xs text-[var(--text-muted)]">{m.email}</div>
                    </div>
                    <button
                      onClick={() => handleRemoveMember(m.id)}
                      disabled={saving !== null}
                      className="h-9 px-4 bg-transparent border border-[var(--border)] text-[var(--red-text)] text-[10px] font-bold tracking-[0.1em] uppercase hover:border-[var(--red-text)] hover:bg-[var(--red-bg)] transition-colors rounded-xl flex-shrink-0 disabled:opacity-50"
                    >
                      {saving === `remove-${m.id}` ? '…' : 'Remove'}
                    </button>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
