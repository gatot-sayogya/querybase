'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { User } from '@/types';
import UserList from './UserList';
import UserForm from './UserForm';

export default function UserManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setSelectedUser(null);
    setView('create');
  };

  const handleEditUser = (user: User) => {
    setSelectedUser(user);
    setView('edit');
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
      } else if (view === 'edit' && selectedUser) {
        await apiClient.updateUser(selectedUser.id, {
          email: data.email,
          username: data.username,
          full_name: data.full_name,
          role: data.role,
          is_active: data.is_active,
        });
      }

      setView('list');
      setSelectedUser(null);
      setRefreshKey((prev) => prev + 1);
    } catch (err) {
      alert(`Failed to save user: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const handleCancel = () => {
    setView('list');
    setSelectedUser(null);
  };

  return (
    <div className="space-y-6">
      {view === 'list' && (
        <>
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Users</h1>
            <button
              onClick={handleCreateNew}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Add User
            </button>
          </div>
          <UserList key={refreshKey} onEditUser={handleEditUser} selectedId={null} />
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
              {view === 'create' ? 'Add User' : 'Edit User'}
            </h1>
          </div>
          <div className="max-w-2xl">
            <UserForm
              user={selectedUser || undefined}
              onSave={handleSave}
              onCancel={handleCancel}
            />
          </div>
        </>
      )}
    </div>
  );
}
