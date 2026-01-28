'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { DataSource } from '@/types';
import DataSourceList from './DataSourceList';
import DataSourceForm from './DataSourceForm';

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
      {view === 'list' && (
        <>
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Data Sources</h1>
            <button
              onClick={handleCreateNew}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Add Data Source
            </button>
          </div>
          <DataSourceList key={refreshKey} onEditDataSource={handleEditDataSource} selectedId={null} />
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
              {view === 'create' ? 'Add Data Source' : 'Edit Data Source'}
            </h1>
          </div>
          <div className="max-w-2xl">
            <DataSourceForm
              dataSource={selectedDataSource || undefined}
              onSave={handleSave}
              onCancel={handleCancel}
            />
          </div>
        </>
      )}
    </div>
  );
}
