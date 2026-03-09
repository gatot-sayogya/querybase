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
    <div className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 pt-4">
        <div className="space-y-1">
          <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
            Data Infrastructure
          </h1>
          <p className="text-slate-500 dark:text-slate-400 font-medium">
            Configure system bridges and secure data source connections.
          </p>
        </div>
        
        <div className="flex items-center gap-4">
          <button
            onClick={handleCreateNew}
            className="btn btn-primary h-11 px-8 rounded-2xl text-sm font-bold sleek-shadow"
          >
            <span className="text-xl mr-2">+</span>
            Provision Source
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="space-y-6">
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
