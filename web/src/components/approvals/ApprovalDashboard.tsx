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
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Approval List */}
      <div className="lg:col-span-1">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Approval Requests
          </h2>
          <ApprovalList
            key={refreshKey}
            onSelectApproval={handleSelectApproval}
            selectedId={selectedApprovalId}
          />
        </div>
      </div>

      {/* Approval Detail */}
      <div className="lg:col-span-2">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <ApprovalDetail
            key={`${selectedApprovalId}-${refreshKey}`}
            approvalId={selectedApprovalId}
            onRefresh={handleRefresh}
          />
        </div>
      </div>
    </div>
  );
}
