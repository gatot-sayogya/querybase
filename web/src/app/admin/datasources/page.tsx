'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { motion } from 'framer-motion';
import { useAuthStore } from '@/stores/auth-store';
import AppLayout from '@/components/layout/AppLayout';
import DataSourceManager from '@/components/admin/DataSourceManager';
import Loading from '@/components/ui/Loading';

export default function DataSourcesPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading, user, isHydrating } = useAuthStore();

  useEffect(() => {
    if (!isHydrating && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isHydrating, router]);

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'admin') {
      router.push('/dashboard');
    }
  }, [isAuthenticated, isLoading, user, router]);

  if (isHydrating || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.3 }}
        >
          <Loading variant="spinner" size="lg" text="Loading data sources..." />
        </motion.div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  if (user?.role !== 'admin') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Access Denied
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            You don&apos;t have permission to access this page.
          </p>
        </div>
      </div>
    );
  }

  return (
    <AppLayout>
      <DataSourceManager />
    </AppLayout>
  );
}
