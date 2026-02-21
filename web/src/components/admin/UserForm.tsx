'use client';

import { useState, useEffect } from 'react';
import type { User } from '@/types';

interface UserFormProps {
  user?: User;
  onSave: (data: {
    email: string;
    username: string;
    password: string;
    full_name: string;
    role: 'admin' | 'user' | 'viewer';
    is_active: boolean;
  }) => void;
  onCancel: () => void;
}

export default function UserForm({ user, onSave, onCancel }: UserFormProps) {
  const [formData, setFormData] = useState({
    email: user?.email || '',
    username: user?.username || '',
    password: '',
    full_name: user?.full_name || '',
    role: (user?.role || 'user') as 'admin' | 'user' | 'viewer',
    is_active: user?.is_active ?? true,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
    }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Invalid email format';
    }

    if (!formData.username.trim()) {
      newErrors.username = 'Username is required';
    } else if (formData.username.length < 3) {
      newErrors.username = 'Username must be at least 3 characters';
    }

    if (!formData.full_name.trim()) {
      newErrors.full_name = 'Full name is required';
    }

    if (!user && !formData.password.trim()) {
      newErrors.password = 'Password is required for new users';
    } else if (formData.password && formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    setSaving(true);
    // Pass the form data back to parent
    onSave(formData);
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="form-group">
        <label htmlFor="email">Email <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <input
          type="email"
          id="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          className="input-field"
          style={errors.email ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="user@example.com"
        />
        {errors.email && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.email}</div>}
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
          placeholder="johndoe"
        />
        {errors.username && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.username}</div>}
      </div>

      <div className="form-group">
        <label htmlFor="full_name">Full Name <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <input
          type="text"
          id="full_name"
          name="full_name"
          value={formData.full_name}
          onChange={handleChange}
          className="input-field"
          style={errors.full_name ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="John Doe"
        />
        {errors.full_name && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.full_name}</div>}
      </div>

      <div className="form-group">
        <label htmlFor="role">Role <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <select
          id="role"
          name="role"
          value={formData.role}
          onChange={handleChange}
          className="input-field"
        >
          <option value="user">User</option>
          <option value="admin">Admin</option>
          <option value="viewer">Viewer</option>
        </select>
        <div style={{ color: 'var(--text-muted)', fontSize: '13px', marginTop: '4px' }}>
          {formData.role === 'admin' && 'Full access to all features including user management'}
          {formData.role === 'user' && 'Can execute queries and submit approval requests'}
          {formData.role === 'viewer' && 'Read-only access to queries and results'}
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="password">
          Password {!user && <span style={{ color: 'var(--accent-red)' }}>*</span>}
          {user && <span style={{ color: 'var(--text-muted)', fontWeight: 400, marginLeft: '6px' }}>(leave empty to keep current)</span>}
        </label>
        <input
          type="password"
          id="password"
          name="password"
          value={formData.password}
          onChange={handleChange}
          className="input-field"
          style={errors.password ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="••••••••••"
        />
        {errors.password && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.password}</div>}
      </div>

      {/* Active Status */}
      {user && (
        <div className="form-group" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <input
            type="checkbox"
            id="is_active"
            name="is_active"
            checked={formData.is_active}
            onChange={handleChange}
            style={{ width: '16px', height: '16px', cursor: 'pointer' }}
          />
          <label htmlFor="is_active" style={{ marginBottom: 0, fontWeight: 500, cursor: 'pointer' }}>
            Active User
          </label>
        </div>
      )}

      {/* Actions */}
      <div style={{ marginTop: '24px', display: 'flex', gap: '12px', alignItems: 'center' }}>
        <button
          type="submit"
          disabled={saving}
          className="btn btn-primary"
          style={saving ? { opacity: 0.7, cursor: 'not-allowed' } : {}}
        >
          {saving ? 'Saving...' : user ? 'Update User' : 'Save User'}
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
