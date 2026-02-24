'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { DataSource } from '@/types';
import DataSourceList from './DataSourceList';
import DataSourceForm from './DataSourceForm';
import Modal from '../Modal';

export default function DataSourceManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [selectedDataSource, setSelectedDataSource] = useState<DataSource | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setSelectedDataSource(null);
    setView('create');
  };

  const handleEditDataSource = (dataSource: DataSource) => {
    setSelectedDataSource(dataSource);
    setView('edit');
  };

  const handleSave = () => {
    setView('list');
    setSelectedDataSource(null);
    setRefreshKey((prev) => prev + 1);
  };

  const handleCancel = () => {
    setView('list');
    setSelectedDataSource(null);
  };

  return (
    <div className="space-y-6">
      <div className="page-header">
        <div>
          <h1 className="page-title">Data Sources</h1>
          <p className="page-subtitle">Manage database connections available to users</p>
        </div>
        <button
          onClick={handleCreateNew}
          className="btn btn-primary"
        >
          + Add Data Source
        </button>
      </div>
      <div className="-mx-2 sm:mx-0">
        <DataSourceList key={refreshKey} onEditDataSource={handleEditDataSource} selectedId={null} />
      </div>

      <Modal 
        isOpen={view === 'create' || view === 'edit'} 
        onClose={handleCancel}
        title={view === 'create' ? 'Add Data Source' : 'Edit Data Source'}
      >
        <DataSourceForm
          dataSource={selectedDataSource || undefined}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      </Modal>
    </div>
  );
}
