'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group } from '@/types';
import GroupList from './GroupList';
import GroupForm from './GroupForm';
import GroupMembersTab from './GroupMembersTab';
import GroupPoliciesTab from './GroupPoliciesTab';
import Modal from '../Modal';

export default function GroupManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [activeTab, setActiveTab] = useState<'details' | 'members' | 'policies'>('details');
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setSelectedGroup(null);
    setView('create');
    setActiveTab('details');
  };

  const handleEditGroup = (group: Group) => {
    setSelectedGroup(group);
    setView('edit');
    setActiveTab('details');
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
        size={view === 'edit' ? 'lg' : 'md'}
      >
        {view === 'edit' && (
          <div className="flex border-b border-[var(--border)] mb-4">
            <button
              className={`px-4 py-2 text-sm font-medium ${
                activeTab === 'details'
                  ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('details')}
            >
              Details
            </button>
            <button
              className={`px-4 py-2 text-sm font-medium ${
                activeTab === 'members'
                  ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('members')}
            >
              Members
            </button>
            <button
              className={`px-4 py-2 text-sm font-medium ${
                activeTab === 'policies'
                  ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('policies')}
            >
              Policies
            </button>
          </div>
        )}
        
        {activeTab === 'details' && (
          <GroupForm
            group={selectedGroup || undefined}
            onSave={handleSave}
            onCancel={handleCancel}
          />
        )}
        
        {activeTab === 'members' && selectedGroup && (
          <GroupMembersTab group={selectedGroup} />
        )}
        
        {activeTab === 'policies' && selectedGroup && (
          <GroupPoliciesTab group={selectedGroup} />
        )}
      </Modal>
    </div>
  );
}
