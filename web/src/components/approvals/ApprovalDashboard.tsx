'use client';

import { useState } from 'react';
import ApprovalList from '@/components/approvals/ApprovalList';
import ApprovalDetail from '@/components/approvals/ApprovalDetail';

export default function ApprovalDashboard() {
  const [selectedApprovalId, setSelectedApprovalId] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

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
