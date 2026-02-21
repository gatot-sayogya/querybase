'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import PageTransition from '@/components/layout/PageTransition';

export default function LoginPage() {
  const router = useRouter();
  const { login, isLoading, error } = useAuthStore();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(username, password);
      router.push('/dashboard');
    } catch (err) {
      // Error is handled by the store
    }
  };

  return (
    <PageTransition>
      <div className="min-h-screen flex items-center justify-center p-6 relative overflow-hidden bg-slate-50 dark:bg-slate-900">
        {/* Subtle background glow from login.html body::before */}
        <div style={{
          position: 'fixed', top: '-120px', left: '50%', transform: 'translateX(-20%)',
          width: '640px', height: '640px', background: 'radial-gradient(circle, rgba(37, 99, 235, 0.25) 0%, transparent 70%)',
          pointerEvents: 'none'
        }} />

        <div className="login-card">
          {/* Logo */}
          <div className="logo-row">
            <div className="logo-mark">Q</div>
            <span className="logo-name">QueryBase</span>
          </div>

          {/* Title */}
          <div className="title-block">
            <h1>Welcome back</h1>
            <p>Sign in to your account to continue</p>
          </div>

          {/* Form */}
          <form className="form-fields" onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="form-label" htmlFor="username">Username</label>
              <input
                id="username"
                name="username"
                className="form-input"
                type="text"
                required
                placeholder="Enter username"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div className="form-group">
              <label className="form-label" htmlFor="password">Password</label>
              <input
                id="password"
                name="password"
                className="form-input"
                type="password"
                required
                placeholder="••••••••"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            {/* Error message */}
            {error && (
              <div className="error-msg show" id="errorMsg">
                ⚠ {error}
              </div>
            )}

            {/* CTA */}
            <button 
              type="submit" 
              className="signin-btn" 
              id="signInBtn"
              disabled={isLoading}
            >
              {isLoading ? 'Signing in...' : 'Sign in'}
            </button>
            <div className="mt-2 text-center">
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Demo Credentials: <span className="font-mono">admin / admin123</span>
              </p>
            </div>
          </form>

          {/* Footer */}
          <p className="card-footer">Secure database access for your team</p>
        </div>
      </div>
    </PageTransition>
  );
}
