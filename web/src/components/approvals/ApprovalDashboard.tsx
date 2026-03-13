'use client';

import { useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import ApprovalList from '@/components/approvals/ApprovalList';
import ApprovalDetail from '@/components/approvals/ApprovalDetail';
import { motion } from 'framer-motion';
import { slideLeft, slideRight, springConfig, duration, reducedMotionVariants } from '@/lib/animations';
import { useReducedMotion } from 'framer-motion';

export default function ApprovalDashboard() {
  const searchParams = useSearchParams();
  const idFromUrl = searchParams.get('id');
  const shouldReduceMotion = useReducedMotion();

  const [selectedApprovalId, setSelectedApprovalId] = useState<string | null>(idFromUrl);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    if (idFromUrl) {
      setSelectedApprovalId(idFromUrl);
    }
  }, [idFromUrl]);

  const handleSelectApproval = (approvalId: string) => {
    setSelectedApprovalId(approvalId);
  };

  const handleRefresh = () => {
    setRefreshKey((prev) => prev + 1);
  };

  const listVariants = shouldReduceMotion ? reducedMotionVariants : slideLeft;
  const detailVariants = shouldReduceMotion ? reducedMotionVariants : slideRight;

  return (
    <div className="approvals-grid">
      <motion.div
        className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-sm p-6 approval-list"
        variants={listVariants}
        initial="initial"
        animate="animate"
        transition={{ duration: duration.slow, ...springConfig.gentle }}
      >
        <ApprovalList
          key={refreshKey}
          onSelectApproval={handleSelectApproval}
          selectedId={selectedApprovalId}
          initialFilter={idFromUrl ? 'all' : 'pending'}
        />
      </motion.div>

      <motion.div
        className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-sm p-6 detail-col"
        variants={detailVariants}
        initial="initial"
        animate="animate"
        transition={{ duration: duration.slow, delay: 0.1, ...springConfig.gentle }}
      >
        <ApprovalDetail
          key={`${selectedApprovalId}-${refreshKey}`}
          approvalId={selectedApprovalId}
          onRefresh={handleRefresh}
        />
      </motion.div>
    </div>
  );
}
