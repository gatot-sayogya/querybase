'use client';

import { useState, useEffect } from 'react';
import type { Group } from '@/types';

interface GroupFormProps {
  group?: Group;
  onSave: (data: { name: string; description: string }) => void;
  onCancel: () => void;
}

export default function GroupForm({ group, onSave, onCancel }: GroupFormProps) {
  const [formData, setFormData] = useState({
    name: group?.name || '',
    description: group?.description || '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Group name is required';
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
    // The actual API call will be handled by the parent component
    // Pass the form data back to parent
    onSave(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-10 mt-6 md:max-w-3xl">
      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start relative group">
        <div className="md:sticky md:top-4 md:mt-2">
          <label htmlFor="name" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Group Name <span className="text-[var(--accent-red)]">*</span>
          </label>
          <p className="text-xs text-[var(--text-faint)] max-w-[200px] leading-relaxed">Unique identifier for this group of users.</p>
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
            placeholder="e.g. Data Analysts"
          />
          {errors.name && <div className="text-[var(--red-text)] text-xs mt-2">{errors.name}</div>}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start group">
        <div className="md:mt-2">
          <label htmlFor="description" className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1">
            Description
          </label>
          <p className="text-xs text-[var(--text-faint)] max-w-[200px] leading-relaxed">What is the primary purpose of this group?</p>
        </div>
        <div className="relative">
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows={3}
            className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)] resize-y"
            placeholder="Users who can access and analyze data..."
          />
        </div>
      </div>

      {/* Info message & Actions separated from grid */}
      <div className="mt-12 pt-8 border-t border-[var(--border-light)] grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8">
        <div className="hidden md:block"></div>
        <div className="flex flex-col gap-10">
          <div className="flex gap-4 p-5 bg-[var(--bg-page)] border border-[var(--border-light)] shadow-sm" style={{ borderRadius: '2px' }}>
            <svg className="shrink-0 text-[var(--accent-blue)] w-5 h-5 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm text-[var(--text-muted)] leading-relaxed">
              {group ? (
                <>After updating the group, you can manage users from the group detail page.</>
              ) : (
                <>After creating the group, you can add users from the group detail page.</>
              )}
            </p>
          </div>

          <div className="flex gap-6 items-center">
            <button
              type="submit"
              disabled={saving}
              className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50"
              style={{ borderRadius: '2px' }}
            >
              {saving ? 'Saving...' : group ? 'Update Group' : 'Save Group'}
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
      </div>
    </form>
  );
}
