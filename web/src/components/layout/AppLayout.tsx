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
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
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
        <aside
          className={`hidden md:flex flex-col bg-white/80 dark:bg-gray-800/80 backdrop-blur-md border-r border-gray-200 dark:border-gray-700 transition-all duration-300 ease-in-out ${sidebarCollapsed ? 'w-20' : 'w-64'
            }`}
        >
          {/* Logo */}
          <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200 dark:border-gray-700">
            {!sidebarCollapsed && (
              <Link href="/" className="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2 logomark">
                <span className="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center text-white text-sm shadow-lg shadow-blue-500/20">
                  QB
                </span>
                <span className="bg-clip-text text-transparent bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300">
                  QueryBase
                </span>
              </Link>
            )}
            {sidebarCollapsed && (
              <Link href="/" className="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center text-white text-sm mx-auto shadow-lg shadow-blue-500/20">
                QB
              </Link>
            )}
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-2 py-4 space-y-1 overflow-y-auto scrollbar-hide">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  title={sidebarCollapsed ? item.name : undefined}
                  className={`group flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200 ${isActive
                      ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 shadow-sm'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700/50 hover:translate-x-1'
                    }`}
                >
                  <span className={`text-lg transition-transform duration-200 ${isActive ? 'scale-110' : 'group-hover:scale-110'} ${sidebarCollapsed ? 'mx-auto' : 'mr-3'}`}>
                    {item.icon}
                  </span>
                  {!sidebarCollapsed && <span>{item.name}</span>}
                </Link>
              );
            })}

            {isAdmin && (
              <>
                <div className={`pt-4 pb-2 ${sidebarCollapsed ? 'px-1' : 'px-3'}`}>
                  {!sidebarCollapsed && (
                    <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider animate-fade-in">
                      Admin
                    </p>
                  )}
                  {sidebarCollapsed && (
                    <div className="w-full h-px bg-gray-200 dark:bg-gray-700"></div>
                  )}
                </div>
                {adminNavigation.map((item) => {
                  const isActive = pathname === item.href;
                  return (
                    <Link
                      key={item.name}
                      href={item.href}
                      title={sidebarCollapsed ? item.name : undefined}
                      className={`group flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200 ${isActive
                          ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 shadow-sm'
                          : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700/50 hover:translate-x-1'
                        }`}
                    >
                      <span className={`text-lg transition-transform duration-200 ${isActive ? 'scale-110' : 'group-hover:scale-110'} ${sidebarCollapsed ? 'mx-auto' : 'mr-3'}`}>
                        {item.icon}
                      </span>
                      {!sidebarCollapsed && <span>{item.name}</span>}
                    </Link>
                  );
                })}
              </>
            )}
          </nav>

          {/* Collapse Toggle */}
          <div className="p-4 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
              className="w-full flex items-center justify-center px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            >
              <svg
                className={`w-5 h-5 transition-transform duration-300 ${sidebarCollapsed ? 'rotate-180' : ''}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
                />
              </svg>
            </button>
          </div>
        </aside>

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col overflow-hidden bg-gray-50/50 dark:bg-gray-900/50">
          {/* Top Header */}
          <header className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-700 h-16 sticky top-0 z-20">
            <div className="flex items-center justify-between h-full px-4 sm:px-6">
              {/* Left: Mobile menu button + Breadcrumbs */}
              <div className="flex items-center gap-4">
                <button
                  onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                  className="md:hidden p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
                >
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 6h16M4 12h16M4 18h16"
                    />
                  </svg>
                </button>

                {/* Breadcrumbs */}
                <nav className="hidden sm:flex items-center space-x-2 text-sm">
                  {breadcrumbs.map((crumb, index) => (
                    <div key={crumb.href} className="flex items-center animate-fade-in" style={{ animationDelay: `${index * 100}ms` }}>
                      {index > 0 && (
                        <svg
                          className="w-4 h-4 mx-2 text-gray-400"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 5l7 7-7 7"
                          />
                        </svg>
                      )}
                      <Link
                        href={crumb.href}
                        className={`transition-colors duration-200 ${index === breadcrumbs.length - 1
                            ? 'text-gray-900 dark:text-gray-100 font-medium'
                            : 'text-gray-500 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400'
                          }`}
                      >
                        {crumb.label}
                      </Link>
                    </div>
                  ))}
                </nav>
              </div>

              {/* Right: User menu */}
              <div className="flex items-center gap-4">
                {/* Theme toggle */}
                <ThemeToggle />

                <div className="flex items-center gap-3">
                  <div className="hidden sm:block text-right">
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {user?.username}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 capitalize">
                      {user?.role}
                    </p>
                  </div>
                  <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-sm font-semibold shadow-md ring-2 ring-white dark:ring-gray-800">
                    {userInitials}
                  </div>
                </div>
                <button
                  onClick={handleLogout}
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 dark:bg-blue-600 dark:hover:bg-blue-500 rounded-lg transition-all shadow-sm hover:shadow-md active:scale-95"
                >
                  Logout
                </button>
              </div>
            </div>
          </header>

          {/* Main Content */}
          <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8 custom-scrollbar">
            <div className="max-w-7xl mx-auto h-full">{children}</div>
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
