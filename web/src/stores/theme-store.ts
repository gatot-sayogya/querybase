import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Theme = 'light' | 'dark' | 'system';
export type EffectiveTheme = 'light' | 'dark';

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  getEffectiveTheme: () => EffectiveTheme;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'system',

      setTheme: (theme) => {
        set({ theme });
        // Apply theme immediately
        applyTheme(theme);
      },

      getEffectiveTheme: () => {
        const { theme } = get();
        if (theme !== 'system') {
          return theme;
        }
        // Check system preference
        return window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light';
      },
    }),
    {
      name: 'theme-storage',
    }
  )
);

// Helper function to apply theme to DOM
function applyTheme(theme: Theme) {
  const root = document.documentElement;
  const effectiveTheme =
    theme === 'system'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
      : theme;

  if (effectiveTheme === 'dark') {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }
}

// Listen for system theme changes when in 'system' mode
if (typeof window !== 'undefined') {
  window
    .matchMedia('(prefers-color-scheme: dark)')
    .addEventListener('change', () => {
      const { theme } = useThemeStore.getState();
      if (theme === 'system') {
        applyTheme('system');
      }
    });
}
