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
    <form onSubmit={handleSubmit} className="flex flex-col gap-10 mt-6 md:max-w-4xl">
      {/* Connection Status Banner */}
      {connectionStatus.type && (
        <div
          className={`p-4 rounded-sm border ${
            connectionStatus.type === 'success'
              ? 'bg-[var(--green-bg)] border-[var(--green-text)]'
              : 'bg-[var(--red-bg)] border-[var(--red-border)]'
          }`}
        >
          <div className="flex items-start">
            <div className="flex-shrink-0">
              {connectionStatus.type === 'success' ? (
                <svg className="h-5 w-5 text-[var(--green-text)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              ) : (
                <svg className="h-5 w-5 text-[var(--red-text)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              )}
            </div>
            <div className="ml-3 flex-1">
              <h3 className={`text-sm font-bold tracking-[0.05em] uppercase ${connectionStatus.type === 'success' ? 'text-[var(--text-primary)]' : 'text-[var(--red-text)]'}`}>
                {connectionStatus.type === 'success' ? 'Connection Successful' : 'Connection Failed'}
              </h3>
              <p className={`mt-1 text-sm ${connectionStatus.type === 'success' ? 'text-[var(--text-muted)]' : 'text-[var(--red-text)] opacity-80'}`}>
                {connectionStatus.message}
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="name" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Name <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.name ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="Production Database"
          />
          {errors.name && <div className="text-[var(--red-text)] text-xs mt-2">{errors.name}</div>}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="type" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Database Type <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <select
            id="type"
            name="type"
            value={formData.type}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 pr-8 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none cursor-pointer"
          >
            <option value="postgresql" className="bg-[var(--card-bg)] text-[var(--text-primary)]">PostgreSQL</option>
            <option value="mysql" className="bg-[var(--card-bg)] text-[var(--text-primary)]">MySQL</option>
          </select>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="host" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Host & Port <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="flex gap-4 sm:gap-6 w-full">
          <div className="relative flex-1">
            <input
              type="text"
              id="host"
              name="host"
              value={formData.host}
              onChange={handleChange}
              className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
              style={errors.host ? { borderBottomColor: 'var(--red-text)' } : {}}
              placeholder="db.example.com"
            />
            {errors.host && <div className="text-[var(--red-text)] text-xs mt-2">{errors.host}</div>}
          </div>
          <div className="relative w-24 sm:w-32 shrink-0">
            <input
              type="number"
              id="port"
              name="port"
              value={formData.port || ''}
              onChange={handleNumberChange}
              min={1}
              max={65535}
              className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
              style={errors.port ? { borderBottomColor: 'var(--red-text)' } : {}}
              placeholder="5432"
            />
            {errors.port && <div className="text-[var(--red-text)] text-xs mt-2">{errors.port}</div>}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="database_name" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Database Name <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <input
            type="text"
            id="database_name"
            name="database_name"
            value={formData.database_name}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.database_name ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="querybase"
          />
          {errors.database_name && <div className="text-[var(--red-text)] text-xs mt-2">{errors.database_name}</div>}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="username" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Username <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <input
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.username ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="dbuser"
          />
          {errors.username && <div className="text-[var(--red-text)] text-xs mt-2">{errors.username}</div>}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="password" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Password {!dataSource && <span className="text-[var(--accent-red)]">*</span>}
          </label>
          {dataSource && <p className="text-xs text-[var(--text-faint)] mt-1">Leave empty to keep current</p>}
        </div>
        <div className="relative">
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password || ''}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.password ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="••••••••••"
          />
          {errors.password && <div className="text-[var(--red-text)] text-xs mt-2">{errors.password}</div>}
        </div>
      </div>

      {/* Actions */}
      <div className="mt-8 pt-8 border-t border-[var(--border-light)] grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8">
        <div className="hidden md:block"></div>
        <div className="flex gap-4 items-center flex-wrap">
          <button
            type="submit"
            disabled={saving}
            className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50"
            style={{ borderRadius: '2px' }}
          >
            {saving ? 'Saving...' : dataSource ? 'Update Connection' : 'Save Connection'}
          </button>
          <button
            type="button"
            onClick={handleTestConnection}
            disabled={testing}
            className="h-12 px-8 bg-transparent border border-[var(--border)] text-[var(--text-primary)] text-sm font-bold tracking-[0.1em] uppercase hover:border-[var(--accent-blue)] transition-colors disabled:opacity-50"
            style={{ borderRadius: '2px' }}
          >
            {testing ? 'Testing...' : 'Test Connection'}
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="h-12 px-4 text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    </form>
  );
}
