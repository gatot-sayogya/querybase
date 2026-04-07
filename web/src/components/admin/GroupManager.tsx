'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { motion } from 'framer-motion';
import { apiClient } from '@/lib/api-client';
import type { Group } from '@/types';
import GroupList from './GroupList';
import GroupForm from './GroupForm';
import GroupMembersTab from './GroupMembersTab';
import GroupDataSourcesTab from './GroupDataSourcesTab';
import Modal from '../Modal';
import { springConfig, staggerContainer, staggerItem } from '@/lib/animations';

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
    <motion.div 
      className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      {/* Header */}
      <motion.div 
        className="flex flex-col md:flex-row md:items-center justify-between gap-6 pt-4"
        variants={staggerItem}
      >
        <div className="space-y-1">
          <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
            Groups
          </h1>
          <p className="text-slate-500 dark:text-slate-400 font-medium">
            Create and manage user groups to organize access.
          </p>
        </div>
        
        <motion.div 
          className="flex items-center gap-4"
          variants={staggerItem}
        >
          <motion.button
            onClick={handleCreateNew}
            className="btn btn-primary h-11 px-8 rounded-2xl text-sm font-bold sleek-shadow"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            transition={springConfig.micro}
          >
            <span className="text-xl mr-2">+</span>
            Create Group
          </motion.button>
        </motion.div>
      </motion.div>

      {/* Content */}
      <motion.div 
        className="space-y-6"
        variants={staggerItem}
      >
        <GroupList key={refreshKey} onEditGroup={handleEditGroup} selectedId={null} />
      </motion.div>

      <Modal 
        isOpen={view === 'create' || view === 'edit'} 
        onClose={handleCancel}
        title={view === 'create' ? 'Add Group' : 'Edit Group'}
        size={view === 'edit' ? 'lg' : 'md'}
      >
        {view === 'edit' && (
          <div className="flex bg-[var(--input-bg)] p-1 rounded-xl mb-6 relative overflow-hidden">
            <motion.div
              className="absolute top-1 bottom-1 bg-[var(--card-bg)] rounded-lg shadow-sm"
              initial={false}
              animate={{
                left: activeTab === 'details' ? '4px' : activeTab === 'members' ? '33.33%' : '66.66%',
                width: 'calc(33.33% - 5px)'
              }}
              transition={springConfig.smooth}
            />
            <button
              className={`flex-1 py-2 text-xs font-bold tracking-[0.1em] uppercase relative z-10 transition-colors ${
                activeTab === 'details'
                  ? 'text-[var(--text-primary)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('details')}
            >
              Details
            </button>
            <button
              className={`flex-1 py-2 text-xs font-bold tracking-[0.1em] uppercase relative z-10 transition-colors ${
                activeTab === 'members'
                  ? 'text-[var(--text-primary)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('members')}
            >
              Members
            </button>
            <button
              className={`flex-1 py-2 text-xs font-bold tracking-[0.1em] uppercase relative z-10 transition-colors ${
                activeTab === 'data-sources'
                  ? 'text-[var(--text-primary)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('data-sources')}
            >
              Data Sources
            </button>
          </div>
        )}
        
        {/* Persistent Form - Hidden when not active tab so global Update button can target it */}
        <div className={activeTab === 'details' ? 'block' : 'hidden'}>
          <GroupForm
            formId="group-form"
            group={selectedGroup || undefined}
            onSave={handleSave}
            onCancel={handleCancel}
            hideActions={true}
          />
        </div>
        
        {activeTab === 'members' && selectedGroup && (
          <GroupMembersTab group={selectedGroup} />
        )}

        {activeTab === 'data-sources' && selectedGroup && (
          <GroupDataSourcesTab group={selectedGroup} />
        )}

        {/* Global Footer Actions */}
        <div className="mt-8 pt-6 border-t border-[var(--border-light)] flex justify-end gap-3 w-full">
          <button
            type="button"
            onClick={handleCancel}
            className="h-12 px-6 bg-[var(--input-bg)] text-[var(--text-primary)] text-sm font-bold tracking-[0.1em] uppercase hover:bg-[var(--border)] transition-colors rounded-xl"
          >
            Cancel
          </button>
          <button
            form="group-form"
            type="submit"
            className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50 rounded-xl"
          >
            {view === 'edit' ? 'Update' : 'Save'}
          </button>
        </div>
      </Modal>
    </motion.div>
  );
}
