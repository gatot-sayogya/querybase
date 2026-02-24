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
    <form onSubmit={handleSubmit} className="flex flex-col gap-10 mt-6 md:max-w-4xl">
      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="email" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Email <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.email ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="user@example.com"
          />
          {errors.email && <div className="text-[var(--red-text)] text-xs mt-2">{errors.email}</div>}
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
            placeholder="johndoe"
          />
          {errors.username && <div className="text-[var(--red-text)] text-xs mt-2">{errors.username}</div>}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="full_name" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Full Name <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <input
            type="text"
            id="full_name"
            name="full_name"
            value={formData.full_name}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.full_name ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="John Doe"
          />
          {errors.full_name && <div className="text-[var(--red-text)] text-xs mt-2">{errors.full_name}</div>}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="role" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Role <span className="text-[var(--accent-red)]">*</span>
          </label>
        </div>
        <div className="relative">
          <select
            id="role"
            name="role"
            value={formData.role}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 pr-8 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none cursor-pointer"
          >
            <option value="user" className="bg-[var(--card-bg)] text-[var(--text-primary)]">User</option>
            <option value="admin" className="bg-[var(--card-bg)] text-[var(--text-primary)]">Admin</option>
            <option value="viewer" className="bg-[var(--card-bg)] text-[var(--text-primary)]">Viewer</option>
          </select>
          <div className="text-[var(--text-faint)] text-xs mt-3 leading-relaxed">
            {formData.role === 'admin' && 'Full access to all features including user management.'}
            {formData.role === 'user' && 'Can execute queries and submit approval requests.'}
            {formData.role === 'viewer' && 'Read-only access to queries and results.'}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
        <div className="md:mt-2">
          <label htmlFor="password" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Password {!user && <span className="text-[var(--accent-red)]">*</span>}
          </label>
          {user && <p className="text-xs text-[var(--text-faint)] mt-1">Leave empty to keep current</p>}
        </div>
        <div className="relative">
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]"
            style={errors.password ? { borderBottomColor: 'var(--red-text)' } : {}}
            placeholder="••••••••••"
          />
          {errors.password && <div className="text-[var(--red-text)] text-xs mt-2">{errors.password}</div>}
        </div>
      </div>

      {user && (
        <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
          <div className="md:mt-2">
            <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
              Account Status
            </label>
          </div>
          <div className="relative flex items-center gap-4 mt-2">
            <input
              type="checkbox"
              id="is_active"
              name="is_active"
              checked={formData.is_active}
              onChange={handleChange}
              className="w-5 h-5 accent-[var(--accent-blue)] cursor-pointer"
            />
            <label htmlFor="is_active" className="text-sm font-medium text-[var(--text-primary)] cursor-pointer select-none">
              Active User
            </label>
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="mt-8 pt-8 border-t border-[var(--border-light)] grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8">
        <div className="hidden md:block"></div>
        <div className="flex gap-6 items-center">
          <button
            type="submit"
            disabled={saving}
            className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50"
            style={{ borderRadius: '2px' }}
          >
            {saving ? 'Saving...' : user ? 'Update User' : 'Save User'}
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
