'use client';

import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { DataSource, CreateDataSourceRequest } from '@/types';

interface DataSourceFormProps {
  dataSource?: DataSource;
  onSave: () => void;
  onCancel: () => void;
}

export default function DataSourceForm({ dataSource, onSave, onCancel }: DataSourceFormProps) {
  const [formData, setFormData] = useState<CreateDataSourceRequest>({
    name: dataSource?.name || '',
    type: dataSource?.type || 'postgresql',
    host: dataSource?.host || '',
    port: dataSource?.port || (dataSource?.type === 'mysql' ? 3306 : 5432),
    database_name: dataSource?.database_name || '',
    username: dataSource?.username || '',
    password: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<{
    type: 'success' | 'error' | null;
    message: string;
  }>({ type: null, message: '' });

  useEffect(() => {
    if (formData.type === 'mysql') {
      setFormData((prev) => ({ ...prev, port: 3306 }));
    } else {
      setFormData((prev) => ({ ...prev, port: 5432 }));
    }
  }, [formData.type]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
    }
    // Clear connection status when form changes
    if (connectionStatus.type) {
      setConnectionStatus({ type: null, message: '' });
    }
  };

  const handleNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: parseInt(value) || 0 }));
    if (connectionStatus.type) {
      setConnectionStatus({ type: null, message: '' });
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    if (!formData.host.trim()) {
      newErrors.host = 'Host is required';
    }
    if (!formData.port || formData.port < 1 || formData.port > 65535) {
      newErrors.port = 'Valid port is required (1-65535)';
    }
    if (!formData.database_name.trim()) {
      newErrors.database_name = 'Database name is required';
    }
    if (!formData.username.trim()) {
      newErrors.username = 'Username is required';
    }
    if (!dataSource && !formData.password.trim()) {
      newErrors.password = 'Password is required for new data sources';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      setSaving(true);

      if (dataSource) {
        // Update existing data source
        const updateData: Partial<CreateDataSourceRequest> = { ...formData };
        if (!formData.password) {
          delete updateData.password;
        }
        await apiClient.updateDataSource(dataSource.id, updateData);
        toast.success('Data source updated successfully! ✓', {
          duration: 5000,
        });
      } else {
        // Create new data source
        await apiClient.createDataSource(formData);
        toast.success('Data source created successfully! ✓', {
          duration: 5000,
        });
      }

      onSave();
    } catch (err: any) {
      const errorMessage = err?.response?.data?.error || err?.message || 'Unknown error';
      toast.error(`Failed to save data source: ${errorMessage}`, {
        duration: 7000,
      });
    } finally {
      setSaving(false);
    }
  };

  const handleTestConnection = async () => {
    if (!validate()) {
      return;
    }

    try {
      setTesting(true);
      setConnectionStatus({ type: null, message: '' });
      
      if (dataSource) {
        await apiClient.testDataSourceConnection(dataSource.id, formData);
      } else {
        await apiClient.testNewDataSourceConnection(formData);
      }
      
      setConnectionStatus({
        type: 'success',
        message: 'Connection successful! Your database credentials are valid.',
      });
      toast.success('Connection successful! ✓', {
        duration: 5000,
      });
    } catch (err: any) {
      const errorMessage = err?.response?.data?.error || err?.message || 'Unknown error';
      setConnectionStatus({
        type: 'error',
        message: errorMessage,
      });
      toast.error(`Connection failed: ${errorMessage}`, {
        duration: 7000,
      });
    } finally {
      setTesting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Connection Status Banner */}
      {connectionStatus.type && (
        <div
          className={`p-4 rounded-md mb-6 ${
            connectionStatus.type === 'success'
              ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
              : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
          }`}
        >
          <div className="flex items-start">
            <div className="flex-shrink-0">
              {connectionStatus.type === 'success' ? (
                <svg className="h-5 w-5 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              ) : (
                <svg className="h-5 w-5 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              )}
            </div>
            <div className="ml-3 flex-1">
              <h3 className={`text-sm font-medium ${connectionStatus.type === 'success' ? 'text-green-800 dark:text-green-200' : 'text-red-800 dark:text-red-200'}`}>
                {connectionStatus.type === 'success' ? 'Connection Successful' : 'Connection Failed'}
              </h3>
              <p className={`mt-1 text-sm ${connectionStatus.type === 'success' ? 'text-green-700 dark:text-green-300' : 'text-red-700 dark:text-red-300'}`}>
                {connectionStatus.message}
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="form-group">
        <label htmlFor="name">Name <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <input
          type="text"
          id="name"
          name="name"
          value={formData.name}
          onChange={handleChange}
          className="input-field"
          style={errors.name ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="Production Database"
        />
        {errors.name && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.name}</div>}
      </div>

      <div className="form-group">
        <label htmlFor="type">Database Type <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <select
          id="type"
          name="type"
          value={formData.type}
          onChange={handleChange}
          className="input-field"
        >
          <option value="postgresql">PostgreSQL</option>
          <option value="mysql">MySQL</option>
        </select>
      </div>

      <div style={{ display: 'flex', gap: '16px', marginBottom: '16px' }}>
        <div className="form-group" style={{ flex: 2, marginBottom: 0 }}>
          <label htmlFor="host">Host <span style={{ color: 'var(--accent-red)' }}>*</span></label>
          <input
            type="text"
            id="host"
            name="host"
            value={formData.host}
            onChange={handleChange}
            className="input-field"
            style={errors.host ? { borderColor: 'var(--accent-red)' } : {}}
            placeholder="db.example.com"
          />
          {errors.host && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.host}</div>}
        </div>
        <div className="form-group" style={{ flex: 1, marginBottom: 0 }}>
          <label htmlFor="port">Port <span style={{ color: 'var(--accent-red)' }}>*</span></label>
          <input
            type="number"
            id="port"
            name="port"
            value={formData.port || ''}
            onChange={handleNumberChange}
            min={1}
            max={65535}
            className="input-field"
            style={errors.port ? { borderColor: 'var(--accent-red)' } : {}}
            placeholder="5432"
          />
          {errors.port && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.port}</div>}
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="database_name">Database Name <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <input
          type="text"
          id="database_name"
          name="database_name"
          value={formData.database_name}
          onChange={handleChange}
          className="input-field"
          style={errors.database_name ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="querybase"
        />
        {errors.database_name && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.database_name}</div>}
      </div>

      <div className="form-group">
        <label htmlFor="username">Username <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <input
          type="text"
          id="username"
          name="username"
          value={formData.username}
          onChange={handleChange}
          className="input-field"
          style={errors.username ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="dbuser"
        />
        {errors.username && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.username}</div>}
      </div>

      <div className="form-group">
        <label htmlFor="password">
          Password {!dataSource && <span style={{ color: 'var(--accent-red)' }}>*</span>}
          {dataSource && <span style={{ color: 'var(--text-muted)', fontWeight: 400, marginLeft: '6px' }}>(leave empty to keep current)</span>}
        </label>
        <input
          type="password"
          id="password"
          name="password"
          value={formData.password || ''}
          onChange={handleChange}
          className="input-field"
          style={errors.password ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="••••••••••"
        />
        {errors.password && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.password}</div>}
      </div>

      <div style={{ marginTop: '24px', display: 'flex', gap: '12px', alignItems: 'center' }}>
        <button
          type="submit"
          disabled={saving}
          className="btn btn-primary"
          style={saving ? { opacity: 0.7, cursor: 'not-allowed' } : {}}
        >
          {saving ? 'Saving...' : dataSource ? 'Update Connection' : 'Save Connection'}
        </button>
        <button
          type="button"
          onClick={handleTestConnection}
          disabled={testing}
          className="btn btn-ghost"
          style={testing ? { opacity: 0.7, cursor: 'not-allowed' } : {}}
        >
          {testing ? 'Testing...' : 'Test Connection'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          style={{ marginLeft: 'auto', color: 'var(--text-muted)', background: 'none', border: 'none', cursor: 'pointer', fontSize: '14px', fontWeight: 500 }}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
