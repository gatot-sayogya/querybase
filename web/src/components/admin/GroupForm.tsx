'use client';

import { useState, useEffect } from 'react';
import type { Group } from '@/types';

interface GroupFormProps {
  group?: Group;
  onSave: () => void;
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
    onSave();
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="form-group">
        <label htmlFor="name">Group Name <span style={{ color: 'var(--accent-red)' }}>*</span></label>
        <input
          type="text"
          id="name"
          name="name"
          value={formData.name}
          onChange={handleChange}
          className="input-field"
          style={errors.name ? { borderColor: 'var(--accent-red)' } : {}}
          placeholder="Data Analysts"
        />
        {errors.name && <div style={{ color: 'var(--accent-red)', fontSize: '13px', marginTop: '4px' }}>{errors.name}</div>}
      </div>

      <div className="form-group">
        <label htmlFor="description">Description</label>
        <textarea
          id="description"
          name="description"
          value={formData.description}
          onChange={handleChange}
          className="input-field"
          rows={3}
          style={{ resize: 'vertical' }}
          placeholder="Users who can access and analyze data"
        />
      </div>

      {/* Info message */}
      <div style={{ backgroundColor: 'var(--bg-hover)', padding: '16px', borderRadius: '8px', display: 'flex', gap: '12px', marginBottom: '24px' }}>
        <svg
          style={{ flexShrink: 0, color: 'var(--accent-blue)', width: '20px', height: '20px' }}
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path
            fillRule="evenodd"
            d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 102 0 1 0 2 0 1 0 00-2z"
            clipRule="evenodd"
          />
        </svg>
        <div style={{ fontSize: '14px', color: 'var(--text-primary)', lineHeight: 1.5 }}>
          {group ? (
            <>After updating the group, you can manage users from the group detail page.</>
          ) : (
            <>After creating the group, you can add users from the group detail page.</>
          )}
        </div>
      </div>

      {/* Actions */}
      <div style={{ marginTop: '24px', display: 'flex', gap: '12px', alignItems: 'center' }}>
        <button
          type="submit"
          disabled={saving}
          className="btn btn-primary"
          style={saving ? { opacity: 0.7, cursor: 'not-allowed' } : {}}
        >
          {saving ? 'Saving...' : group ? 'Update Group' : 'Save Group'}
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
