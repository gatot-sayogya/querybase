'use client';

import { useState, useEffect } from 'react';
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
  };

  const handleNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: parseInt(value) || 0 }));
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
      } else {
        // Create new data source
        await apiClient.createDataSource(formData);
      }

      onSave();
    } catch (err) {
      alert(`Failed to save data source: ${err instanceof Error ? err.message : 'Unknown error'}`);
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
      if (dataSource) {
        await apiClient.testDataSourceConnection(dataSource.id, formData);
      }
      alert('Connection successful!');
    } catch (err) {
      alert(`Connection failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setTesting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
          {dataSource ? 'Edit Data Source' : 'Add New Data Source'}
        </h2>
      </div>

      {/* Name */}
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Name *
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={formData.name}
          onChange={handleChange}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${
            errors.name ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
          }`}
          placeholder="Production Database"
        />
        {errors.name && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.name}</p>}
      </div>

      {/* Type */}
      <div>
        <label htmlFor="type" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Database Type *
        </label>
        <select
          id="type"
          name="type"
          value={formData.type}
          onChange={handleChange}
          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
        >
          <option value="postgresql">PostgreSQL</option>
          <option value="mysql">MySQL</option>
        </select>
      </div>

      {/* Host */}
      <div>
        <label htmlFor="host" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Host *
        </label>
        <input
          type="text"
          id="host"
          name="host"
          value={formData.host}
          onChange={handleChange}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${
            errors.host ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
          }`}
          placeholder="localhost or db.example.com"
        />
        {errors.host && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.host}</p>}
      </div>

      {/* Port */}
      <div>
        <label htmlFor="port" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Port *
        </label>
        <input
          type="number"
          id="port"
          name="port"
          value={formData.port}
          onChange={handleNumberChange}
          min={1}
          max={65535}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${
            errors.port ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
          }`}
          placeholder="5432"
        />
        {errors.port && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.port}</p>}
      </div>

      {/* Database Name */}
      <div>
        <label htmlFor="database_name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Database Name *
        </label>
        <input
          type="text"
          id="database_name"
          name="database_name"
          value={formData.database_name}
          onChange={handleChange}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${
            errors.database_name ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
          }`}
          placeholder="querybase"
        />
        {errors.database_name && (
          <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.database_name}</p>
        )}
      </div>

      {/* Username */}
      <div>
        <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Username *
        </label>
        <input
          type="text"
          id="username"
          name="username"
          value={formData.username}
          onChange={handleChange}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${
            errors.username ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
          }`}
          placeholder="dbuser"
        />
        {errors.username && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.username}</p>}
      </div>

      {/* Password */}
      <div>
        <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Password {!dataSource && '*'}
          {dataSource && '(leave empty to keep current)'}
        </label>
        <input
          type="password"
          id="password"
          name="password"
          value={formData.password}
          onChange={handleChange}
          className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${
            errors.password ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
          }`}
          placeholder="••••••••••"
        />
        {errors.password && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.password}</p>}
      </div>

      {/* Actions */}
      <div className="flex justify-between items-center pt-4 border-t border-gray-200 dark:border-gray-700">
        <div className="flex space-x-3">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          >
            {saving ? 'Saving...' : dataSource ? 'Update' : 'Create'}
          </button>
          <button
            type="button"
            onClick={handleTestConnection}
            disabled={testing}
            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50"
          >
            {testing ? 'Testing...' : 'Test Connection'}
          </button>
        </div>
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
