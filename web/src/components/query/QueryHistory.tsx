'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Query } from '@/types';
import { formatDate } from '@/lib/utils';

export default function QueryHistory() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuthStore();
  const [queries, setQueries] = useState<Query[]>([]);
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
        const data = await apiClient.getQueryHistory(page, 20, debouncedSearch);
        setQueries(data.queries);
        setTotal(data.total);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load query history');
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
  }, [isAuthenticated, page, debouncedSearch]);

  const getStatusBadgeColor = (status: string) => {
    switch (status) {
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
        ) : error ? null : queries.length === 0 ? (
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
                {queries.map((query) => {
                  let badgeClass = 'badge-slate';
                  if (query.status === 'completed') badgeClass = 'badge-green';
                  else if (query.status === 'failed') badgeClass = 'badge-red text-red-700 bg-red-50';
                  else if (query.status === 'running' || query.status === 'pending') badgeClass = 'badge-amber';

                  return (
                    <tr key={query.id}>
                      <td style={{ paddingTop: '4px', paddingBottom: '4px' }}>
                        <div style={{ fontWeight: 500, color: 'var(--text-primary)' }}>
                          {query.name || 'Unnamed Query'}
                        </div>
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '2px', maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                          {query.query_text}
                        </div>
                      </td>
                      <td>{query.data_source_name || 'Unknown'}</td>
                      <td>
                        <span className={`badge ${badgeClass}`}>{query.status}</span>
                      </td>
                      <td style={{ color: 'var(--text-muted)' }}>
                        {query.created_at ? formatDate(query.created_at) : 'N/A'}
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <button
                          onClick={() => router.push(`/dashboard/queries/${query.id}`)}
                          style={{ color: 'var(--accent-blue)', fontSize: '13px', textDecoration: 'none', background: 'none', border: 'none', cursor: 'pointer' }}
                        >
                          View Results
                        </button>
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
