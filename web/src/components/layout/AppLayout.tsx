'use client';

import { usePathname, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/stores/auth-store';
import { useState, useRef, useEffect } from 'react';
import { useThemeStore } from '@/stores/theme-store';

interface AppLayoutProps {
  children: React.ReactNode;
}

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const { theme } = useThemeStore();

  const handleLogout = async () => {
    await logout();
    router.push('/login');
  };

  const isAdmin = user?.role === 'admin';

  return (
    <div className="app-shell">
      {/* Sidebar */}
      <nav className="sidebar">
        <div className="sidebar-logo">
          <Link href="/dashboard" className="sidebar-mark" style={{ textDecoration: 'none' }}>Q</Link>
        </div>
        
        <div className="sidebar-nav">
          <Link 
            href="/dashboard/query" 
            className={`nav-item ${pathname === '/dashboard/query' ? 'active' : ''}`} 
            data-tooltip="Query Editor"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
          </Link>
          <Link 
            href="/dashboard/history" 
            className={`nav-item ${pathname === '/dashboard/history' ? 'active' : ''}`} 
            data-tooltip="History"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </Link>
          <Link 
            href="/dashboard/approvals" 
            className={`nav-item ${pathname === '/dashboard/approvals' ? 'active' : ''}`} 
            data-tooltip="Approvals"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
          </Link>
        </div>
        
        <div className="sidebar-bottom">
          {isAdmin && (
            <div className="settings-wrapper" id="settingsWrapper">
              <button className={`nav-item ${pathname?.startsWith('/admin') ? 'active' : ''}`} id="settingsBtn">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
              </button>
              <div className="settings-menu" id="settingsMenu">
                <div className="settings-menu-label">Settings</div>
                <Link href="/admin/users" className={`settings-item ${pathname === '/admin/users' ? 'active-page' : ''}`}>
                  <span className="settings-icon" style={{background:'#EFF6FF',color:'#2563EB'}}><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg></span>
                  Users
                </Link>
                <Link href="/admin/groups" className={`settings-item ${pathname === '/admin/groups' ? 'active-page' : ''}`}>
                  <span className="settings-icon" style={{background:'#F5F3FF',color:'#7C3AED'}}><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg></span>
                  Groups
                </Link>
                <Link href="/admin/datasources" className={`settings-item ${pathname === '/admin/datasources' ? 'active-page' : ''}`}>
                  <span className="settings-icon" style={{background:'#F0FDF4',color:'#059669'}}><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg></span>
                  Data Sources
                </Link>
              </div>
            </div>
          )}

          <div className="avatar-wrapper">
            <button 
              className="nav-item avatar-btn" 
              title="Account"
            >
              {(user?.full_name || user?.username || 'A').charAt(0).toUpperCase()}
            </button>
            <div className="avatar-menu">
              <div className="menu-user-info">
                <div className="menu-user-name">{user?.full_name || user?.username}</div>
                <div className="menu-user-role">{user?.role}</div>
              </div>
              <Link href="/profile" className="menu-item">
                <span className="menu-icon" style={{background:'#EFF6FF',color:'#2563EB'}}>👤</span>
                My Profile
              </Link>
              <Link href="/dashboard/query" className="menu-item">
                <span className="menu-icon" style={{background:'#F0FDF4',color:'#059669'}}>⌕</span>
                New Query
              </Link>
              <div className="menu-divider"></div>
              
              <button 
                className="menu-item w-full flex justify-between items-center" 
                onClick={(e) => {
                  e.stopPropagation();
                  const nextTheme = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light';
                  useThemeStore.getState().setTheme(nextTheme);
                }}
              >
                <div className="flex items-center gap-2">
                  <span className="menu-icon" style={{background:'#F3F4F6',color:'#374151', ...(theme === 'dark' ? {background:'#374151', color:'#F3F4F6'} : {})}}>
                    {theme === 'light' ? '☀️' : theme === 'dark' ? '🌙' : '💻'}
                  </span>
                  Theme
                </div>
                <span className="text-xs text-gray-500 capitalize">{theme}</span>
              </button>
              
              <div className="menu-divider"></div>
              <button className="menu-item danger" onClick={() => { handleLogout(); }}>
                <span className="menu-icon">🚪</span>
                Sign out
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main */}
      <main className={pathname === '/dashboard/query' ? "flex-1 flex flex-col h-[100vh] overflow-hidden" : "main-content"}>
        {children}
      </main>
    </div>
  );
}
