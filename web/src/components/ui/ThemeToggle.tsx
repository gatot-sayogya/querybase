'use client';

import { useThemeStore, Theme } from '@/stores/theme-store';
import { useEffect, useState } from 'react';

export function ThemeToggle() {
  const { theme, setTheme, getEffectiveTheme } = useThemeStore();
  const [mounted, setMounted] = useState(false);
  const effectiveTheme = getEffectiveTheme();

  // Prevent hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <button
        className="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        aria-label="Toggle theme"
      >
        <div className="w-5 h-5" />
      </button>
    );
  }

  const cycleTheme = () => {
    const themes: Theme[] = ['light', 'dark', 'system'];
    const currentIndex = themes.indexOf(theme);
    const nextTheme = themes[(currentIndex + 1) % themes.length];
    setTheme(nextTheme);
  };

  return (
    <button
      onClick={cycleTheme}
      className="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors relative"
      aria-label="Toggle theme"
      title={`Current: ${theme}`}
    >
      {/* Sun icon for light mode */}
      {effectiveTheme === 'light' && (
        <svg
          className="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
      )}

      {/* Moon icon for dark mode */}
      {effectiveTheme === 'dark' && (
        <svg
          className="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
          />
        </svg>
      )}

      {/* Small indicator dot for 'system' mode */}
      {theme === 'system' && (
        <span className="absolute bottom-0.5 right-0.5 w-1.5 h-1.5 bg-blue-500 rounded-full" />
      )}
    </button>
  );
}
