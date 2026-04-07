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
          <div className="flex border-b border-[var(--border)] mb-4 relative">
            <button
              className={`px-4 py-2 text-sm font-medium relative ${
                activeTab === 'details'
                  ? 'text-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('details')}
            >
              Details
            </button>
            <button
              className={`px-4 py-2 text-sm font-medium relative ${
                activeTab === 'members'
                  ? 'text-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('members')}
            >
              Members
            </button>
            <button
              className={`px-4 py-2 text-sm font-medium relative ${
                activeTab === 'data-sources'
                  ? 'text-[var(--accent-blue)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
              }`}
              onClick={() => setActiveTab('data-sources')}
            >
              Data Sources
            </button>
            <motion.div
              className="absolute bottom-0 h-0.5 bg-[var(--accent-blue)]"
              initial={false}
              animate={{
                left: activeTab === 'details' ? '0%' : activeTab === 'members' ? '33.33%' : '66.66%',
                width: '33.33%'
              }}
              transition={springConfig.smooth}
            />
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
    </motion.div>
  );
}
