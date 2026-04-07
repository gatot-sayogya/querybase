import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import { useAuthStore } from '@/stores/auth-store';
import { springConfig } from '@/lib/animations';
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

  // Search state
  const [searchQuery, setSearchQuery] = useState('');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
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
      setSearchQuery('');
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

  const filteredGroups = groupsToAdd.filter((g) => 
    g.name.toLowerCase().includes(searchQuery.toLowerCase())
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
      {/* Add Group Form */}
      <div className="flex flex-col gap-3">
        <label className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Add to Group
        </label>
        <div className="flex flex-col sm:flex-row gap-3 items-start relative">
          <div className="relative flex-1 w-full relative">
            <div className="relative">
              <input
                type="text"
                placeholder="Search group name..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setIsDropdownOpen(true);
                }}
                onFocus={() => setIsDropdownOpen(true)}
                className="w-full bg-[var(--input-bg)] px-4 py-3 pr-10 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent"
                disabled={saving !== null}
              />
              <div className="absolute inset-y-0 right-4 flex items-center pointer-events-none text-[var(--text-muted)]">
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
            </div>

            {/* Dropdown Results */}
            <AnimatePresence>
              {isDropdownOpen && (searchQuery.length > 0 || filteredGroups.length > 0) && (
                <>
                  <div 
                    className="fixed inset-0 z-10" 
                    onClick={() => setIsDropdownOpen(false)} 
                  />
                  <motion.div
                    initial={{ opacity: 0, y: -10, scale: 0.95 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: -10, scale: 0.95 }}
                    transition={springConfig.micro}
                    className="absolute z-20 top-full left-0 right-0 mt-2 bg-[var(--card-bg)] border border-[var(--border)] rounded-xl shadow-2xl overflow-hidden max-h-60 overflow-y-auto sleek-scrollbar"
                  >
                    {filteredGroups.length > 0 ? (
                      <div className="py-1">
                        {filteredGroups.map((g) => (
                          <button
                            key={g.id}
                            type="button"
                            onClick={() => {
                              setSelectedGroupId(g.id);
                              setSearchQuery(g.name);
                              setIsDropdownOpen(false);
                            }}
                            className={`w-full text-left px-4 py-3 hover:bg-[var(--bg-hover)] transition-colors ${selectedGroupId === g.id ? 'bg-[var(--bg-hover)]' : ''}`}
                          >
                            <span className="font-medium text-[var(--text-primary)] text-base">
                              {g.name}
                            </span>
                          </button>
                        ))}
                      </div>
                    ) : (
                      <div className="px-4 py-6 text-center text-sm text-[var(--text-muted)]">
                        No groups found matching &quot;{searchQuery}&quot;
                      </div>
                    )}
                  </motion.div>
                </>
              )}
            </AnimatePresence>

            {groupsToAdd.length === 0 && availableGroups.length > 0 && (
              <p className="text-xs text-[var(--text-muted)] mt-2 pl-1">
                User is already in all available groups.
              </p>
            )}
          </div>

          <button
            onClick={handleAddGroup}
            disabled={!selectedGroupId || saving !== null}
            className="w-full sm:w-auto h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50 rounded-xl whitespace-nowrap flex-shrink-0"
          >
            {saving === 'add' ? 'Adding…' : 'Add to Group'}
          </button>
        </div>
      </div>

      {/* Current Memberships */}
      <div className="flex flex-col gap-3">
        <h3 className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1 flex items-center">
          Current Memberships
          <span className="ml-2 bg-[var(--input-bg)] text-[var(--text-primary)] px-2 py-0.5 rounded-full text-[10px]">
            {userGroups.length}
          </span>
        </h3>

        {userGroups.length === 0 ? (
          <div className="text-center py-10 text-sm text-[var(--text-muted)] bg-[var(--input-bg)] rounded-xl">
            This user is not a member of any groups yet.
          </div>
        ) : (
          <div className="bg-[var(--input-bg)] rounded-xl overflow-hidden">
            <div className="divide-y divide-[var(--border)] divide-opacity-30">
              {userGroups.map((ug) => {
                const isBusy = saving === `remove-${ug.group_id}`;
                return (
                  <div
                    key={ug.group_id}
                    className={`flex items-center justify-between p-4 transition-colors ${isBusy ? 'opacity-60' : 'hover:bg-black/5 dark:hover:bg-white/5'}`}
                  >
                    <div className="font-medium text-[var(--text-primary)] text-sm">{ug.group_name}</div>
                    <button
                      onClick={() => handleRemoveGroup(ug.group_id)}
                      disabled={saving !== null}
                      className="h-9 px-4 bg-transparent border border-[var(--border)] text-[var(--red-text)] text-[10px] font-bold tracking-[0.1em] uppercase hover:border-[var(--red-text)] hover:bg-[var(--red-bg)] transition-colors rounded-xl flex-shrink-0 disabled:opacity-50"
                    >
                      {saving === `remove-${ug.group_id}` ? '…' : 'Remove'}
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
