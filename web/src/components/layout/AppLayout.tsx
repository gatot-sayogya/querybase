'use client';

import { usePathname, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/stores/auth-store';
import { useThemeStore } from '@/stores/theme-store';
import { ThemeToggle } from '@/components/ui/ThemeToggle';
import { useState, useEffect } from 'react';

interface AppLayoutProps {
  children: React.ReactNode;
}

export default function AppLayout({ children }: AppLayoutProps) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const { theme, getEffectiveTheme } = useThemeStore();
  // No toggle state needed for fixed rail layout
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Initialize theme on mount
  useEffect(() => {
    const root = document.documentElement;
    const effectiveTheme = getEffectiveTheme();

    if (effectiveTheme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }, [theme, getEffectiveTheme]);

  const handleLogout = async () => {
    await logout();
    router.push('/login');
  };

  const navigation = [
    { name: 'Query Editor', href: '/dashboard', icon: '🔍' },
    { name: 'Query History', href: '/dashboard/history', icon: '📜' },
    { name: 'Approvals', href: '/dashboard/approvals', icon: '✅' },
  ];

  const adminNavigation = [
    { name: 'Data Sources', href: '/admin/datasources', icon: '💾' },
    { name: 'Users', href: '/admin/users', icon: '👥' },
    { name: 'Groups', href: '/admin/groups', icon: '🏢' },
  ];

  const isAdmin = user?.role === 'admin';

  // Generate user avatar with gradient
  const userInitials = user?.username
    ?.split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2) || 'UQ';

  // Breadcrumb generation
  const getBreadcrumbs = () => {
    const segments = pathname.split('/').filter(Boolean);
    return segments.map((segment, index) => {
      const href = '/' + segments.slice(0, index + 1).join('/');
      const label = segment
        .split('-')
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
      return { href, label };
    });
  };

  const breadcrumbs = getBreadcrumbs();

  return (
    <div className="min-h-screen bg-background">
      <div className="flex h-screen overflow-hidden">
        {/* Sidebar - Desktop */}
        {/* Sidebar - Desktop */}
        <aside className="hidden md:flex flex-col w-20 bg-gray-900 border-r border-gray-800 flex-shrink-0 z-30">
          {/* Logo */}
          <div className="flex items-center justify-center h-16 border-b border-gray-800">
            <Link href="/" className="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center text-white text-sm font-bold shadow-lg shadow-blue-500/20">
              QB
            </Link>
          </div>

          {/* Navigation */}
          <nav className="flex-1 py-4 flex flex-col items-center space-y-2 overflow-y-auto scrollbar-hide">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={`group flex flex-col items-center justify-center w-16 h-16 rounded-xl transition-all duration-200 ${isActive
                      ? 'bg-blue-600/10 text-blue-500'
                      : 'text-gray-400 hover:text-gray-100 hover:bg-gray-800'
                    }`}
                >
                  <span className={`text-2xl mb-1 transition-transform duration-200 ${isActive ? 'scale-110' : 'group-hover:scale-110'}`}>
                    {item.icon}
                  </span>
                  <span className="text-[10px] font-medium tracking-tight text-center leading-none">
                    {item.name.split(' ')[0]}
                  </span>
                </Link>
              );
            })}
          </nav>

          {/* Bottom Section */}
          <div className="py-4 flex flex-col items-center space-y-4 border-t border-gray-800 bg-gray-900">
            {/* Help */}
            <button
              className="group flex flex-col items-center justify-center w-12 h-12 rounded-xl text-gray-400 hover:text-gray-100 hover:bg-gray-800 transition-all duration-200"
              title="Help"
            >
              <svg className="w-6 h-6 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-[10px] font-medium">Help</span>
            </button>

            {/* Admin / Settings */}
            {isAdmin && (
              <Link
                href="/admin/datasources"
                className={`group flex flex-col items-center justify-center w-12 h-12 rounded-xl transition-all duration-200 ${pathname?.startsWith('/admin')
                    ? 'text-blue-500 bg-blue-600/10'
                    : 'text-gray-400 hover:text-gray-100 hover:bg-gray-800'
                  }`}
                title="Settings"
              >
                <svg className="w-6 h-6 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span className="text-[10px] font-medium">Settings</span>
              </Link>
            )}

            {/* User Profile & Menu */}
            <div className="relative group">
              <button
                className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-sm font-semibold shadow-md border-2 border-gray-800 hover:border-gray-600 transition-colors"
              >
                {userInitials}
              </button>
              
              {/* Popover Menu */}
              <div className="absolute left-14 bottom-0 w-64 bg-gray-900 border border-gray-700 rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 transform translate-x-2 group-hover:translate-x-0 z-50">
                <div className="p-4 border-b border-gray-800">
                  <p className="text-sm font-medium text-white">{user?.username}</p>
                  <p className="text-xs text-gray-400 capitalize">{user?.role}</p>
                </div>
                
                <div className="py-2">
                  <button className="w-full text-left px-4 py-2 text-sm text-gray-300 hover:bg-gray-800 hover:text-white transition-colors">
                    Profile
                  </button>
                  <button 
                    onClick={() => theme === 'dark' ? document.documentElement.classList.remove('dark') : document.documentElement.classList.add('dark')}
                    className="w-full text-left px-4 py-2 text-sm text-gray-300 hover:bg-gray-800 hover:text-white transition-colors flex items-center justify-between"
                  >
                    <span>Theme</span>
                    {/* ThemeToggle functionality logic would be here, simplifying for direct toggle */}
                    <ThemeToggle />
                  </button>
                  <button className="w-full text-left px-4 py-2 text-sm text-gray-300 hover:bg-gray-800 hover:text-white transition-colors">
                    System Status
                  </button>
                </div>

                <div className="border-t border-gray-800 py-2">
                  <button
                    onClick={handleLogout}
                    className="w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-gray-800 hover:text-red-300 transition-colors"
                  >
                    Log out
                  </button>
                </div>

                <div className="px-4 py-2 bg-gray-950 rounded-b-lg border-t border-gray-800">
                  <p className="text-[10px] text-gray-500">Version: 10.1.0 (2589bef1)</p>
                </div>
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col overflow-hidden bg-gray-50/50 dark:bg-gray-900/50">
          {/* Mobile Header - Only visible on small screens */}
          <div className="md:hidden flex items-center justify-between h-14 px-4 bg-white/80 dark:bg-gray-800/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-700">
            <Link href="/" className="font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span className="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center text-white text-sm">
                QB
              </span>
              QueryBase
            </Link>
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          </div>

          {/* Main Content */}
          <main className="flex-1 overflow-y-auto w-full custom-scrollbar">
            <div className="w-full h-full">{children}</div>
          </main>
        </div>
      </div>

      {/* Mobile Menu Overlay */}
      {mobileMenuOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/50 md:hidden"
            onClick={() => setMobileMenuOpen(false)}
          ></div>
          <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 shadow-xl md:hidden">
            <div className="flex flex-col h-full">
              {/* Mobile Logo */}
              <div className="flex items-center justify-between h-16 px-4 border-b border-border">
                <Link href="/" className="text-xl font-bold text-foreground flex items-center gap-2">
                  <span className="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center text-white text-sm">
                    QB
                  </span>
                  QueryBase
                </Link>
                <button
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
                >
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {/* Mobile Navigation */}
              <nav className="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
                {navigation.map((item) => {
                  const isActive = pathname === item.href;
                  return (
                    <Link
                      key={item.name}
                      href={item.href}
                      onClick={() => setMobileMenuOpen(false)}
                      className={`flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-colors ${isActive
                          ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                          : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                        }`}
                    >
                      <span className="mr-3 text-lg">{item.icon}</span>
                      <span>{item.name}</span>
                    </Link>
                  );
                })}

                {isAdmin && (
                  <>
                    <div className="pt-4 pb-2 px-3">
                      <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Admin
                      </p>
                    </div>
                    {adminNavigation.map((item) => {
                      const isActive = pathname === item.href;
                      return (
                        <Link
                          key={item.name}
                          href={item.href}
                          onClick={() => setMobileMenuOpen(false)}
                          className={`flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-colors ${isActive
                              ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                            }`}
                        >
                          <span className="mr-3 text-lg">{item.icon}</span>
                          <span>{item.name}</span>
                        </Link>
                      );
                    })}
                  </>
                )}
              </nav>

              {/* Mobile User Info */}
              <div className="p-4 border-t border-border">
                <div className="flex items-center gap-3 mb-3">
                  <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-sm font-semibold">
                    {userInitials}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {user?.username}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 capitalize">
                      {user?.role}
                    </p>
                  </div>
                </div>
                <div className="flex items-center justify-between mb-4">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Theme</span>
                    <ThemeToggle />
                </div>
                <button
                  onClick={handleLogout}
                  className="w-full px-4 py-2 text-sm font-medium text-white bg-primary hover:bg-blue-700 rounded-lg transition-colors"
                >
                  Logout
                </button>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
