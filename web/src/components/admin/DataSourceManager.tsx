'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
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
        </>
      )}

      {(view === 'create' || view === 'edit') && (
        <>
          <div className="page-header" style={{ marginBottom: '20px' }}>
            <div>
              <button
                onClick={handleCancel}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: '13px', cursor: 'pointer', padding: 0, marginBottom: '8px' }}
              >
                ← Back to Data Sources
              </button>
              <h1 className="page-title">
                {view === 'create' ? 'Add Data Source' : 'Edit Data Source'}
              </h1>
            </div>
          </div>
          <div className="card card-padded" style={{ maxWidth: '600px' }}>
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
