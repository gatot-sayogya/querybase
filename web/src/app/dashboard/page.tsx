'use client';

import { useEffect, useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import AppLayout from '@/components/layout/AppLayout';
import Loading from '@/components/ui/Loading';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import { StaggerContainer, StaggerItem } from '@/components/layout/PageTransition';
import Link from 'next/link';
import { useDashboardStats } from '@/hooks/useDashboardStats';
import { apiClient } from '@/lib/api-client';
import type { Query, ApprovalRequest, DataSource, HealthStatus } from '@/types';
import { motion, useReducedMotion } from 'framer-motion';
import { staggerContainerSlow, staggerItem, springConfig, duration, reducedMotionVariants } from '@/lib/animations';
import {
  PlusIcon,
  ClockIcon,
  CircleStackIcon,
  ArrowRightIcon,
  CheckCircleIcon,
  XCircleIcon,
  ExclamationCircleIcon,
  BoltIcon
} from '@heroicons/react/24/outline';

function AnimatedCounter({ value, loading }: { value: number; loading: boolean }) {
  const [displayValue, setDisplayValue] = useState(0);
  const shouldReduceMotion = useReducedMotion();

  useEffect(() => {
    if (loading || shouldReduceMotion) {
      setDisplayValue(value);
      return;
    }

    const duration = 800;
    const steps = 30;
    const increment = value / steps;
    let current = 0;
    let step = 0;

    const timer = setInterval(() => {
      step++;
      current = Math.min(Math.round(increment * step), value);
      setDisplayValue(current);
      if (step >= steps) clearInterval(timer);
    }, duration / steps);

    return () => clearInterval(timer);
  }, [value, loading, shouldReduceMotion]);

  if (loading) return <span>...</span>;
  return <span>{displayValue}</span>;
}

export default function DashboardPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading, isHydrating } = useAuthStore();
  const { stats, isLoading: statsLoading } = useDashboardStats();
  const shouldReduceMotion = useReducedMotion();

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

    setRecentQueriesLoading(true);
    setPendingApprovalsLoading(true);
    setMyRequestsLoading(true);
    setDataSourcesLoading(true);

    apiClient.getQueryHistory(1, 6)
      .then(res => setRecentQueries(res.queries))
      .catch(err => console.error("Recent queries error:", err))
      .finally(() => setRecentQueriesLoading(false));

    apiClient.getApprovals({ status: 'pending', page: 1 })
      .then(res => setPendingApprovals(res.slice(0, 4)))
      .catch(err => console.error("Pending approvals error:", err))
      .finally(() => setPendingApprovalsLoading(false));

    apiClient.getApprovals({ page: 1 })
      .then(res => setMyRequests(res.slice(0, 4)))
      .catch(err => console.error("My requests error:", err))
      .finally(() => setMyRequestsLoading(false));

    apiClient.getApprovalCounts()
      .then(res => setApprovalCounts(res))
      .catch(err => console.error("Counts error:", err));

    if (user?.role === 'admin') {
      apiClient.getDataSources()
        .then(async (sources) => {
          const topSources = sources.slice(0, 4);
          const sourcesWithHealth = await Promise.all(
            topSources.map(async (source) => {
              try {
                const health = await apiClient.getDataSourceHealth(source.id);
                return { ...source, health };
              } catch {
                return {
                  ...source,
                  health: { status: 'unhealthy', latency_ms: 0, last_checked: new Date().toISOString(), data_source_id: source.id, message: 'Offline' } as HealthStatus
                };
              }
            })
          );
          setDataSources(sourcesWithHealth);
        })
        .finally(() => setDataSourcesLoading(false));
    } else {
      setDataSourcesLoading(false);
    }
  }, [isAuthenticated, isHydrating, user?.role]);

  if (isHydrating || isLoading) {
    return <div className="min-h-screen flex items-center justify-center"><Loading variant="spinner" size="lg" /></div>;
  }

  if (!isAuthenticated) return null;
  const isAdmin = user?.role === 'admin';

  const containerVariants = shouldReduceMotion ? reducedMotionVariants : staggerContainerSlow;
  const itemVariants = shouldReduceMotion ? reducedMotionVariants : staggerItem;

  return (
    <AppLayout>
      <div className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6">
        <motion.div
          className="flex flex-col md:flex-row md:items-end justify-between gap-6 pt-4"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: duration.normal, ...springConfig.gentle }}
        >
          <div className="space-y-1">
            <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
              Welcome back, {user?.username}
            </h1>
            <p className="text-slate-500 dark:text-slate-400 font-medium">
              System status and telemetry is healthy.
            </p>
          </div>

          <Link href="/dashboard/query">
            <Button size="lg" className="w-full md:w-auto gap-2 group">
              <PlusIcon className="w-5 h-5 group-hover:rotate-90 transition-transform duration-300" />
              New Query
            </Button>
          </Link>
        </motion.div>

        <motion.div
          className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5 gap-4"
          variants={containerVariants}
          initial="initial"
          animate="animate"
        >
          <motion.div variants={itemVariants}>
            <Card variant="glass" className="p-6 space-y-4 h-full">
              <div className="flex justify-between items-center text-blue-600">
                <ClockIcon className="w-6 h-6" />
                <span className="text-xs font-bold uppercase tracking-wider opacity-60">Today</span>
              </div>
              <div>
                <div className="text-3xl font-bold">
                  <AnimatedCounter value={stats?.my_queries_today || 0} loading={statsLoading} />
                </div>
                <div className="text-xs text-slate-500 font-semibold uppercase">Executed Queries</div>
              </div>
            </Card>
          </motion.div>

          <motion.div variants={itemVariants}>
            <Card variant="glass" className="p-6 space-y-4 h-full">
              <div className="flex justify-between items-center text-emerald-500">
                <CircleStackIcon className="w-6 h-6" />
                <span className="text-xs font-bold uppercase tracking-wider opacity-60">Security</span>
              </div>
              <div>
                <div className="text-3xl font-bold">
                  <AnimatedCounter value={stats?.db_access_count || 0} loading={statsLoading} />
                </div>
                <div className="text-xs text-slate-500 font-semibold uppercase">Authorized Bases</div>
              </div>
            </Card>
          </motion.div>

          <motion.div variants={itemVariants}>
            <Card variant="glass" className="p-6 space-y-4 border-l-4 border-l-amber-500/50 h-full">
              <div className="flex justify-between items-center text-amber-500">
                <BoltIcon className="w-6 h-6" />
                <span className="text-xs font-bold uppercase tracking-wider opacity-60">Pipeline</span>
              </div>
              <div>
                <div className="text-3xl font-bold">
                  <AnimatedCounter value={stats?.pending_approvals || 0} loading={statsLoading} />
                </div>
                <div className="text-xs text-slate-500 font-semibold uppercase">Pending Reviews</div>
              </div>
            </Card>
          </motion.div>

          <motion.div variants={itemVariants}>
            <Card variant="glass" className="p-6 space-y-4 h-full">
              <div className="flex justify-between items-center text-indigo-500">
                <CheckCircleIcon className="w-6 h-6" />
                <span className="text-xs font-bold uppercase tracking-wider opacity-60">Status</span>
              </div>
              <div>
                <div className="text-3xl font-bold">
                  <AnimatedCounter value={approvalCounts.approved || 0} loading={false} />
                </div>
                <div className="text-xs text-slate-500 font-semibold uppercase">Completed Cycles</div>
              </div>
            </Card>
          </motion.div>

          {isAdmin && (
            <motion.div variants={itemVariants}>
              <Card variant="default" className="p-6 space-y-4 bg-slate-900 border-slate-800 text-white h-full">
                <div className="flex justify-between items-center opacity-50">
                  <ExclamationCircleIcon className="w-6 h-6" />
                  <span className="text-xs font-bold uppercase tracking-wider">Health</span>
                </div>
                <div>
                  <div className="text-3xl font-bold">
                    <AnimatedCounter value={stats?.total_users || 0} loading={statsLoading} />
                  </div>
                  <div className="text-xs opacity-50 font-semibold uppercase">Active Personnel</div>
                </div>
              </Card>
            </motion.div>
          )}
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <motion.div
            className="lg:col-span-8 space-y-6"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: duration.slow, delay: 0.2, ...springConfig.gentle }}
          >
            <Card variant="default" className="flex flex-col border-none sleek-shadow">
              <div className="flex justify-between items-center p-6 border-b border-slate-50 dark:border-slate-800/50">
                <h2 className="text-lg font-bold flex items-center gap-2">
                  <ClockIcon className="w-5 h-5 text-blue-500" />
                  Query Streams
                </h2>
                <Link href="/dashboard/history">
                  <Button variant="ghost" size="sm" className="gap-2">
                    Full Logs <ArrowRightIcon className="w-4 h-4" />
                  </Button>
                </Link>
              </div>
              <div className="p-0">
                {recentQueriesLoading ? (
                  <div className="py-20 text-center"><Loading variant="spinner" /></div>
                ) : recentQueries.length === 0 ? (
                  <div className="py-20 text-center text-slate-500 font-medium">No recent streams detected.</div>
                ) : (
                  <motion.div
                    className="divide-y divide-slate-50 dark:divide-slate-800/50"
                    variants={containerVariants}
                    initial="initial"
                    animate="animate"
                  >
                    {recentQueries.map((query, index) => (
                      <motion.div
                        key={query.id}
                        variants={itemVariants}
                        className="p-6 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors flex items-center justify-between group"
                      >
                        <div className="flex-1 min-w-0 pr-4">
                          <div className="font-mono text-sm text-slate-700 dark:text-slate-300 truncate font-semibold">
                            {query.query_text}
                          </div>
                          <div className="flex items-center gap-4 mt-2">
                            <span className={`text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full ${
                              query.status === 'completed' ? 'bg-emerald-500/10 text-emerald-600' :
                              query.status === 'failed' ? 'bg-red-500/10 text-red-600' : 'bg-blue-500/10 text-blue-600'
                            }`}>
                              {query.status}
                            </span>
                            <span className="text-[11px] text-slate-400 font-medium">
                              {new Date(query.created_at).toLocaleString()}
                            </span>
                          </div>
                        </div>
                        <Link href="/dashboard/query">
                          <Button variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100 transition-opacity">
                            Run Again
                          </Button>
                        </Link>
                      </motion.div>
                    ))}
                  </motion.div>
                )}
              </div>
            </Card>
          </motion.div>

          <motion.div
            className="lg:col-span-4 space-y-6"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: duration.slow, delay: 0.3, ...springConfig.gentle }}
          >
            <Card variant="glass" className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h3 className="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  Pipeline
                </h3>
                <Link href="/dashboard/history" className="text-xs font-bold text-blue-600 hover:underline px-2 py-1 bg-blue-500/10 rounded-lg">
                  View All
                </Link>
              </div>

              <motion.div
                className="space-y-4"
                variants={containerVariants}
                initial="initial"
                animate="animate"
              >
                {myRequestsLoading ? (
                  <Loading size="sm" />
                ) : myRequests.length === 0 ? (
                  <div className="py-8 text-center text-xs text-slate-500">No active cycles.</div>
                ) : (
                  myRequests.map((request) => (
                    <motion.div key={request.id} variants={itemVariants}>
                      <Link href={`/dashboard/approvals?id=${request.id}`} className="block group">
                        <Card
                          variant="interactive"
                          padding="sm"
                          className="border-slate-100 dark:border-slate-800/50 bg-white/40 dark:bg-slate-800/40 shadow-none"
                        >
                          <div className="flex justify-between items-start mb-2">
                            <span className="text-[10px] font-black uppercase tracking-tighter px-1.5 py-0.5 bg-slate-900 text-white rounded">
                              {request.operation_type || 'WRITE'}
                            </span>
                            <motion.div
                              className={`w-2 h-2 rounded-full ${
                                request.status === 'pending' ? 'bg-amber-400' :
                                request.status === 'approved' ? 'bg-emerald-500' : 'bg-rose-500'
                              }`}
                              animate={request.status === 'pending' && !shouldReduceMotion ? {
                                scale: [1, 1.2, 1],
                                opacity: [1, 0.7, 1],
                              } : {}}
                              transition={{ duration: 2, repeat: Infinity }}
                            />
                          </div>
                          <div className="text-xs font-mono font-bold text-slate-700 dark:text-slate-300 truncate mb-2">
                            {request.query_text}
                          </div>
                          <div className="flex gap-1">
                            <div className={`h-1 flex-1 rounded-full transition-colors duration-300 ${request.status === 'pending' || request.status === 'approved' || request.status === 'rejected' ? 'bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'}`} />
                            <div className={`h-1 flex-1 rounded-full transition-colors duration-300 ${request.status === 'approved' || request.status === 'rejected' ? (request.status === 'approved' ? 'bg-emerald-500' : 'bg-rose-500') : 'bg-slate-200 dark:bg-slate-700'}`} />
                            <div className={`h-1 flex-1 rounded-full transition-colors duration-300 ${request.status === 'approved' ? 'bg-emerald-500' : 'bg-slate-200 dark:bg-slate-700'}`} />
                          </div>
                        </Card>
                      </Link>
                    </motion.div>
                  ))
                )}
              </motion.div>
            </Card>

            {isAdmin && (
              <Card variant="default" className="p-6 bg-slate-50 dark:bg-slate-900/50 border-none shadow-none">
                <h3 className="font-bold text-slate-900 dark:text-white mb-6">Network Health</h3>
                <motion.div
                  className="space-y-4"
                  variants={containerVariants}
                  initial="initial"
                  animate="animate"
                >
                  {dataSourcesLoading ? (
                    <Loading size="sm" />
                  ) : dataSources.length === 0 ? (
                    <div className="text-xs text-slate-500">Nodes detached.</div>
                  ) : (
                    dataSources.map((source) => (
                      <motion.div key={source.id} variants={itemVariants}>
                        <div className="flex items-center justify-between p-3 rounded-2xl bg-white dark:bg-slate-900 border border-slate-100 dark:border-slate-800/80 sleek-shadow">
                          <div className="flex flex-col">
                            <span className="text-sm font-bold">{source.name}</span>
                            <span className="text-[10px] font-black opacity-30 uppercase tracking-widest">{source.type}</span>
                          </div>
                          <motion.div
                            className={`w-3 h-3 rounded-full ${source.health?.status === 'healthy' ? 'bg-emerald-500' : 'bg-rose-500'}`}
                            animate={!shouldReduceMotion && source.health?.status === 'healthy' ? {
                              boxShadow: [
                                '0 0 0 0 rgba(16, 185, 129, 0.4)',
                                '0 0 0 8px rgba(16, 185, 129, 0)',
                              ],
                            } : {}}
                            transition={{ duration: 2, repeat: Infinity }}
                            style={source.health?.status === 'healthy' ? { boxShadow: '0 0 8px rgba(16, 185, 129, 0.5)' } : {}}
                          />
                        </div>
                      </motion.div>
                    ))
                  )}
                </motion.div>
              </Card>
            )}
          </motion.div>
        </div>
      </div>
    </AppLayout>
  );
}
