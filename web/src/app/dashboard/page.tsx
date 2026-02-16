'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import AppLayout from '@/components/layout/AppLayout';
import QueryExecutor from '@/components/query/QueryExecutor';
import Loading from '@/components/ui/Loading';
import PageTransition from '@/components/layout/PageTransition';

export default function DashboardPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuthStore();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <Loading variant="spinner" size="lg" text="Loading dashboard..." />
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <PageTransition animation="fade">
      <AppLayout>
        <QueryExecutor />
      </AppLayout>
    </PageTransition>
  );
}
