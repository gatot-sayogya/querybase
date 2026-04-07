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
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 mt-6 md:max-w-xl w-full">
      <div className="flex flex-col gap-1.5 items-start relative">
        <label htmlFor="email" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Email <span className="text-[var(--accent-red)]">*</span>
        </label>
        <div className="relative w-full">
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            className="w-full bg-[var(--input-bg)] px-4 py-3 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent"
            style={errors.email ? { borderColor: 'var(--red-text)', backgroundColor: 'var(--red-bg)' } : {}}
            placeholder="user@example.com"
          />
          {errors.email && <div className="text-[var(--red-text)] text-xs mt-2 pl-1">{errors.email}</div>}
        </div>
      </div>

      <div className="flex flex-col gap-1.5 items-start relative">
        <label htmlFor="username" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Username <span className="text-[var(--accent-red)]">*</span>
        </label>
        <div className="relative w-full">
          <input
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            className="w-full bg-[var(--input-bg)] px-4 py-3 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent"
            style={errors.username ? { borderColor: 'var(--red-text)', backgroundColor: 'var(--red-bg)' } : {}}
            placeholder="johndoe"
          />
          {errors.username && <div className="text-[var(--red-text)] text-xs mt-2 pl-1">{errors.username}</div>}
        </div>
      </div>

      <div className="flex flex-col gap-1.5 items-start relative">
        <label htmlFor="full_name" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Full Name <span className="text-[var(--accent-red)]">*</span>
        </label>
        <div className="relative w-full">
          <input
            type="text"
            id="full_name"
            name="full_name"
            value={formData.full_name}
            onChange={handleChange}
            className="w-full bg-[var(--input-bg)] px-4 py-3 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent"
            style={errors.full_name ? { borderColor: 'var(--red-text)', backgroundColor: 'var(--red-bg)' } : {}}
            placeholder="John Doe"
          />
          {errors.full_name && <div className="text-[var(--red-text)] text-xs mt-2 pl-1">{errors.full_name}</div>}
        </div>
      </div>

      <div className="flex flex-col gap-1.5 items-start relative">
        <label htmlFor="role" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Role <span className="text-[var(--accent-red)]">*</span>
        </label>
        <div className="relative w-full">
          <select
            id="role"
            name="role"
            value={formData.role}
            onChange={handleChange}
            className="w-full bg-[var(--input-bg)] px-4 py-3 pr-10 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl cursor-pointer border border-transparent appearance-none"
          >
            <option value="user" className="bg-[var(--card-bg)] text-[var(--text-primary)]">User</option>
            <option value="admin" className="bg-[var(--card-bg)] text-[var(--text-primary)]">Admin</option>
            <option value="viewer" className="bg-[var(--card-bg)] text-[var(--text-primary)]">Viewer</option>
          </select>
          <div className="pointer-events-none absolute top-[14px] right-4 flex items-center text-[var(--text-muted)]">
            <svg className="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"/></svg>
          </div>
          <div className="text-[var(--text-faint)] text-xs mt-3 leading-relaxed px-2">
            {formData.role === 'admin' && 'Full access to all features including user management.'}
            {formData.role === 'user' && 'Can execute queries and submit approval requests.'}
            {formData.role === 'viewer' && 'Read-only access to queries and results.'}
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-1.5 items-start relative">
        <label htmlFor="password" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1 flex items-center gap-2">
          Password {!user && <span className="text-[var(--accent-red)]">*</span>}
          {user && <span className="text-[10px] normal-case tracking-normal opacity-70 font-normal mt-0.5">(Leave empty to keep current)</span>}
        </label>
        <div className="relative w-full">
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className="w-full bg-[var(--input-bg)] px-4 py-3 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent"
            style={errors.password ? { borderColor: 'var(--red-text)', backgroundColor: 'var(--red-bg)' } : {}}
            placeholder="••••••••••"
          />
          {errors.password && <div className="text-[var(--red-text)] text-xs mt-2 pl-1">{errors.password}</div>}
        </div>
      </div>

      {user && (
        <div className="flex items-center gap-3 mt-2 pl-1">
          <input
            type="checkbox"
            id="is_active"
            name="is_active"
            checked={formData.is_active}
            onChange={handleChange}
            className="w-5 h-5 cursor-pointer rounded"
          />
          <label htmlFor="is_active" className="text-sm font-medium text-[var(--text-primary)] cursor-pointer select-none">
            Active User
          </label>
        </div>
      )}

      {/* Actions */}
      <div className="mt-8 pt-6 border-t border-[var(--border-light)] flex justify-end gap-3 w-full">
        <button
          type="button"
          onClick={onCancel}
          className="h-12 px-6 bg-[var(--input-bg)] text-[var(--text-primary)] text-sm font-bold tracking-[0.1em] uppercase hover:bg-[var(--border)] transition-colors rounded-xl"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={saving}
          className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50 rounded-xl"
        >
          {saving ? 'Saving...' : user ? 'Update' : 'Save'}
        </button>
      </div>
    </form>
  );
}
