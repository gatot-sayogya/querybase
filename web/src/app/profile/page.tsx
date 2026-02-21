'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import AppLayout from '@/components/layout/AppLayout';
import ProfileSettings from '@/components/profile/ProfileSettings';

export default function ProfilePage() {
  const router = useRouter();
  const { isAuthenticated, isLoading, isHydrating } = useAuthStore();

  useEffect(() => {
    if (!isHydrating && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isHydrating, router]);

  if (isHydrating || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600 dark:text-gray-400">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) return null;

  return (
    <AppLayout>
      <ProfileSettings />
    </AppLayout>
  );
}
