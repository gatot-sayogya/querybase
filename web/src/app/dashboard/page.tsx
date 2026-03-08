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

  const [myRequests, setMyRequests] = useState<ApprovalRequest[]>([]);
  const [myRequestsLoading, setMyRequestsLoading] = useState(true);
  const [approvalCounts, setApprovalCounts] = useState<Record<string, number>>({ pending: 0, approved: 0, rejected: 0 });

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
    setMyRequestsLoading(true);
    setDataSourcesLoading(true);

    // Fetch Recent Queries
    apiClient.getQueryHistory(1, 4)
      .then(res => setRecentQueries(res.queries))
      .catch(err => console.error("Failed to fetch recent queries:", err))
      .finally(() => setRecentQueriesLoading(false));

    // Fetch Pending Approvals & Data Sources Health for everyone
    apiClient.getApprovals({ status: 'pending', page: 1 })
      .then(res => setPendingApprovals(res.slice(0, 3)))
      .catch(err => console.error("Failed to fetch pending approvals:", err))
      .finally(() => setPendingApprovalsLoading(false));

    // Fetch My Recent Requests (All Statuses) and Counts
    apiClient.getApprovals({ page: 1 })
      .then(res => setMyRequests(res.slice(0, 5)))
      .catch(err => console.error("Failed to fetch my requests:", err))
      .finally(() => setMyRequestsLoading(false));

    apiClient.getApprovalCounts()
      .then(res => setApprovalCounts(res))
      .catch(err => console.error("Failed to fetch approval counts:", err));

    if (user?.role === 'admin') {
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
      setDataSourcesLoading(false);
    }
  }, [isAuthenticated, isHydrating, user?.role]);

  if (isHydrating || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900 border-none">
        <div className="text-center">
          <Loading variant="spinner" size="lg" text="Initializing Matrix..." />
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
        {/* Header - Neo Technical */}
        <div className="relative overflow-hidden bg-slate-900 border border-slate-800 rounded-none shadow-sm mb-6 mt-2 ml-2 mr-2">
          {/* Brutalist Grid pattern overlay */}
          <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px)', backgroundSize: '32px 32px' }}></div>
          
          <div className="relative p-6 sm:p-8 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
            <div>
              <div className="text-emerald-500 font-mono text-xs uppercase tracking-widest mb-1 flex items-center gap-2">
                <span className="w-2 h-2 bg-emerald-500 rounded-none animate-pulse"></span>
                System Terminals Active
              </div>
              <h1 className="text-2xl font-bold text-white tracking-tight">
                Welcome back, {user?.username}
              </h1>
              <p className="mt-1 text-slate-400 text-sm max-w-xl">
                Current operational state and telemetry for your database queries.
              </p>
            </div>
            
            <div className="flex-shrink-0">
              <Link 
                href="/dashboard/query" 
                className="group relative inline-flex items-center justify-center gap-2 px-6 py-2.5 text-sm font-bold text-slate-900 transition-all duration-200 bg-emerald-400 border border-emerald-400 hover:bg-emerald-300 shadow-[4px_4px_0px_0px_rgba(255,255,255,0.1)] hover:shadow-[2px_2px_0px_0px_rgba(255,255,255,0.1)] hover:translate-x-[2px] hover:translate-y-[2px]"
                style={{ borderRadius: '0px' }}
              >
                <svg className="w-5 h-5 transition-transform group-hover:rotate-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="square" strokeLinejoin="miter" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                <span className="uppercase tracking-wide">Initialize Query</span>
              </Link>
            </div>
          </div>
        </div>

        {/* Content Wrapper */}
        <div className="px-2 pb-6 flex flex-col gap-6">
          {/* Stats Row - Asymmetrical grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-4">
            {/* My Queries */}
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 rounded-none flex flex-col justify-between h-32 hover:border-blue-500 transition-colors">
              <div className="flex justify-between items-start">
                <div className="text-xs font-mono text-slate-500 uppercase tracking-wider">Executed Today</div>
                <div className="w-8 h-8 flex items-center justify-center bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 rounded-none">⌕</div>
              </div>
              <div className="text-3xl font-bold text-slate-900 dark:text-white tracking-tighter">
                {statsLoading ? '--' : stats?.my_queries_today || 0}
              </div>
            </div>

            {/* DB Access */}
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 rounded-none flex flex-col justify-between h-32 hover:border-emerald-500 transition-colors">
              <div className="flex justify-between items-start">
                <div className="text-xs font-mono text-slate-500 uppercase tracking-wider">DB Access</div>
                <div className="w-8 h-8 flex items-center justify-center bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 rounded-none">◉</div>
              </div>
              <div className="text-3xl font-bold text-slate-900 dark:text-white tracking-tighter">
                {statsLoading ? '--' : stats?.db_access_count || 0}
              </div>
            </div>

            {/* Pending Approvals (Everyone) */}
             <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 rounded-none flex flex-col justify-between h-32 hover:border-amber-500 border-l-4 border-l-amber-500 transition-colors">
              <div className="flex justify-between items-start">
                <div className="text-xs font-mono text-slate-500 uppercase tracking-wider">{isAdmin ? 'Pending Action' : 'My Pending'}</div>
                <div className="w-8 h-8 flex items-center justify-center bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 rounded-none">⏳</div>
              </div>
              <div className="text-3xl font-bold text-amber-600 tracking-tighter">
                {statsLoading ? '--' : stats?.pending_approvals || 0}
              </div>
            </div>

            {/* Approved Activity (Self or Global depending on role) */}
             <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 rounded-none flex flex-col justify-between h-32 hover:border-blue-500 transition-colors">
              <div className="flex justify-between items-start">
                <div className="text-xs font-mono text-slate-500 uppercase tracking-wider">{isAdmin ? 'Global Approved' : 'My Approved'}</div>
                <div className="w-8 h-8 flex items-center justify-center bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 rounded-none">✅</div>
              </div>
              <div className="text-3xl font-bold text-blue-600 tracking-tighter">
                {approvalCounts.approved || 0}
              </div>
            </div>

            {isAdmin && (
              <>
                {/* Total queries */}
                <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 rounded-none flex flex-col justify-between h-32 hover:border-slate-400 transition-colors">
                  <div className="flex justify-between items-start">
                    <div className="text-xs font-mono text-slate-500 uppercase tracking-wider">Global Exec</div>
                    <div className="w-8 h-8 flex items-center justify-center bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-none">↻</div>
                  </div>
                  <div className="text-3xl font-bold text-slate-900 dark:text-white tracking-tighter">
                    {statsLoading ? '--' : stats?.total_queries?.toLocaleString() || 0}
                  </div>
                </div>

                {/* Users */}
                <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 rounded-none flex flex-col justify-between h-32 hover:border-slate-400 transition-colors">
                  <div className="flex justify-between items-start">
                    <div className="text-xs font-mono text-slate-500 uppercase tracking-wider">User Count</div>
                    <div className="w-8 h-8 flex items-center justify-center bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-none">👥</div>
                  </div>
                  <div className="text-3xl font-bold text-slate-900 dark:text-white tracking-tighter">
                    {statsLoading ? '--' : stats?.total_users || 0}
                  </div>
                </div>
              </>
            )}
          </div>

          <div className={isAdmin ? "grid grid-cols-1 lg:grid-cols-3 gap-6" : "grid grid-cols-1 lg:grid-cols-2 gap-6"}>
            
            {/* Recent Query Activity - Takes up more space */ }
            <div className={isAdmin ? "lg:col-span-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-none flex flex-col" : "bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-none flex flex-col"}>
              <div className="flex justify-between items-center px-6 pt-5 pb-4 border-b border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-3">
                  <div className="w-2 h-6 bg-blue-500"></div>
                  <div className="font-mono text-xs uppercase tracking-widest text-slate-900 dark:text-white font-bold">Execution Telemetry</div>
                </div>
                <Link href="/dashboard/history" className="text-xs font-mono text-blue-600 dark:text-blue-400 hover:text-blue-500 hover:underline uppercase tracking-wider">View Log →</Link>
              </div>
              <div className="flex flex-col flex-1 p-0">
                {recentQueriesLoading ? (
                  <div className="py-12 flex justify-center"><Loading variant="spinner" size="sm" /></div>
                ) : recentQueries.length === 0 ? (
                  <div className="py-12 flex flex-col items-center justify-center text-slate-500 dark:text-slate-400">
                    <div className="font-mono text-xs uppercase tracking-widest mb-2 opacity-50">[NO SIGNAL]</div>
                    <div className="text-sm">No recent queries detected in the log.</div>
                  </div>
                ) : (
                  <div className="flex flex-col">
                    {recentQueries.map((query, idx) => {
                      let statusColor = 'bg-slate-500';
                      let statusText = 'text-slate-500';
                      
                      if (query.status === 'completed') {
                        statusColor = 'bg-emerald-500';
                        statusText = 'text-emerald-500 dark:text-emerald-400';
                      } else if (query.status === 'failed') {
                        statusColor = 'bg-rose-500';
                        statusText = 'text-rose-600 dark:text-rose-400';
                      } else if (query.status === 'running') {
                        statusColor = 'bg-blue-500';
                        statusText = 'text-blue-600 dark:text-blue-400';
                      } else if (query.status === 'pending') {
                        statusColor = 'bg-amber-500';
                        statusText = 'text-amber-600 dark:text-amber-400';
                      }

                      const dateObj = new Date(query.created_at);
                      const timeStr = dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

                      return (
                        <div key={query.id} className={`flex items-start gap-4 px-6 py-4 border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors ${idx === recentQueries.length - 1 ? 'border-b-0' : ''}`}>
                          <div className={`mt-1.5 w-1.5 h-1.5 rounded-none flex-shrink-0 ${statusColor}`}></div>
                          <div className="flex-1 min-w-0">
                            <div className="font-mono text-[13px] text-slate-800 dark:text-slate-200 truncate" title={query.query_text}>
                              {query.query_text}
                            </div>
                            <div className="flex items-center gap-3 mt-1.5">
                              <span className={`text-[10px] uppercase font-mono font-bold tracking-wider ${statusText}`}>[{query.status}]</span>
                              <span className="text-[11px] font-mono text-slate-500 dark:text-slate-500">
                                {dateObj.toLocaleDateString() === new Date().toLocaleDateString() ? timeStr : dateObj.toLocaleDateString()}
                              </span>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            {/* Request Pipeline Widget */ }
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-none flex flex-col">
              <div className="flex justify-between items-center px-6 pt-5 pb-4 border-b border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-3">
                  <div className="w-2 h-6 bg-indigo-500"></div>
                  <div className="font-mono text-xs uppercase tracking-widest text-slate-900 dark:text-white font-bold">Request Pipeline</div>
                </div>
                <Link href="/dashboard/history" className="text-xs font-mono text-blue-600 dark:text-blue-400 hover:text-blue-500 hover:underline uppercase tracking-wider">View All →</Link>
              </div>
              <div className="flex flex-col flex-1 p-0">
                {myRequestsLoading ? (
                  <div className="py-12 flex justify-center"><Loading variant="spinner" size="sm" /></div>
                ) : myRequests.length === 0 ? (
                  <div className="py-12 flex flex-col items-center justify-center text-slate-500 dark:text-slate-400">
                    <div className="font-mono text-xs uppercase tracking-widest mb-2 opacity-50">[NO REQUESTS]</div>
                    <div className="text-sm">You have not submitted any write requests.</div>
                  </div>
                ) : (
                  <div className="flex flex-col">
                    {myRequests.map((request, idx) => {
                      const isPending = request.status === 'pending';
                      const isApproved = request.status === 'approved';
                      const isRejected = request.status === 'rejected';

                      return (
                        <div key={request.id} className={`flex flex-col gap-2 px-6 py-4 border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors ${idx === myRequests.length - 1 ? 'border-b-0' : ''}`}>
                          <div className="flex justify-between items-start">
                            <div className="flex items-center gap-2">
                               <span className="px-1.5 py-0.5 border border-slate-300 dark:border-slate-700 text-[10px] font-mono font-bold uppercase tracking-wider text-slate-900 dark:text-slate-300 rounded-none bg-slate-100 dark:bg-slate-800">
                                  {request.operation_type || 'WRITE'}
                              </span>
                              <div className="font-mono text-[13px] text-slate-800 dark:text-slate-200 truncate max-w-xs" title={request.query_text}>
                                {request.query_text}
                              </div>
                            </div>
                            <Link href={`/dashboard/approvals?id=${request.id}`} className="text-[11px] font-mono text-blue-600 dark:text-blue-400 hover:underline flex-shrink-0">
                               Details →
                            </Link>
                          </div>
                          
                          {/* Pipeline visualization */}
                          <div className="mt-2 flex items-center w-full gap-1">
                             <div className={`flex-1 h-1 ${isPending || isApproved || isRejected ? 'bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'}`}></div>
                             <div className={`flex-1 h-1 ${isApproved || isRejected ? (isApproved ? 'bg-green-500' : 'bg-red-500') : 'bg-slate-200 dark:bg-slate-700'}`}></div>
                             <div className={`flex-1 h-1 ${isApproved ? 'bg-green-500' : 'bg-slate-200 dark:bg-slate-700'}`}></div>
                          </div>
                          <div className="flex justify-between w-full mt-1 px-1">
                             <span className={`text-[9px] uppercase font-mono tracking-wider ${isPending || isApproved || isRejected ? 'text-blue-600 dark:text-blue-400 font-bold' : 'text-slate-400'}`}>Submitted</span>
                             <span className={`text-[9px] uppercase font-mono tracking-wider ${isApproved ? 'text-green-600 dark:text-green-400 font-bold' : isRejected ? 'text-red-600 dark:text-red-400 font-bold' : isPending ? 'text-blue-600 dark:text-blue-400 animate-pulse font-bold' : 'text-slate-400'}`}>Review</span>
                             <span className={`text-[9px] uppercase font-mono tracking-wider ${isApproved ? 'text-green-600 dark:text-green-400 font-bold' : 'text-slate-400'}`}>Executed</span>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            {/* Pending approvals breakdown */}
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-none flex flex-col h-full">
              <div className="flex justify-between items-center px-6 pt-5 pb-4 border-b border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-3">
                  <div className="w-2 h-6 bg-amber-500"></div>
                  <div className="font-mono text-xs uppercase tracking-widest text-slate-900 dark:text-white font-bold">Awaiting Action</div>
                </div>
                <Link href="/dashboard/approvals" className="text-xs font-mono text-amber-600 dark:text-amber-500 hover:text-amber-400 hover:underline uppercase tracking-wider">Queue →</Link>
              </div>
              <div className="flex-1">
                {pendingApprovalsLoading ? (
                  <div className="py-12 flex justify-center"><Loading variant="spinner" size="sm" /></div>
                ) : pendingApprovals.length === 0 ? (
                  <div className="py-12 flex flex-col items-center justify-center text-slate-500 dark:text-slate-400">
                    <div className="font-mono text-xs uppercase tracking-widest mb-2 opacity-50">[QUEUE EMPTY]</div>
                    <div className="text-[13px]">No pending actions required.</div>
                  </div>
                ) : (
                  <div className="flex flex-col">
                    {pendingApprovals.map((approval, idx) => {
                      const dateObj = new Date(approval.created_at);
                      const isToday = dateObj.toLocaleDateString() === new Date().toLocaleDateString();
                      const timeStr = isToday ? dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : dateObj.toLocaleDateString();

                      return (
                        <div key={approval.id} className={`flex flex-col px-6 py-4 border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors gap-2 ${idx === pendingApprovals.length - 1 ? 'border-b-0' : ''}`}>
                          <div className="flex items-center gap-2">
                             <span className="px-1.5 py-0.5 border border-slate-300 dark:border-slate-700 text-[10px] font-mono font-bold uppercase tracking-wider text-slate-900 dark:text-slate-300 rounded-none bg-slate-100 dark:bg-slate-800">{approval.operation_type}</span>
                             <span className="text-[11px] font-mono text-slate-500 dark:text-slate-400 whitespace-nowrap ml-auto">
                              {timeStr}
                            </span>
                          </div>
                          <span className="text-[13px] font-mono text-slate-700 dark:text-slate-300 overflow-hidden text-ellipsis whitespace-nowrap" title={approval.query_text}>
                            {approval.query_text}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

             {/* Admin Data source health */}
            {isAdmin && (
              <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-none flex flex-col h-full lg:col-span-1">
                <div className="flex justify-between items-center px-6 pt-5 pb-4 border-b border-slate-100 dark:border-slate-800">
                  <div className="flex items-center gap-3">
                    <div className="w-2 h-6 bg-emerald-500"></div>
                    <div className="font-mono text-xs uppercase tracking-widest text-slate-900 dark:text-white font-bold">Node Status</div>
                  </div>
                  <Link href="/admin/datasources" className="text-xs font-mono text-emerald-600 dark:text-emerald-500 hover:text-emerald-400 hover:underline uppercase tracking-wider">Nodes →</Link>
                </div>
                <div className="flex-1">
                  {dataSourcesLoading ? (
                    <div className="py-12 flex justify-center"><Loading variant="spinner" size="sm" /></div>
                  ) : dataSources.length === 0 ? (
                    <div className="py-12 flex flex-col items-center justify-center text-slate-500 dark:text-slate-400">
                      <div className="font-mono text-xs uppercase tracking-widest mb-2 opacity-50">[OFFLINE]</div>
                      <div className="text-[13px]">No target nodes available.</div>
                    </div>
                  ) : (
                    <div className="flex flex-col">
                      {dataSources.map((source, idx) => {
                        const isConnected = source.health?.status === 'healthy';
                        
                        return (
                          <div key={source.id} className={`flex items-center justify-between px-6 py-4 border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors ${idx === dataSources.length - 1 ? 'border-b-0' : ''}`}>
                            <div className="flex flex-col gap-1">
                              <span className="text-[13px] font-mono font-bold text-slate-900 dark:text-slate-100">{source.name}</span>
                              <span className="text-[10px] font-mono uppercase tracking-widest text-slate-500">{source.type}</span>
                            </div>
                            <div className={`flex items-center gap-1.5 px-2 py-1 border text-[10px] font-mono uppercase font-bold tracking-widest rounded-none ${isConnected ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-600 dark:text-emerald-400' : 'bg-rose-500/10 border-rose-500/20 text-rose-600 dark:text-rose-400'}`}>
                              {isConnected ? 'ONLINE' : 'ERROR'}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </AppLayout>
    </PageTransition>
  );
}
