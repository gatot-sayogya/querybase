'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group } from '@/types';
import GroupList from './GroupList';
import GroupForm from './GroupForm';
import Modal from '../Modal';

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

  const handleSave = async (data: { name: string; description?: string }) => {
    try {
      if (view === 'create') {
        await apiClient.createGroup(data);
      } else if (view === 'edit' && selectedGroup) {
        await apiClient.updateGroup(selectedGroup.id, data);
      }

      setView('list');
      setSelectedGroup(null);
      setRefreshKey((prev) => prev + 1);
    } catch (err) {
      toast.error(`Failed to save group: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const handleCancel = () => {
    setView('list');
    setSelectedGroup(null);
  };

  return (
    <div className="space-y-6">
      <div className="page-header">
        <div>
          <h1 className="page-title">Groups</h1>
          <p className="page-subtitle">Manage user groups and data source permissions</p>
        </div>
        <button
          onClick={handleCreateNew}
          className="btn btn-primary"
        >
          + Add Group
        </button>
      </div>
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <GroupList key={refreshKey} onEditGroup={handleEditGroup} selectedId={null} />
      </div>

      <Modal 
        isOpen={view === 'create' || view === 'edit'} 
        onClose={handleCancel}
        title={view === 'create' ? 'Add Group' : 'Edit Group'}
      >
        <GroupForm
          group={selectedGroup || undefined}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      </Modal>
    </div>
  );
}
