'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Group } from '@/types';
import GroupList from './GroupList';
import GroupForm from './GroupForm';
import GroupMembersTab from './GroupMembersTab';
import GroupDataSourcesTab from './GroupDataSourcesTab';
import Modal from '../Modal';

export default function GroupManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [activeTab, setActiveTab] = useState<'details' | 'members' | 'data-sources'>('details');
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
    <div className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 pt-4">
        <div className="space-y-1">
          <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
            Permission Groups
          </h1>
          <p className="text-slate-500 dark:text-slate-400 font-medium">
            Define access clusters and map data source policies to user roles.
          </p>
        </div>
        
        <div className="flex items-center gap-4">
          <button
            onClick={handleCreateNew}
            className="btn btn-primary h-11 px-8 rounded-2xl text-sm font-bold sleek-shadow"
          >
            <span className="text-xl mr-2">+</span>
            New Ensemble
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="space-y-6">
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
                activeTab === 'data-sources'
                  ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('data-sources')}
            >
              Data Sources
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

        {activeTab === 'data-sources' && selectedGroup && (
          <GroupDataSourcesTab group={selectedGroup} />
        )}
      </Modal>
    </div>
  );
}
