'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { Group } from '@/types';
import GroupList from './GroupList';
import GroupForm from './GroupForm';

export default function GroupManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setSelectedGroup(null);
    setView('create');
  };

  const handleEditGroup = (group: Group) => {
    setSelectedGroup(group);
    setView('edit');
  };

  const handleSave = async () => {
    try {
      if (view === 'create') {
        await apiClient.createGroup({
          name: selectedGroup?.name || '',
          description: selectedGroup?.description,
        });
      } else if (view === 'edit' && selectedGroup) {
        await apiClient.updateGroup(selectedGroup.id, {
          name: selectedGroup.name,
          description: selectedGroup.description,
        });
      }

      setView('list');
      setSelectedGroup(null);
      setRefreshKey((prev) => prev + 1);
    } catch (err) {
      alert(`Failed to save group: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const handleCancel = () => {
    setView('list');
    setSelectedGroup(null);
  };

  return (
    <div className="space-y-6">
      {view === 'list' && (
        <>
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Groups</h1>
            <button
              onClick={handleCreateNew}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Add Group
            </button>
          </div>
          <GroupList key={refreshKey} onEditGroup={handleEditGroup} selectedId={null} />
        </>
      )}

      {(view === 'create' || view === 'edit') && (
        <>
          <div className="flex items-center space-x-4">
            <button
              onClick={handleCancel}
              className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
            >
              ← Back to List
            </button>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              {view === 'create' ? 'Add Group' : 'Edit Group'}
            </h1>
          </div>
          <GroupForm
            group={selectedGroup || undefined}
            onSave={handleSave}
            onCancel={handleCancel}
          />
        </>
      )}
    </div>
  );
}
