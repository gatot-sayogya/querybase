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
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 mt-6 md:max-w-xl w-full">
      <div className="flex flex-col gap-1.5 items-start relative">
        <label htmlFor="name" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Group Name <span className="text-[var(--accent-red)]">*</span>
        </label>
        <p className="text-[10px] text-[var(--text-faint)] leading-relaxed pl-1 -mt-1 mb-1">Unique identifier for this group of users.</p>
        <div className="relative w-full">
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            className="w-full bg-[var(--input-bg)] px-4 py-3 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent"
            style={errors.name ? { borderColor: 'var(--red-text)', backgroundColor: 'var(--red-bg)' } : {}}
            placeholder="e.g. Data Analysts"
          />
          {errors.name && <div className="text-[var(--red-text)] text-xs mt-2 pl-1">{errors.name}</div>}
        </div>
      </div>

      <div className="flex flex-col gap-1.5 items-start relative mt-2">
        <label htmlFor="description" className="text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] pl-1">
          Description
        </label>
        <p className="text-[10px] text-[var(--text-faint)] leading-relaxed pl-1 -mt-1 mb-1">What is the primary purpose of this group?</p>
        <div className="relative w-full">
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows={3}
            className="w-full bg-[var(--input-bg)] px-4 py-3 text-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-blue)] transition-all rounded-xl placeholder-[var(--text-faint)] border border-transparent resize-y"
            placeholder="Users who can access and analyze data..."
          />
        </div>
      </div>

      {/* Info message & Actions separated from grid */}
      <div className="mt-8 pt-6 border-t border-[var(--border-light)] flex flex-col gap-6 w-full">
        <div className="flex gap-4 p-4 bg-[var(--input-bg)] text-[var(--text-muted)] rounded-xl items-start">
          <svg className="shrink-0 text-[var(--accent-blue)] w-5 h-5 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm leading-relaxed">
            {group ? (
              <>After updating the group, you can manage users from the group detail page.</>
            ) : (
              <>After creating the group, you can add users from the group detail page.</>
            )}
          </p>
        </div>

        <div className="flex justify-end gap-3 w-full">
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
            {saving ? 'Saving...' : group ? 'Update' : 'Save'}
          </button>
        </div>
      </div>
    </form>
  );
}
