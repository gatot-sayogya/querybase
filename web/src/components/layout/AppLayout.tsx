'use client';

import { usePathname, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/stores/auth-store';
import { useState, useRef, useEffect } from 'react';
import { useThemeStore } from '@/stores/theme-store';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { scaleIn, springConfig, duration, reducedMotionVariants } from '@/lib/animations';
import {
  UserIcon,
  CodeBracketIcon,
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon,
  ArrowRightOnRectangleIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/outline';

interface AppLayoutProps {
  children: React.ReactNode;
}

const navItemVariants = {
  initial: { scale: 1 },
  hover: { scale: 1.05 },
  tap: { scale: 0.95 },
};

const menuVariants = {
  initial: { opacity: 0, scale: 0.95, y: 8 },
  animate: { opacity: 1, scale: 1, y: 0 },
  exit: { opacity: 0, scale: 0.95, y: 8 },
};

const menuItemVariants = {
  initial: { opacity: 0, x: -8 },
  animate: { opacity: 1, x: 0 },
};

export default function AppLayout({ children }: AppLayoutProps) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const { theme } = useThemeStore();
  const shouldReduceMotion = useReducedMotion();

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [avatarOpen, setAvatarOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
    router.push('/login');
  };

  const isAdmin = user?.role === 'admin';

  const ThemeIcon = theme === 'light' ? SunIcon : theme === 'dark' ? MoonIcon : ComputerDesktopIcon;

  const menuTransition = shouldReduceMotion
    ? { duration: duration.fast }
    : { ...springConfig.snappy, duration: duration.normal };

  return (
    <div className="app-shell">
      <nav className="sidebar">
        <div className="sidebar-logo">
          <Link href="/dashboard" className="sidebar-mark" style={{ textDecoration: 'none' }}>
            Q
          </Link>
        </div>

        <div className="sidebar-nav">
          <motion.div
            variants={navItemVariants}
            whileHover={shouldReduceMotion ? {} : 'hover'}
            whileTap={shouldReduceMotion ? {} : 'tap'}
            transition={springConfig.snappy}
          >
            <Link
              href="/dashboard/query"
              className={`nav-item ${pathname === '/dashboard/query' ? 'active' : ''}`}
              data-tooltip="Query Editor"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="16 18 22 12 16 6" />
                <polyline points="8 6 2 12 8 18" />
              </svg>
            </Link>
          </motion.div>

          <motion.div
            variants={navItemVariants}
            whileHover={shouldReduceMotion ? {} : 'hover'}
            whileTap={shouldReduceMotion ? {} : 'tap'}
            transition={springConfig.snappy}
          >
            <Link
              href="/dashboard/history"
              className={`nav-item ${pathname === '/dashboard/history' ? 'active' : ''}`}
              data-tooltip="History"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
            </Link>
          </motion.div>

          <motion.div
            variants={navItemVariants}
            whileHover={shouldReduceMotion ? {} : 'hover'}
            whileTap={shouldReduceMotion ? {} : 'tap'}
            transition={springConfig.snappy}
          >
            <Link
              href="/dashboard/approvals"
              className={`nav-item ${pathname === '/dashboard/approvals' ? 'active' : ''}`}
              data-tooltip="Approvals"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </svg>
            </Link>
          </motion.div>
        </div>

        <div className="sidebar-bottom">
          {isAdmin && (
            <div
              className="settings-wrapper"
              onMouseEnter={() => setSettingsOpen(true)}
              onMouseLeave={() => setSettingsOpen(false)}
            >
              <motion.button
                className={`nav-item ${pathname?.startsWith('/admin') ? 'active' : ''}`}
                variants={navItemVariants}
                whileHover={shouldReduceMotion ? {} : 'hover'}
                whileTap={shouldReduceMotion ? {} : 'tap'}
                transition={springConfig.snappy}
              >
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="3" />
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
                </svg>
              </motion.button>

              <AnimatePresence>
                {settingsOpen && (
                  <motion.div
                    className="settings-menu"
                    variants={shouldReduceMotion ? reducedMotionVariants : menuVariants}
                    initial="initial"
                    animate="animate"
                    exit="exit"
                    transition={menuTransition}
                  >
                    <div className="settings-menu-label">Settings</div>
                    <motion.div
                      variants={{
                        initial: {},
                        animate: { transition: { staggerChildren: 0.05 } },
                      }}
                      initial="initial"
                      animate="animate"
                    >
                      <motion.div variants={menuItemVariants} transition={menuTransition}>
                        <Link href="/admin/users" className={`settings-item ${pathname === '/admin/users' ? 'active-page' : ''}`}>
                          <span className="settings-icon" style={{ background: '#EFF6FF', color: '#2563EB' }}>
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                              <circle cx="9" cy="7" r="4" />
                              <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                              <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                            </svg>
                          </span>
                          Users
                        </Link>
                      </motion.div>
                      <motion.div variants={menuItemVariants} transition={menuTransition}>
                        <Link href="/admin/groups" className={`settings-item ${pathname === '/admin/groups' ? 'active-page' : ''}`}>
                          <span className="settings-icon" style={{ background: '#ccfbf1', color: '#0d9488' }}>
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <rect x="3" y="3" width="7" height="7" rx="1" />
                              <rect x="14" y="3" width="7" height="7" rx="1" />
                              <rect x="3" y="14" width="7" height="7" rx="1" />
                              <rect x="14" y="14" width="7" height="7" rx="1" />
                            </svg>
                          </span>
                          Groups
                        </Link>
                      </motion.div>
                      <motion.div variants={menuItemVariants} transition={menuTransition}>
                        <Link href="/admin/datasources" className={`settings-item ${pathname === '/admin/datasources' ? 'active-page' : ''}`}>
                          <span className="settings-icon" style={{ background: '#F0FDF4', color: '#059669' }}>
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <ellipse cx="12" cy="5" rx="9" ry="3" />
                              <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
                              <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
                            </svg>
                          </span>
                          Data Sources
                        </Link>
                      </motion.div>
                    </motion.div>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          )}

          <div
            className="avatar-wrapper"
            onMouseEnter={() => setAvatarOpen(true)}
            onMouseLeave={() => setAvatarOpen(false)}
          >
            <motion.button
              className="nav-item avatar-btn"
              title="Account"
              variants={navItemVariants}
              whileHover={shouldReduceMotion ? {} : 'hover'}
              whileTap={shouldReduceMotion ? {} : 'tap'}
              transition={springConfig.snappy}
            >
              {(user?.full_name || user?.username || 'A').charAt(0).toUpperCase()}
            </motion.button>

            <AnimatePresence>
              {avatarOpen && (
                <motion.div
                  className="avatar-menu"
                  variants={shouldReduceMotion ? reducedMotionVariants : menuVariants}
                  initial="initial"
                  animate="animate"
                  exit="exit"
                  transition={menuTransition}
                >
                  <div className="menu-user-info">
                    <div className="menu-user-name">{user?.full_name || user?.username}</div>
                    <div className="menu-user-role">{user?.role}</div>
                  </div>

                  <motion.div
                    variants={{
                      initial: {},
                      animate: { transition: { staggerChildren: 0.04 } },
                    }}
                    initial="initial"
                    animate="animate"
                  >
                    <motion.div variants={menuItemVariants} transition={menuTransition}>
                      <Link href="/profile" className="menu-item">
                        <span className="menu-icon" style={{ background: '#EFF6FF', color: '#2563EB' }}>
                          <UserIcon className="w-4 h-4" />
                        </span>
                        My Profile
                      </Link>
                    </motion.div>
                    <motion.div variants={menuItemVariants} transition={menuTransition}>
                      <Link href="/dashboard/query" className="menu-item">
                        <span className="menu-icon" style={{ background: '#F0FDF4', color: '#059669' }}>
                          <MagnifyingGlassIcon className="w-4 h-4" />
                        </span>
                        New Query
                      </Link>
                    </motion.div>
                    <div className="menu-divider" />
                    <motion.div variants={menuItemVariants} transition={menuTransition}>
                      <button
                        className="menu-item w-full flex justify-between items-center"
                        onClick={(e) => {
                          e.stopPropagation();
                          const nextTheme = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light';
                          useThemeStore.getState().setTheme(nextTheme);
                        }}
                      >
                        <div className="flex items-center gap-2">
                          <span className="menu-icon" style={{ background: '#F3F4F6', color: '#374151', ...(theme === 'dark' ? { background: '#374151', color: '#F3F4F6' } : {}) }}>
                            <ThemeIcon className="w-4 h-4" />
                          </span>
                          Theme
                        </div>
                        <span className="text-xs text-gray-500 capitalize">{theme}</span>
                      </button>
                    </motion.div>
                    <div className="menu-divider" />
                    <motion.div variants={menuItemVariants} transition={menuTransition}>
                      <button className="menu-item danger" onClick={() => { handleLogout(); }}>
                        <span className="menu-icon" style={{ background: 'rgba(220, 38, 38, 0.2)', color: '#f87171' }}>
                          <ArrowRightOnRectangleIcon className="w-4 h-4" />
                        </span>
                        Sign out
                      </button>
                    </motion.div>
                  </motion.div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </nav>

      <main className={pathname === '/dashboard/query' ? "flex-1 flex flex-col h-[100vh] overflow-hidden" : "main-content"}>
        {children}
      </main>
    </div>
  );
}
