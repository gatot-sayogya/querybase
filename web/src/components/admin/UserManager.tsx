'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { User } from '@/types';
import UserList from './UserList';
import UserForm from './UserForm';
import UserGroupsTab from './UserGroupsTab';
import Modal from '@/components/ui/Modal';
import { motion, AnimatePresence } from 'framer-motion';
import { fadeIn, slideUp, springConfig, duration } from '@/lib/animations';
import { PlusIcon } from '@heroicons/react/24/outline';

export default function UserManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [activeTab, setActiveTab] = useState<'details' | 'groups'>('details');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setSelectedUser(null);
    setView('create');
    setActiveTab('details');
  };

  const handleEditUser = (user: User) => {
    setSelectedUser(user);
    setView('edit');
    setActiveTab('details');
  };

  const handleSave = async (data: {
    email: string;
    username: string;
    password: string;
    full_name: string;
    role: 'admin' | 'user' | 'viewer';
    is_active: boolean;
  }) => {
    try {
      if (view === 'create') {
        await apiClient.createUser({
          email: data.email,
          username: data.username,
          password: data.password,
          full_name: data.full_name,
          role: data.role,
        });
        toast.success('User created successfully');
      } else if (view === 'edit' && selectedUser) {
        await apiClient.updateUser(selectedUser.id, {
          email: data.email,
          username: data.username,
          full_name: data.full_name,
          role: data.role,
          is_active: data.is_active,
        });
        toast.success('User updated successfully');
      }

      setView('list');
      setSelectedUser(null);
      setRefreshKey((prev) => prev + 1);
    } catch (err) {
      toast.error(`Failed to save user: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const handleCancel = () => {
    setView('list');
    setSelectedUser(null);
  };

  const tabs = ['details', 'groups'] as const;
  const activeTabIndex = tabs.indexOf(activeTab);

  return (
    <div className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6">
      <motion.div
        className="flex flex-col md:flex-row md:items-center justify-between gap-6 pt-4"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: duration.normal, ...springConfig.gentle }}
      >
        <div className="space-y-1">
          <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
            Users
          </h1>
          <p className="text-slate-500 dark:text-slate-400 font-medium">
            Manage system users, roles, and status.
          </p>
        </div>

        <motion.div
          className="flex items-center gap-4"
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: duration.normal, delay: 0.1 }}
        >
          <motion.button
            onClick={handleCreateNew}
            className="btn btn-primary h-11 px-8 rounded-2xl text-sm font-bold sleek-shadow inline-flex items-center gap-2"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            <PlusIcon className="w-5 h-5" />
            Create User
          </motion.button>
        </motion.div>
      </motion.div>

      <motion.div
        className="space-y-6"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: duration.slow, delay: 0.2 }}
      >
        <UserList key={refreshKey} onEditUser={handleEditUser} selectedId={null} />
      </motion.div>

      <AnimatePresence>
        {(view === 'create' || view === 'edit') && (
          <Modal
            isOpen={true}
            onClose={handleCancel}
            title={view === 'create' ? 'Add User' : 'Edit User'}
            size="lg"
          >
            {view === 'edit' && (
              <motion.div
                className="flex border-b border-slate-200 dark:border-slate-700 mb-4 relative"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
              >
                {tabs.map((tab) => (
                  <button
                    key={tab}
                    className={`px-4 py-2 text-sm font-medium relative z-10 transition-colors ${
                      activeTab === tab
                        ? 'text-blue-600'
                        : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
                    }`}
                    onClick={() => setActiveTab(tab)}
                  >
                    {tab.charAt(0).toUpperCase() + tab.slice(1)}
                  </button>
                ))}
                <motion.div
                  className="absolute bottom-0 h-0.5 bg-blue-500"
                  initial={false}
                  animate={{
                    x: activeTabIndex * 80,
                    width: 70,
                  }}
                  transition={{ duration: 0.25, ...springConfig.snappy }}
                />
              </motion.div>
            )}

            <AnimatePresence mode="wait">
              {activeTab === 'details' ? (
                <motion.div
                  key="details"
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: 10 }}
                  transition={{ duration: 0.2 }}
                >
                  <UserForm
                    user={selectedUser || undefined}
                    onSave={handleSave}
                    onCancel={handleCancel}
                  />
                </motion.div>
              ) : (
                <motion.div
                  key="groups"
                  initial={{ opacity: 0, x: 10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -10 }}
                  transition={{ duration: 0.2 }}
                >
                  {selectedUser && <UserGroupsTab user={selectedUser} />}
                </motion.div>
              )}
            </AnimatePresence>
          </Modal>
        )}
      </AnimatePresence>
    </div>
  );
}
