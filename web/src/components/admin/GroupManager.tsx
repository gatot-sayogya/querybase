'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
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
      toast.error(`Failed to save group: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
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
        </>
      )}

      {(view === 'create' || view === 'edit') && (
        <>
          <div className="page-header" style={{ marginBottom: '20px' }}>
            <div>
              <button
                onClick={handleCancel}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: '13px', cursor: 'pointer', padding: 0, marginBottom: '8px' }}
              >
                ← Back to Groups
              </button>
              <h1 className="page-title">
                {view === 'create' ? 'Add Group' : 'Edit Group'}
              </h1>
            </div>
          </div>
          <div className="card card-padded" style={{ maxWidth: '600px' }}>
            <GroupForm
              group={selectedGroup || undefined}
              onSave={handleSave}
              onCancel={handleCancel}
            />
          </div>
        </>
      )}
    </div>
  );
}
