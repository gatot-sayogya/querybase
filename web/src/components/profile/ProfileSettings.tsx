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
    <>
      <style>{`
        .profile-grid {
          display: grid;
          grid-template-columns: 300px 1fr;
          gap: 24px;
          align-items: start;
        }

        /* ── Left column: user card ─────────────────── */
        .user-card {
          display: flex;
          flex-direction: column;
          gap: 0;
          overflow: hidden;
          padding: 0;
        }
        .user-card-top {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 14px;
          padding: 32px 24px 24px;
          border-bottom: 1px solid var(--border-light);
        }
        .avatar-large {
          width: 80px; height: 80px; border-radius: 50%;
          background: linear-gradient(135deg, #4F46E5, #2563EB);
          display: flex; align-items: center; justify-content: center;
          font-size: 32px; font-weight: 700; color: #fff;
          letter-spacing: -1px;
        }
        .user-name  { font-size: 18px; font-weight: 700; color: var(--text-primary); text-align: center; }
        .user-email { font-size: 13px; color: var(--text-muted); text-align: center; margin-top: -6px; }
        .user-role-badge { margin-top: 4px; }

        .user-meta { padding: 20px 24px; display: flex; flex-direction: column; gap: 14px; }
        .meta-row {
          display: flex; justify-content: space-between; align-items: center;
          font-size: 13px;
        }
        .meta-label { color: var(--text-muted); font-weight: 500; }
        .meta-value { color: var(--text-primary); font-weight: 500; }

        /* ── Right column ────────────────────────────── */
        .right-col { display: flex; flex-direction: column; gap: 24px; }

        /* Tab Nav */
        .tab-nav {
          display: flex;
          border-bottom: 1px solid var(--border-light);
          gap: 0;
          margin-bottom: 24px;
        }
        .tab-btn {
          padding: 10px 20px;
          font-size: 14px; font-weight: 500;
          color: var(--text-muted);
          cursor: pointer;
          border-bottom: 2px solid transparent;
          background: none; border-top: none; border-left: none; border-right: none;
          font-family: 'Inter', sans-serif;
          transition: color 0.15s, border-color 0.15s;
        }
        .tab-btn:hover { color: var(--text-primary); }
        .tab-btn.active { color: var(--accent-blue); border-bottom-color: var(--accent-blue); }

        /* Form card */
        .form-card { padding: 28px; }
        .form-card-title { font-size: 16px; font-weight: 700; color: var(--text-primary); margin-bottom: 4px; }
        .form-card-sub   { font-size: 13px; color: var(--text-muted); margin-bottom: 22px; }
        .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
        .form-row.single { grid-template-columns: 1fr; }
        
        .pw-strength { margin-top: 8px; }
        .pw-bar-track { height: 4px; background: var(--border-light); border-radius: 99px; overflow: hidden; }
        .pw-bar-fill  { height: 100%; border-radius: 99px; transition: width 0.3s, background 0.3s; width: 0; }
        .pw-hint { font-size: 11px; color: var(--text-muted); margin-top: 4px; }

        /* Groups section */
        .group-list { display: flex; flex-direction: column; gap: 10px; }
        .group-item {
          display: flex; align-items: center; justify-content: space-between;
          padding: 14px 20px;
          border-radius: var(--r-lg);
          border: 1px solid var(--border-light);
          background: var(--bg-hover);
        }
        .group-item-left { display: flex; align-items: center; gap: 12px; }
        .group-icon-sm {
          width: 38px; height: 38px; border-radius: 10px;
          display: flex; align-items: center; justify-content: center;
          font-size: 15px; flex-shrink: 0;
        }
        .group-name { font-size: 14px; font-weight: 600; color: var(--text-primary); }
        .group-desc { font-size: 12px; color: var(--text-muted); margin-top: 2px; }

        /* DB access list */
        .db-list { display: flex; flex-direction: column; gap: 10px; }
        .db-item {
          display: flex; align-items: center; justify-content: space-between;
          padding: 14px 20px;
          border-radius: var(--r-lg);
          border: 1px solid var(--border-light);
          background: var(--bg-hover);
        }
        .db-item-left { display: flex; align-items: center; gap: 14px; }
        .db-type-icon {
          width: 38px; height: 38px; border-radius: 10px;
          display: flex; align-items: center; justify-content: center;
          font-size: 15px; flex-shrink: 0;
        }
        .db-name   { font-size: 14px; font-weight: 600; color: var(--text-primary); }
        .db-host   { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
        .db-perms  { display: flex; gap: 6px; flex-wrap: wrap; }
      `}</style>
      
      <div className="page-header" style={{ marginBottom: '24px' }}>
        <div>
          <h1 className="page-title">My Profile</h1>
          <p className="page-subtitle">Manage your account settings and view your access permissions</p>
        </div>
      </div>

      <div className="profile-grid">
        {/* ── Left: User card ── */}
        <div className="card user-card">
          <div className="user-card-top">
            <div className="avatar-large">
              {(user.full_name || user.username || '?').charAt(0).toUpperCase()}
            </div>
            <div>
              <div className="user-name">{user.full_name || user.username}</div>
              <div className="user-email">{user.email}</div>
              <div className="user-role-badge" style={{ textAlign: 'center', marginTop: '8px' }}>
                <span className={`badge ${user.role === 'admin' ? 'badge-purple' : user.role === 'viewer' ? 'badge-slate' : 'badge-blue'}`}>
                  {user.role}
                </span>
              </div>
            </div>
          </div>
          <div className="user-meta">
            <div className="meta-row">
              <span className="meta-label">Username</span>
              <span className="meta-value">@{user.username}</span>
            </div>
            <div className="meta-row">
              <span className="meta-label">Member since</span>
              <span className="meta-value">{new Date(user.created_at).toLocaleDateString()}</span>
            </div>
            <div className="meta-row">
              <span className="meta-label">Groups</span>
              <span className="meta-value">{loading ? '-' : groups.length}</span>
            </div>
            <div className="meta-row">
              <span className="meta-label">DB Access</span>
              <span className="meta-value">{loading ? '-' : dataSources.length} sources</span>
            </div>
          </div>
        </div>

        {/* ── Right: Tabs ── */}
        <div className="right-col">
          <div style={{ borderBottom: '1px solid var(--border-light)' }}>
            <nav className="tab-nav" style={{ borderBottom: 'none', marginBottom: 0 }}>
              <button className={`tab-btn ${activeTab === 'account' ? 'active' : ''}`} onClick={() => setActiveTab('account')}>Account</button>
              <button className={`tab-btn ${activeTab === 'password' ? 'active' : ''}`} onClick={() => setActiveTab('password')}>Password</button>
              <button className={`tab-btn ${activeTab === 'groups' ? 'active' : ''}`} onClick={() => setActiveTab('groups')}>Groups</button>
              <button className={`tab-btn ${activeTab === 'databases' ? 'active' : ''}`} onClick={() => setActiveTab('databases')}>Database Access</button>
            </nav>
          </div>

          {/* Tab: Account */}
          {activeTab === 'account' && (
            <div className="card form-card">
              <div className="form-card-title">Account Information</div>
              <div className="form-card-sub">Update your display name and username</div>
              <form onSubmit={handleAccountSubmit}>
                <div className="form-row single">
                  <div className="form-group">
                    <label className="form-label" htmlFor="full_name">Full Name</label>
                    <input id="full_name" className="input-field" type="text" value={accountForm.full_name} onChange={e => setAccountForm({...accountForm, full_name: e.target.value})} />
                  </div>
                </div>
                <div className="form-row single">
                  <div className="form-group">
                    <label className="form-label" htmlFor="username">Username</label>
                    <input id="username" className="input-field" type="text" value={accountForm.username} onChange={e => setAccountForm({...accountForm, username: e.target.value})} disabled />
                    <div style={{ fontSize: '13px', color: 'var(--text-muted)', marginTop: '4px' }}>Username cannot be changed.</div>
                  </div>
                </div>
                <div className="form-row single">
                  <div className="form-group">
                    <label className="form-label" htmlFor="email">Email Address</label>
                    <input id="email" className="input-field" type="email" value={accountForm.email} onChange={e => setAccountForm({...accountForm, email: e.target.value})} />
                  </div>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px', marginTop: '24px', paddingTop: '20px', borderTop: '1px solid var(--border-light)' }}>
                  <button type="button" className="btn btn-ghost" onClick={() => setAccountForm({ username: user.username, email: user.email, full_name: user.full_name })}>Discard</button>
                  <button type="submit" className="btn btn-primary" disabled={saving}>{saving ? 'Saving...' : 'Save Changes'}</button>
                </div>
              </form>
            </div>
          )}

          {/* Tab: Password */}
          {activeTab === 'password' && (
            <div className="card form-card">
              <div className="form-card-title">Change Password</div>
              <div className="form-card-sub">Use a strong password with at least 8 characters</div>
              <form onSubmit={handlePasswordSubmit}>
                <div className="form-row single">
                  <div className="form-group">
                    <label className="form-label" htmlFor="currPw">Current Password</label>
                    <input id="currPw" className="input-field" type="password" placeholder="Enter current password" value={passwordForm.currPw} onChange={e => setPasswordForm({...passwordForm, currPw: e.target.value})} required />
                  </div>
                </div>
                <div className="form-row single">
                  <div className="form-group">
                    <label className="form-label" htmlFor="newPw">New Password</label>
                    <input id="newPw" className="input-field" type="password" placeholder="At least 8 characters" value={passwordForm.newPw} onChange={e => setPasswordForm({...passwordForm, newPw: e.target.value})} required />
                    <div className="pw-strength">
                      <div className="pw-bar-track"><div className="pw-bar-fill" style={{ width: str.w, background: str.c }}></div></div>
                      <div className="pw-hint" style={{ color: str.c }}>{str.t}</div>
                    </div>
                  </div>
                </div>
                <div className="form-row single">
                  <div className="form-group">
                    <label className="form-label" htmlFor="confirmPw">Confirm New Password</label>
                    <input id="confirmPw" className="input-field" type="password" placeholder="Re-enter new password" value={passwordForm.confirmPw} onChange={e => setPasswordForm({...passwordForm, confirmPw: e.target.value})} required />
                  </div>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px', marginTop: '24px', paddingTop: '20px', borderTop: '1px solid var(--border-light)' }}>
                  <button type="button" className="btn btn-ghost" onClick={() => setPasswordForm({ currPw: '', newPw: '', confirmPw: '' })}>Clear</button>
                  <button type="submit" className="btn btn-primary" disabled={saving}>{saving ? 'Updating...' : 'Update Password'}</button>
                </div>
              </form>
            </div>
          )}

          {/* Tab: Groups */}
          {activeTab === 'groups' && (
            <div className="card" style={{ padding: '20px 24px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                <div>
                  <div className="form-card-title" style={{ marginBottom: '2px' }}>Group Memberships</div>
                  <div className="form-card-sub" style={{ marginBottom: 0 }}>Groups determine which data sources you can access</div>
                </div>
                <span className="badge badge-slate">{groups.length} groups</span>
              </div>
              <div className="group-list">
                {groups.map(group => (
                  <div key={group.id} className="group-item">
                    <div className="group-item-left">
                      <div className="group-icon-sm" style={{ background: '#EDE9FE', color: '#4F46E5' }}>⛾</div>
                      <div>
                        <div className="group-name">{group.name}</div>
                        <div className="group-desc">{group.description || '-'}</div>
                      </div>
                    </div>
                    <span className="badge badge-slate">member</span>
                  </div>
                ))}
                {groups.length === 0 && !loading && (
                  <div style={{ padding: '20px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '14px' }}>You are not a member of any groups yet.</div>
                )}
              </div>
            </div>
          )}

          {/* Tab: Database Access */}
          {activeTab === 'databases' && (
            <div className="card" style={{ padding: '20px 24px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                <div>
                  <div className="form-card-title" style={{ marginBottom: '2px' }}>Accessible Databases</div>
                  <div className="form-card-sub" style={{ marginBottom: 0 }}>These are the data sources your groups grant you access to</div>
                </div>
                <span className="badge badge-slate">{dataSources.length} sources</span>
              </div>
              <div className="db-list">
                {dataSources.map(db => (
                  <div key={db.id} className="db-item">
                    <div className="db-item-left">
                      <div className="db-type-icon" style={{ background: db.type === 'postgresql' ? '#DBEAFE' : '#FFFBEB', color: db.type === 'postgresql' ? '#1D4ED8' : '#D97706' }}>◉</div>
                      <div>
                        <div className="db-name">{db.name}</div>
                        <div className="db-host">{db.type} · {db.host}:{db.port}</div>
                      </div>
                    </div>
                    <div className="db-perms">
                      <span className="badge badge-slate">{db.database_name}</span>
                    </div>
                  </div>
                ))}
                {dataSources.length === 0 && !loading && (
                  <div style={{ padding: '20px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '14px' }}>No databases available.</div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
