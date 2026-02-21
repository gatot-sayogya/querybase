'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import AppLayout from '@/components/layout/AppLayout';
import Loading from '@/components/ui/Loading';
import PageTransition from '@/components/layout/PageTransition';
import Link from 'next/link';
import { useDashboardStats } from '@/hooks/useDashboardStats';
import { apiClient } from '@/lib/api-client';
import type { Query, ApprovalRequest, DataSource, HealthStatus } from '@/types';

export default function DashboardPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading, isHydrating } = useAuthStore();
  const { stats, isLoading: statsLoading } = useDashboardStats();

  const [recentQueries, setRecentQueries] = useState<Query[]>([]);
  const [recentQueriesLoading, setRecentQueriesLoading] = useState(true);

  const [pendingApprovals, setPendingApprovals] = useState<ApprovalRequest[]>([]);
  const [pendingApprovalsLoading, setPendingApprovalsLoading] = useState(true);

  const [dataSources, setDataSources] = useState<(DataSource & { health?: HealthStatus })[]>([]);
  const [dataSourcesLoading, setDataSourcesLoading] = useState(true);

  useEffect(() => {
    if (!isHydrating && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isHydrating, router]);

  useEffect(() => {
    if (!isAuthenticated || isHydrating) return;

    // Reset loading states
    setRecentQueriesLoading(true);
    setPendingApprovalsLoading(true);
    setDataSourcesLoading(true);

    // Fetch Recent Queries
    apiClient.getQueryHistory(1, 4)
      .then(res => setRecentQueries(res.queries))
      .catch(err => console.error("Failed to fetch recent queries:", err))
      .finally(() => setRecentQueriesLoading(false));

    // Fetch Pending Approvals & Data Sources Health
    if (user?.role === 'admin') {
      apiClient.getApprovals({ status: 'pending', page: 1 })
        .then(res => setPendingApprovals(res.slice(0, 3)))
        .catch(err => console.error("Failed to fetch pending approvals:", err))
        .finally(() => setPendingApprovalsLoading(false));
      
      // Fetch Data Sources & Health
      (async () => {
        try {
          const sourcesTimeout = new Promise<DataSource[]>((_, reject) =>
            setTimeout(() => reject(new Error('timeout')), 5000)
          );
          
          const sources = await Promise.race([apiClient.getDataSources(), sourcesTimeout]);
          const topSources = sources.slice(0, 4);
          
          const sourcesWithHealth = await Promise.all(
            topSources.map(async (source) => {
              const healthTimeout = new Promise<HealthStatus>((_, reject) =>
                setTimeout(() => reject(new Error('timeout')), 3000)
              );
              try {
                const health = await Promise.race([apiClient.getDataSourceHealth(source.id), healthTimeout]);
                return { ...source, health };
              } catch {
                return { 
                  ...source, 
                  health: { 
                    status: 'unhealthy' as const, 
                    latency_ms: 0, 
                    last_checked: new Date().toISOString(), 
                    data_source_id: source.id, 
                    message: 'Unavailable' 
                  } as HealthStatus 
                };
              }
            })
          );
          setDataSources(sourcesWithHealth);
        } catch (err) {
          console.error('Failed to fetch datasources or health:', err);
        } finally {
          setDataSourcesLoading(false);
        }
      })();
    } else {
      setPendingApprovalsLoading(false);
      setDataSourcesLoading(false);
    }
  }, [isAuthenticated, isHydrating, user?.role]);

  if (isHydrating || isLoading) {
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

  const isAdmin = user?.role === 'admin';

  return (
    <PageTransition animation="fade">
      <AppLayout>
        {/* Header */}
        <div className="relative overflow-hidden bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl shadow-sm mb-8">
          {/* Decorative background element */}
          <div className="absolute top-0 right-0 -mt-4 -mr-4 w-32 h-32 bg-gradient-to-br from-blue-100 to-indigo-50 dark:from-blue-900/40 dark:to-indigo-800/20 rounded-full blur-2xl opacity-70 border-none pointer-events-none"></div>
          
          <div className="relative p-8 sm:p-10 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
            <div>
              <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300">
                Welcome back, {user?.username}! 👋
              </h1>
              <p className="mt-2 text-gray-500 dark:text-gray-400 text-base max-w-xl">
                Here is what&apos;s happening with your database queries and approvals today.
              </p>
            </div>
            
            <div className="flex-shrink-0">
              <Link 
                href="/dashboard/query" 
                className="group relative inline-flex items-center justify-center gap-2 px-6 py-3 text-sm font-medium text-white transition-all duration-200 bg-blue-600 border border-transparent rounded-xl shadow-sm hover:bg-blue-700 hover:shadow-md focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 overflow-hidden"
              >
                <div className="absolute inset-0 w-full h-full -x-100 bg-gradient-to-r from-transparent via-white/20 to-transparent group-hover:animate-shimmer"></div>
                <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                <span>New Query</span>
              </Link>
            </div>
          </div>
        </div>

        {/* Content Wrapper */}
        <div className="content-wrapper">
          {/* Stats Row */}
          <div className={`stats-row ${isAdmin ? 'admin-view' : 'user-view'}`} id="statsRow">
            {/* My Queries */}
            <div className="card stat-card">
              <div className="stat-top">
                <div className="stat-icon" style={{ background: '#EFF6FF', color: '#2563EB' }}>⌕</div>
              </div>
              <div className="stat-label">My Queries Today</div>
              <div className="stat-value">{statsLoading ? '...' : stats?.my_queries_today || 0}</div>
            </div>

            {isAdmin && (
              <>
                {/* Pending approvals */}
                <div className="card stat-card">
                  <div className="stat-top">
                    <div className="stat-icon" style={{ background: '#FEF2F2', color: '#DC2626' }}>✔</div>
                  </div>
                  <div className="stat-label">Pending Approvals</div>
                  <div className="stat-value text-red-600">{statsLoading ? '...' : stats?.pending_approvals || 0}</div>
                </div>
                {/* Total queries */}
                <div className="card stat-card">
                  <div className="stat-top">
                    <div className="stat-icon" style={{ background: '#F0FDF4', color: '#059669' }}>↻</div>
                  </div>
                  <div className="stat-label">Total Queries (All Users)</div>
                  <div className="stat-value">{statsLoading ? '...' : stats?.total_queries?.toLocaleString() || 0}</div>
                </div>
              </>
            )}

            {/* DB Access */}
            <div className="card stat-card">
              <div className="stat-top">
                <div className="stat-icon" style={{ background: '#EDE9FE', color: '#4F46E5' }}>◉</div>
              </div>
              <div className="stat-label">DB Access</div>
              <div className="stat-value">{statsLoading ? '...' : stats?.db_access_count || 0}</div>
            </div>

            {isAdmin && (
              /* Users */
              <div className="card stat-card">
                <div className="stat-top">
                  <div className="stat-icon" style={{ background: '#FEF3C7', color: '#92400E' }}>👥</div>
                </div>
                <div className="stat-label">Total Users</div>
                <div className="stat-value">{statsLoading ? '...' : stats?.total_users || 0}</div>
              </div>
            )}
          </div>

          <div className="card mt-6">
            <div className="flex justify-between items-center px-5 pt-5 pb-0">
              <div className="font-bold text-[15px] text-slate-900 dark:text-white">Recent Activity</div>
              <Link href="/dashboard/history" className="text-[13px] text-blue-600 dark:text-blue-400 hover:underline">View all →</Link>
            </div>
            <ul className="flex flex-col mt-4">
              {recentQueriesLoading ? (
                <div className="py-8 flex justify-center"><Loading variant="spinner" size="sm" /></div>
              ) : recentQueries.length === 0 ? (
                <div className="py-8 text-center text-sm text-slate-500 dark:text-slate-400">No recent queries.</div>
              ) : (
                recentQueries.map(query => {
                  let statusColor = 'bg-slate-500';
                  let pillClass = 'pill-slate';
                  if (query.status === 'completed') {
                    statusColor = 'bg-green-500';
                    pillClass = 'pill-green';
                  } else if (query.status === 'failed') {
                    statusColor = 'bg-red-600';
                    pillClass = 'pill-red';
                  } else if (query.status === 'running') {
                    statusColor = 'bg-blue-500';
                    pillClass = 'pill-blue';
                  } else if (query.status === 'pending') {
                    statusColor = 'bg-amber-500';
                    pillClass = 'pill-amber';
                  }

                  const dateObj = new Date(query.created_at);
                  const timeStr = dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                  return (
                    <li key={query.id} className="flex items-center gap-[14px] px-4 py-[14px] border-b border-slate-200 dark:border-slate-700 last:border-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                      <div className={`w-2 h-2 rounded-full flex-shrink-0 ${statusColor}`}></div>
                      <span className="text-[13px] font-medium text-slate-900 dark:text-slate-100 flex-1 overflow-hidden text-ellipsis whitespace-nowrap max-w-[500px]" title={query.query_text}>
                        {query.query_text}
                      </span>
                      <span className={`pill ${pillClass}`}>{query.status}</span>
                      <span className="text-[12px] text-slate-500 dark:text-slate-400 whitespace-nowrap ml-auto text-right">
                        {dateObj.toLocaleDateString() === new Date().toLocaleDateString() ? timeStr : dateObj.toLocaleDateString()}
                      </span>
                    </li>
                  );
                })
              )}
            </ul>
          </div>

          {/* Admin System Overview */}
          {isAdmin && (
            <div className="mt-6">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {/* Pending approvals breakdown */}
                <div className="card overflow-hidden">
                  <div className="flex justify-between items-center border-b border-slate-100 dark:border-slate-700/50 px-4 pt-4 pb-3">
                    <div className="font-bold text-[15px] text-slate-900 dark:text-slate-100">Pending Approvals</div>
                    <Link href="/dashboard/approvals" className="text-[13px] text-blue-600 dark:text-blue-400 hover:underline">Review →</Link>
                  </div>
                  <div>
                    {pendingApprovalsLoading ? (
                      <div className="py-8 flex justify-center"><Loading variant="spinner" size="sm" /></div>
                    ) : pendingApprovals.length === 0 ? (
                      <div className="py-6 text-center text-sm text-slate-500 dark:text-slate-400">All caught up!</div>
                    ) : (
                      pendingApprovals.map(approval => {
                        const dateObj = new Date(approval.created_at);
                        const isToday = dateObj.toLocaleDateString() === new Date().toLocaleDateString();
                        const timeStr = isToday ? dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : dateObj.toLocaleDateString();

                        return (
                          <div key={approval.id} className="flex flex-col sm:flex-row sm:items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-700/50 last:border-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors gap-2">
                            <span className="text-[13px] text-slate-700 dark:text-slate-300 overflow-hidden text-ellipsis whitespace-nowrap flex-1" title={approval.query_text}>
                              <span className="font-semibold text-slate-900 dark:text-slate-100 mr-2">{approval.operation_type}</span>
                              {approval.query_text}
                            </span>
                            <span className="text-[12px] text-slate-500 dark:text-slate-400 whitespace-nowrap">
                              {timeStr}
                            </span>
                          </div>
                        );
                      })
                    )}
                  </div>
                </div>

                {/* Data source health */}
                <div className="card overflow-hidden">
                  <div className="flex justify-between items-center border-b border-slate-100 dark:border-slate-700/50 px-4 pt-4 pb-3">
                    <div className="font-bold text-[15px] text-slate-900 dark:text-slate-100">Data Source Health</div>
                    <Link href="/admin/datasources" className="text-[13px] text-blue-600 dark:text-blue-400 hover:underline">Manage →</Link>
                  </div>
                  <div>
                    {dataSourcesLoading ? (
                      <div className="py-8 flex justify-center"><Loading variant="spinner" size="sm" /></div>
                    ) : dataSources.length === 0 ? (
                      <div className="py-6 text-center text-sm text-slate-500 dark:text-slate-400">No data sources.</div>
                    ) : (
                      dataSources.map(source => {
                        const isConnected = source.health?.status === 'healthy';
                        
                        return (
                          <div key={source.id} className="flex items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-700/50 last:border-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                            <div className="flex flex-col">
                              <span className="text-[13px] font-medium text-slate-900 dark:text-slate-100">{source.name}</span>
                              <span className="text-[11px] text-slate-500">{source.type}</span>
                            </div>
                            <span className={`pill ${isConnected ? 'pill-green' : 'pill-red'} flex items-center gap-1`}>
                              {isConnected ? (
                                <>
                                  <span className="w-1.5 h-1.5 rounded-full bg-green-500"></span>
                                  Connected
                                </>
                              ) : (
                                <>
                                  <span className="text-red-500">⚠</span>
                                  Error
                                </>
                              )}
                            </span>
                          </div>
                        );
                      })
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

        </div>
      </AppLayout>
    </PageTransition>
  );
}
