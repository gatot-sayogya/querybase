'use client';

import { useEffect } from 'react';
import { useThemeStore } from '@/stores/theme-store';

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const { theme, getEffectiveTheme } = useThemeStore();

  // Apply theme immediately on app load to prevent flash
  useEffect(() => {
    const root = document.documentElement;
    const effectiveTheme = getEffectiveTheme();

    if (effectiveTheme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }, [theme, getEffectiveTheme]);

  return <>{children}</>;
}
