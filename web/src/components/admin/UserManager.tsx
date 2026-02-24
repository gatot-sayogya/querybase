'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { User } from '@/types';
import UserList from './UserList';
import UserForm from './UserForm';
import Modal from '../Modal';

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
      toast.error(`Failed to save user: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const handleCancel = () => {
    setView('list');
    setSelectedUser(null);
  };

  return (
    <div className="space-y-6">
      <div className="page-header">
        <div>
          <h1 className="page-title">Users</h1>
          <p className="page-subtitle">Manage user access and permissions</p>
        </div>
        <button
          onClick={handleCreateNew}
          className="btn btn-primary"
        >
          + Add User
        </button>
      </div>
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <UserList key={refreshKey} onEditUser={handleEditUser} selectedId={null} />
      </div>

      <Modal 
        isOpen={view === 'create' || view === 'edit'} 
        onClose={handleCancel}
        title={view === 'create' ? 'Add User' : 'Edit User'}
      >
        <UserForm
          user={selectedUser || undefined}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      </Modal>
    </div>
  );
}
