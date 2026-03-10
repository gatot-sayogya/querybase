'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import AppLayout from '@/components/layout/AppLayout';
import QueryHistory from '@/components/query/QueryHistory';
import Loading from '@/components/ui/Loading';
import { motion } from 'framer-motion';

export default function QueryHistoryPage() {
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
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.3 }}
        >
          <Loading variant="spinner" size="lg" />
        </motion.div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <AppLayout>
      <QueryHistory />
    </AppLayout>
  );
}
