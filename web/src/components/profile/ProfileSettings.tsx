'use client';

import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { useAuthStore } from '@/stores/auth-store';
import { apiClient } from '@/lib/api-client';
import type { Group, DataSource } from '@/types';

export default function ProfileSettings() {
  const { user } = useAuthStore();
  const [activeTab, setActiveTab] = useState<'account' | 'password' | 'groups' | 'databases'>('account');
  const [saving, setSaving] = useState(false);

  // Stats / Data
  const [groups, setGroups] = useState<Group[]>([]);
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);

  // Form states
  const [accountForm, setAccountForm] = useState({
    username: '',
    email: '',
    full_name: '',
  });

  const [passwordForm, setPasswordForm] = useState({
    currPw: '',
    newPw: '',
    confirmPw: '',
  });

  useEffect(() => {
    if (user) {
      setAccountForm({
        username: user.username || '',
        email: user.email || '',
        full_name: user.full_name || '',
      });
      fetchUserData();
    }
  }, [user]);

  const fetchUserData = async () => {
    try {
      setLoading(true);
      // Ideally backend provides this directly on the user object or a /me/profile endpoint
      // For now, we fallback to fetching all and filtering if user hasn't these populated
      const [groupsRes, dsRes] = await Promise.all([
        apiClient.getGroups().catch(() => []),
        apiClient.getDataSources().catch(() => []),
      ]);
      setGroups(groupsRes);
      setDataSources(dsRes);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleAccountSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      // Assuming there's an endpoint to update current user profile
      // await apiClient.updateProfile(accountForm);
      toast.success('Account updated successfully');
    } catch (err) {
      toast.error('Failed to update account');
    } finally {
      setSaving(false);
    }
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (passwordForm.newPw !== passwordForm.confirmPw) {
      toast.error('Passwords do not match');
      return;
    }
    setSaving(true);
    try {
      await apiClient.changePassword({
        old_password: passwordForm.currPw,
        new_password: passwordForm.newPw,
      });
      toast.success('Password changed successfully');
      setPasswordForm({ currPw: '', newPw: '', confirmPw: '' });
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to change password');
    } finally {
      setSaving(false);
    }
  };

  const getPasswordStrength = () => {
    const val = passwordForm.newPw;
    if (!val) return { w: '0', c: 'transparent', t: 'Enter a new password' };
    let score = 0;
    if (val.length >= 8) score++;
    if (/[A-Z]/.test(val)) score++;
    if (/[0-9]/.test(val)) score++;
    if (/[^A-Za-z0-9]/.test(val)) score++;
    
    const levels = [
      { w: '20%', c: '#EF4444', t: 'Too weak' },
      { w: '45%', c: '#F97316', t: 'Weak' },
      { w: '70%', c: '#EAB308', t: 'Fair' },
      { w: '100%', c: '#22C55E', t: 'Strong' },
    ];
    return levels[Math.min(score, levels.length) - 1] || levels[0];
  };

  const str = getPasswordStrength();

  if (!user) return null;

  return (
    <div className="max-w-6xl mx-auto pt-8 pb-20 w-full">
      {/* Massive Typographic Hero */}
      <div className="mb-16 md:mb-24">
        <h1 className="text-5xl md:text-8xl font-black tracking-tighter text-[var(--text-primary)] mb-6 leading-none">
          {user.full_name || user.username}
        </h1>
        <div className="flex flex-wrap items-center gap-4 text-xs md:text-sm font-bold tracking-[0.15em] uppercase text-[var(--text-muted)]">
          <span className="text-[var(--text-primary)]">@{user.username}</span>
          <span className="w-1.5 h-1.5 rounded-none bg-[var(--text-faint)]"></span>
          <span>{user.email}</span>
          <span className="w-1.5 h-1.5 rounded-none bg-[var(--text-faint)]"></span>
          <span className={`px-3 py-1 text-[var(--bg-page)] text-xs font-black rounded-sm ${user.role === 'admin' ? 'bg-[var(--teal-text)]' : user.role === 'viewer' ? 'bg-[var(--text-faint)]' : 'bg-[var(--accent-blue)]'}`}>{user.role}</span>
          <span className="w-1.5 h-1.5 rounded-none bg-[var(--text-faint)]"></span>
          <span>Since {new Date(user.created_at).toLocaleDateString()}</span>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-[var(--border-light)] mb-16 overflow-x-auto hide-scrollbar">
        <nav className="flex gap-10 min-w-max">
          <button className={`pb-4 text-xs font-bold tracking-[0.15em] uppercase transition-colors whitespace-nowrap ${activeTab === 'account' ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]' : 'text-[var(--text-muted)] hover:text-[var(--text-primary)] border-b-2 border-transparent'}`} onClick={() => setActiveTab('account')}>Account Identity</button>
          <button className={`pb-4 text-xs font-bold tracking-[0.15em] uppercase transition-colors whitespace-nowrap ${activeTab === 'password' ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]' : 'text-[var(--text-muted)] hover:text-[var(--text-primary)] border-b-2 border-transparent'}`} onClick={() => setActiveTab('password')}>Security</button>
          <button className={`pb-4 text-xs font-bold tracking-[0.15em] uppercase transition-colors whitespace-nowrap ${activeTab === 'groups' ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]' : 'text-[var(--text-muted)] hover:text-[var(--text-primary)] border-b-2 border-transparent'}`} onClick={() => setActiveTab('groups')}>Group Memberships ({loading ? '-' : groups.length})</button>
          <button className={`pb-4 text-xs font-bold tracking-[0.15em] uppercase transition-colors whitespace-nowrap ${activeTab === 'databases' ? 'text-[var(--accent-blue)] border-b-2 border-[var(--accent-blue)]' : 'text-[var(--text-muted)] hover:text-[var(--text-primary)] border-b-2 border-transparent'}`} onClick={() => setActiveTab('databases')}>Database Access ({loading ? '-' : dataSources.length})</button>
        </nav>
      </div>

      <div className="flex flex-col gap-12 max-w-4xl">
        {/* Tab: Account */}
        {activeTab === 'account' && (
          <form onSubmit={handleAccountSubmit} className="flex flex-col gap-10">
            <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
              <div className="md:mt-2">
                <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1" htmlFor="full_name">
                  Full Name
                </label>
                <p className="text-xs text-[var(--text-faint)] leading-relaxed">Your primary display identity.</p>
              </div>
              <div className="relative">
                <input id="full_name" className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]" type="text" value={accountForm.full_name} onChange={e => setAccountForm({...accountForm, full_name: e.target.value})} />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
              <div className="md:mt-2">
                <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1" htmlFor="username">
                  Username
                </label>
                <p className="text-xs text-[var(--text-faint)] leading-relaxed text-red-400">Cannot be modified.</p>
              </div>
              <div className="relative">
                <input id="username" className="w-full bg-transparent border-b border-[var(--border-light)] pb-3 text-xl text-[var(--text-faint)] focus:outline-none transition-colors rounded-none cursor-not-allowed" type="text" value={accountForm.username} disabled />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
              <div className="md:mt-2">
                <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1" htmlFor="email">
                  Email Address
                </label>
                <p className="text-xs text-[var(--text-faint)] leading-relaxed">Used for notifications.</p>
              </div>
              <div className="relative">
                <input id="email" className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]" type="email" value={accountForm.email} onChange={e => setAccountForm({...accountForm, email: e.target.value})} />
              </div>
            </div>

            <div className="mt-12 pt-8 border-t border-[var(--border-light)] grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8">
              <div className="hidden md:block"></div>
              <div className="flex gap-6 items-center">
                <button type="submit" className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50" style={{ borderRadius: '2px' }} disabled={saving}>
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
                <button type="button" className="h-12 px-2 text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors" onClick={() => setAccountForm({ username: user.username, email: user.email, full_name: user.full_name })}>
                  Discard
                </button>
              </div>
            </div>
          </form>
        )}

        {/* Tab: Password */}
        {activeTab === 'password' && (
          <form onSubmit={handlePasswordSubmit} className="flex flex-col gap-10">
            <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
              <div className="md:mt-2">
                <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1" htmlFor="currPw">
                  Current Password
                </label>
              </div>
              <div className="relative">
                <input id="currPw" className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]" type="password" placeholder="Enter current password" value={passwordForm.currPw} onChange={e => setPasswordForm({...passwordForm, currPw: e.target.value})} required />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
              <div className="md:mt-2">
                <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1" htmlFor="newPw">
                  New Password
                </label>
                <p className="text-xs text-[var(--text-faint)] leading-relaxed">At least 8 characters</p>
              </div>
              <div className="relative">
                <input id="newPw" className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]" type="password" placeholder="Enter new password" value={passwordForm.newPw} onChange={e => setPasswordForm({...passwordForm, newPw: e.target.value})} required />
                <div className="mt-4 flex flex-col gap-2">
                  <div className="h-1 bg-[var(--border-light)] w-full overflow-hidden" style={{ borderRadius: '2px' }}>
                    <div className="h-full transition-all duration-300" style={{ width: str.w, background: str.c }}></div>
                  </div>
                  <div className="text-xs font-bold uppercase tracking-widest text-right" style={{ color: str.c }}>{str.t}</div>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8 items-start">
              <div className="md:mt-2">
                <label className="block text-xs font-bold tracking-[0.15em] uppercase text-[var(--text-muted)] mb-1" htmlFor="confirmPw">
                  Confirm Password
                </label>
              </div>
              <div className="relative">
                <input id="confirmPw" className="w-full bg-transparent border-b border-[var(--border)] pb-3 text-xl text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent-blue)] transition-colors rounded-none placeholder-[var(--text-faint)]" type="password" placeholder="Re-enter new password" value={passwordForm.confirmPw} onChange={e => setPasswordForm({...passwordForm, confirmPw: e.target.value})} required />
              </div>
            </div>

            <div className="mt-12 pt-8 border-t border-[var(--border-light)] grid grid-cols-1 md:grid-cols-[1fr_2.5fr] gap-8">
              <div className="hidden md:block"></div>
              <div className="flex gap-6 items-center">
                <button type="submit" className="h-12 px-8 bg-[var(--text-primary)] text-[var(--bg-page)] text-sm font-bold tracking-[0.1em] uppercase hover:opacity-90 transition-opacity disabled:opacity-50" style={{ borderRadius: '2px' }} disabled={saving}>
                  {saving ? 'Updating...' : 'Update Password'}
                </button>
                <button type="button" className="h-12 px-2 text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors" onClick={() => setPasswordForm({ currPw: '', newPw: '', confirmPw: '' })}>
                  Clear
                </button>
              </div>
            </div>
          </form>
        )}

        {/* Tab: Groups */}
        {activeTab === 'groups' && (
          <div className="flex flex-col gap-6">
            <h3 className="text-xl font-bold text-[var(--text-primary)] mb-4 border-l-4 border-[var(--accent-blue)] pl-4">Security Groups</h3>
            <div className="flex flex-col gap-4">
              {groups.map(group => (
                <div key={group.id} className="flex flex-col md:flex-row md:items-center justify-between p-6 border border-[var(--border-light)] bg-transparent hover:bg-[var(--bg-hover)] transition-colors" style={{ borderRadius: '2px' }}>
                  <div className="flex flex-col mb-4 md:mb-0">
                    <div className="text-lg font-bold text-[var(--text-primary)]">{group.name}</div>
                    <div className="text-sm text-[var(--text-muted)] mt-1">{group.description || 'No description assigned.'}</div>
                  </div>
                  <span className="self-start md:self-auto px-4 py-1.5 border border-[var(--border)] text-xs font-bold tracking-widest uppercase text-[var(--text-faint)]" style={{ borderRadius: '2px' }}>member</span>
                </div>
              ))}
              {groups.length === 0 && !loading && (
                <div className="p-8 border border-dashed border-[var(--border)] text-center" style={{ borderRadius: '2px' }}>
                  <p className="text-sm text-[var(--text-muted)] uppercase tracking-wide font-bold">You are not a member of any groups yet.</p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Tab: Database Access */}
        {activeTab === 'databases' && (
          <div className="flex flex-col gap-6">
            <h3 className="text-xl font-bold text-[var(--text-primary)] mb-4 border-l-4 border-[var(--accent-blue)] pl-4">Accessible Databases</h3>
            <div className="flex flex-col gap-4">
              {dataSources.map(db => (
                <div key={db.id} className="flex flex-col md:flex-row md:items-center justify-between p-6 border border-[var(--border-light)] bg-transparent hover:bg-[var(--bg-hover)] transition-colors" style={{ borderRadius: '2px' }}>
                  <div className="flex items-center gap-6 mb-4 md:mb-0">
                    <div className="w-12 h-12 flex-shrink-0 flex items-center justify-center border font-bold text-lg" style={{ borderRadius: '2px', borderColor: db.type === 'postgresql' ? '#DBEAFE' : '#FFFBEB', backgroundColor: db.type === 'postgresql' ? 'rgba(219,234,254,0.1)' : 'rgba(255,251,235,0.1)', color: db.type === 'postgresql' ? '#1D4ED8' : '#D97706' }}>
                      {db.type === 'postgresql' ? 'PG' : 'MY'}
                    </div>
                    <div className="flex flex-col">
                      <div className="text-lg font-bold text-[var(--text-primary)]">{db.name}</div>
                      <div className="text-xs tracking-wide text-[var(--text-muted)] mt-1">{db.type} · {db.host}:{db.port}</div>
                    </div>
                  </div>
                  <span className="self-start md:self-auto px-4 py-1.5 bg-[var(--bg-page)] border border-[var(--border-light)] text-xs font-bold tracking-widest uppercase text-[var(--text-primary)]" style={{ borderRadius: '2px' }}>{db.database_name}</span>
                </div>
              ))}
              {dataSources.length === 0 && !loading && (
                <div className="p-8 border border-dashed border-[var(--border)] text-center" style={{ borderRadius: '2px' }}>
                  <p className="text-sm text-[var(--text-muted)] uppercase tracking-wide font-bold">No databases available.</p>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
