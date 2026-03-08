'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Query, ApprovalRequest } from '@/types';
import { formatDate } from '@/lib/utils';

export interface HistoryItem {
  id: string;
  type: 'read' | 'write';
  name: string;
  query_text: string;
  data_source_name: string;
  status: string;
  created_at: string;
  operation_type?: string;
  original: Query | ApprovalRequest;
}

export default function QueryHistory() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuthStore();
  const [historyItems, setHistoryItems] = useState<HistoryItem[]>([]);
  const [activeTab, setActiveTab] = useState<'all' | 'reads' | 'writes'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [debouncedSearch, setDebouncedSearch] = useState(searchQuery);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      setPage(1); // Reset page on new search
    }, 500);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  useEffect(() => {
    const fetchHistory = async () => {
      if (!isAuthenticated) return;

      try {
        setLoading(true);
        setError(null);

        let fetchedQueries: Query[] = [];
        let fetchedApprovals: ApprovalRequest[] = [];
        let newTotal = 0;

        // Fetch Reads
        if (activeTab === 'all' || activeTab === 'reads') {
          const data = await apiClient.getQueryHistory(page, 20, debouncedSearch);
          fetchedQueries = data.queries;
          if (activeTab === 'reads') newTotal = data.total;
        }

        // Fetch Writes
        if (activeTab === 'all' || activeTab === 'writes') {
          const data = await apiClient.getApprovalHistory(page, 20, debouncedSearch);
          fetchedApprovals = data.approvals;
          if (activeTab === 'writes') newTotal = data.total;
        }

        const items: HistoryItem[] = [];
        
        fetchedQueries.forEach(q => {
          items.push({
            id: q.id,
            type: 'read',
            name: q.name || 'Unnamed Query',
            query_text: q.query_text,
            data_source_name: q.data_source_name || 'Unknown',
            status: q.status,
            created_at: q.created_at,
            original: q
          });
        });

        fetchedApprovals.forEach(a => {
          items.push({
            id: a.id,
            type: 'write',
            name: a.operation_type ? `${a.operation_type.toUpperCase()} Request` : 'Write Request',
            query_text: a.query_text,
            data_source_name: a.data_source_name || 'Unknown',
            status: a.status,
            created_at: a.created_at,
            operation_type: a.operation_type || 'UNKNOWN',
            original: a
          });
        });

        items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

        if (activeTab === 'all') {
             // Basic workaround for simple UI, since we're pulling limited pages from both ends
             newTotal = items.length;
        }

        setHistoryItems(items);
        setTotal(newTotal);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load history');
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
  }, [isAuthenticated, page, debouncedSearch, activeTab]);

  const getStatusBadgeColor = (status: string) => {
    switch (status) {
      case 'completed':
      case 'approved':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'failed':
      case 'rejected':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      case 'running':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      case 'pending':
        return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300';
      case 'completed':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'failed':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      case 'running':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600 dark:text-gray-400">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="main-content" style={{ padding: 0 }}>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
          >
            Retry
          </button>
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-4 mb-4" style={{ padding: '0 4px' }}>
        {(['all', 'reads', 'writes'] as const).map(tab => (
            <button
                key={tab}
                onClick={() => { setActiveTab(tab); setPage(1); }}
                className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors duration-200 ${
                  activeTab === tab 
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400' 
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
                }`}
                style={{ background: 'none' }}
            >
                {tab === 'all' ? 'All Activity' : tab === 'reads' ? 'Read Queries' : 'Write Requests'}
            </button>
        ))}
      </div>

      {/* Main Card */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-sm flex flex-col" style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div className="p-8 space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse">
                <div className="h-16 bg-slate-100 rounded-lg"></div>
              </div>
            ))}
          </div>
        ) : error ? null : historyItems.length === 0 ? (
          <div className="p-12 text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-slate-100 text-slate-400 mb-4">
              <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 011.707 0l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <h3 className="text-sm font-medium text-slate-900">No query history</h3>
            <p className="mt-1 text-sm text-slate-500">Execute queries to see them here</p>
          </div>
        ) : (
          <>
            <table className="data-table compact">
              <thead>
                <tr>
                  <th style={{ fontSize: '14px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '12px' }}>
                      <span>QUERY</span>
                      <div style={{ position: 'relative', fontWeight: 400 }}>
                        <svg style={{ position: 'absolute', left: '8px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)', pointerEvents: 'none' }} width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <circle cx="11" cy="11" r="8"></circle>
                          <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                        </svg>
                        <input
                          type="text"
                          style={{ paddingLeft: '28px', height: '28px', fontSize: '12px', width: '200px', border: '1px solid var(--border)', borderRadius: '6px', background: 'var(--bg-card)', color: 'var(--text-primary)', outline: 'none', fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}
                          placeholder="Search queries..."
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                        />
                      </div>
                    </div>
                  </th>
                  <th style={{ fontSize: '14px' }}>DATA SOURCE</th>
                  <th style={{ fontSize: '14px' }}>STATUS</th>
                  <th style={{ fontSize: '14px' }}>TIMESTAMP</th>
                  <th style={{ fontSize: '14px', textAlign: 'right' }}>ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {historyItems.map((item) => {
                  return (
                    <tr key={`${item.type}-${item.id}`}>
                      <td style={{ paddingTop: '4px', paddingBottom: '4px' }}>
                        <div style={{ fontWeight: 500, color: 'var(--text-primary)' }} className="flex items-center gap-2">
                           {item.type === 'write' && (
                              <span className="px-1.5 py-0.5 border border-slate-300 dark:border-slate-700 text-[10px] font-mono font-bold uppercase tracking-wider text-slate-900 dark:text-slate-300 rounded-none bg-slate-100 dark:bg-slate-800">
                                  {item.operation_type}
                              </span>
                          )}
                          {item.name}
                        </div>
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '2px', maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                          {item.query_text}
                        </div>
                      </td>
                      <td>{item.data_source_name}</td>
                      <td>
                        <div className="flex items-center gap-2">
                          <span className={`badge ${getStatusBadgeColor(item.status)}`}>{item.status}</span>
                          {item.type === 'write' && item.status === 'approved' && (item.original as ApprovalRequest).transaction !== undefined && (
                            <span 
                              className={`px-1.5 py-0.5 rounded text-[11px] font-medium font-mono whitespace-nowrap ${(item.original as ApprovalRequest).transaction!.affected_rows > 0 ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800/50' : 'bg-gray-50 dark:bg-gray-800/50 text-gray-500 dark:text-gray-400 border border-gray-200 dark:border-gray-700'}`}
                              title={`${(item.original as ApprovalRequest).transaction!.affected_rows} rows affected in database`}
                            >
                              ⬢ {(item.original as ApprovalRequest).transaction!.affected_rows} rows
                            </span>
                          )}
                        </div>
                      </td>
                      <td style={{ color: 'var(--text-muted)' }}>
                        {item.created_at ? formatDate(item.created_at) : 'N/A'}
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        {item.type === 'read' ? (
                          <button
                            onClick={() => router.push(`/dashboard/queries/${item.id}`)}
                            style={{ color: 'var(--accent-blue)', fontSize: '13px', textDecoration: 'none', background: 'none', border: 'none', cursor: 'pointer' }}
                          >
                            View Results
                          </button>
                        ) : (
                           <button
                             onClick={() => router.push(`/dashboard/approvals?id=${item.id}`)}
                             style={{ color: 'var(--accent-blue)', fontSize: '13px', textDecoration: 'none', background: 'none', border: 'none', cursor: 'pointer' }}
                           >
                             View Details
                           </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>

            {/* Pagination */}
            {total > 0 && (
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '14px 16px', borderTop: '1px solid var(--border-light)' }}>
                <span style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
                  Showing {((page - 1) * 20) + 1} to {Math.min(page * 20, total)} of {total} results
                </span>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <button 
                    className="btn btn-ghost btn-sm" 
                    disabled={page === 1}
                    onClick={() => setPage(Math.max(1, page - 1))}
                    style={page === 1 ? { opacity: 0.4 } : {}}
                  >
                    Previous
                  </button>
                  <button 
                    className="btn btn-ghost btn-sm"
                    disabled={page * 20 >= total}
                    onClick={() => setPage(page + 1)}
                    style={page * 20 >= total ? { opacity: 0.4 } : {}}
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
