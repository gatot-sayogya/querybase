'use client';

import { useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import ApprovalList from '@/components/approvals/ApprovalList';
import ApprovalDetail from '@/components/approvals/ApprovalDetail';

export default function ApprovalDashboard() {
  const searchParams = useSearchParams();
  const idFromUrl = searchParams.get('id');

  const [selectedApprovalId, setSelectedApprovalId] = useState<string | null>(idFromUrl);
  const [refreshKey, setRefreshKey] = useState(0);

  // When URL ?id= changes, auto-select that approval
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

  return (
    <div className="approvals-grid">
      {/* Approval List */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-sm p-6 approval-list">
        <ApprovalList
          key={refreshKey}
          onSelectApproval={handleSelectApproval}
          selectedId={selectedApprovalId}
          initialFilter={idFromUrl ? 'all' : 'pending'}
        />
      </div>

      {/* Approval Detail */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-sm p-6 detail-col">
        <ApprovalDetail
          key={`${selectedApprovalId}-${refreshKey}`}
          approvalId={selectedApprovalId}
          onRefresh={handleRefresh}
        />
      </div>
    </div>
  );
}
