'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { Query, ApprovalRequest } from '@/types';
import { formatDate } from '@/lib/utils';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Loading from '@/components/ui/Loading';
import { StaggerContainer, StaggerItem } from '@/components/layout/PageTransition';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { staggerContainer, staggerItem, fadeIn, springConfig, duration, reducedMotionVariants } from '@/lib/animations';
import {
  MagnifyingGlassIcon,
  ArrowTopRightOnSquareIcon,
  ArchiveBoxIcon,
  CircleStackIcon,
  BoltIcon
} from '@heroicons/react/24/outline';

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
  const shouldReduceMotion = useReducedMotion();

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      setPage(1);
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

        if (activeTab === 'all' || activeTab === 'reads') {
          const data = await apiClient.getQueryHistory(page, 20, debouncedSearch);
          fetchedQueries = data.queries;
          if (activeTab === 'reads') newTotal = data.total;
        }

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
            name: q.name || 'Ad-hoc Read',
            query_text: q.query_text,
            data_source_name: q.data_source_name || 'Generic Base',
            status: q.status,
            created_at: q.created_at,
            original: q
          });
        });

        fetchedApprovals.forEach(a => {
          items.push({
            id: a.id,
            type: 'write',
            name: a.operation_type ? `${a.operation_type.toUpperCase()} Protocol` : 'Write Cycle',
            query_text: a.query_text,
            data_source_name: a.data_source_name || 'Generic Base',
            status: a.status,
            created_at: a.created_at,
            operation_type: a.operation_type || 'UPDATE',
            original: a
          });
        });

        items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

        if (activeTab === 'all') {
          newTotal = items.length;
        }

        setHistoryItems(items);
        setTotal(newTotal);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Telemetry failure');
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
  }, [isAuthenticated, page, debouncedSearch, activeTab]);

  const getStatusStyle = (status: string) => {
    switch (status) {
      case 'completed':
      case 'approved':
        return 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20';
      case 'failed':
      case 'rejected':
        return 'bg-rose-500/10 text-rose-600 border-rose-500/20';
      case 'running':
        return 'bg-blue-500/10 text-blue-600 border-blue-500/20';
      case 'pending':
        return 'bg-amber-500/10 text-amber-600 border-amber-500/20';
      default:
        return 'bg-slate-500/10 text-slate-600 border-slate-500/20';
    }
  };

  if (isLoading) {
    return <div className="min-h-screen flex items-center justify-center"><Loading size="lg" /></div>;
  }

  if (!isAuthenticated) return null;

  const containerVariants = shouldReduceMotion ? reducedMotionVariants : staggerContainer;
  const itemVariants = shouldReduceMotion ? reducedMotionVariants : staggerItem;

  const tabs = ['all', 'reads', 'writes'] as const;
  const activeIndex = tabs.indexOf(activeTab);

  return (
    <div className="max-w-[1600px] mx-auto space-y-6 pb-6 px-4 md:px-6 h-full flex flex-col">

      <motion.div
        className="flex flex-col md:flex-row md:items-center justify-between gap-4 pt-4 shrink-0"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: duration.normal, ...springConfig.gentle }}
      >
        <motion.div
          className="flex items-center gap-2 p-1.5 glass rounded-2xl w-fit sleek-shadow relative"
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: duration.normal }}
        >
          {tabs.map((tab, index) => (
            <button
              key={tab}
              onClick={() => { setActiveTab(tab); setPage(1); }}
              className={`px-6 py-2 text-sm font-bold rounded-xl transition-colors duration-200 relative z-10 ${
                activeTab === tab
                  ? 'text-blue-600'
                  : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
              }`}
            >
              {tab === 'all' ? 'All Logs' : tab === 'reads' ? 'Reads' : 'Writes'}
            </button>
          ))}

          {!shouldReduceMotion && (
            <motion.div
              className="absolute top-1.5 bottom-1.5 bg-white dark:bg-slate-800 rounded-xl shadow-sm"
              initial={false}
              animate={{
                x: activeIndex * 96 + 6,
                width: 84,
              }}
              transition={{ duration: 0.25, ...springConfig.snappy }}
            />
          )}
        </motion.div>

        <motion.div
          className="relative w-full md:w-80 group"
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: duration.normal, delay: 0.1 }}
          whileFocus={{ scale: 1.02 }}
        >
          <MagnifyingGlassIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
          <input
            type="text"
            placeholder="Filter by query text or name..."
            className="w-full pl-12 pr-4 py-2.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-sm font-medium"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </motion.div>
      </motion.div>

      <motion.div
        className="flex-1 min-h-0 overflow-hidden"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: duration.slow, delay: 0.15 }}
      >
        <Card variant="default" className="border-none sleek-shadow h-full flex flex-col overflow-hidden">
          <AnimatePresence mode="wait">
            {loading ? (
              <motion.div
                key="loading"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="flex-1 flex items-center justify-center"
              >
                <Loading />
              </motion.div>
            ) : historyItems.length === 0 ? (
              <motion.div
                key="empty"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                className="flex-1 flex items-center justify-center"
              >
                <div className="text-center space-y-4">
                  <motion.div
                    initial={{ scale: 0.8, rotate: -10 }}
                    animate={{ scale: 1, rotate: 0 }}
                    transition={{ delay: 0.2, ...springConfig.bouncy }}
                  >
                    <ArchiveBoxIcon className="w-16 h-16 text-slate-200 mx-auto" />
                  </motion.div>
                  <motion.div
                    className="space-y-1"
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.3 }}
                  >
                    <h3 className="text-lg font-bold text-slate-400">Log Archive Empty</h3>
                    <p className="text-slate-400 text-sm font-medium">No results found for current telemetry filters.</p>
                  </motion.div>
                </div>
              </motion.div>
            ) : (
              <motion.div
                key="list"
                className="flex-1 overflow-y-auto scrollbar-minimal divide-y divide-slate-50 dark:divide-slate-800/50"
                variants={containerVariants}
                initial="initial"
                animate="animate"
              >
                {historyItems.map((item, index) => (
                  <motion.div
                    key={`${item.type}-${item.id}`}
                    variants={itemVariants}
                    className="p-5 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-all duration-300 group flex items-center gap-6"
                    whileHover={shouldReduceMotion ? {} : { x: 4 }}
                    transition={{ duration: 0.15 }}
                  >
                    <motion.div
                      className={`p-2.5 rounded-xl border shrink-0 ${item.type === 'read' ? 'bg-blue-500/10 border-blue-500/20 text-blue-600' : 'bg-amber-500/10 border-amber-500/20 text-amber-600'}`}
                      whileHover={shouldReduceMotion ? {} : { scale: 1.05 }}
                    >
                      {item.type === 'read' ? <MagnifyingGlassIcon className="w-5 h-5" /> : <BoltIcon className="w-5 h-5" />}
                    </motion.div>

                    <div className="flex-1 min-w-0 space-y-2">
                      <div className="flex items-center gap-3">
                        <span className="font-bold text-slate-800 dark:text-gray-100 truncate">
                          {item.name}
                        </span>
                        <span className={`text-[10px] font-black uppercase tracking-tighter px-2 py-0.5 rounded-lg border shrink-0 ${getStatusStyle(item.status)}`}>
                          {item.status}
                        </span>
                      </div>
                      <div className="font-mono text-sm text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900/50 px-3 py-2 rounded-xl border border-slate-100 dark:border-slate-800/50 truncate">
                        {item.query_text}
                      </div>
                    </div>

                    <div className="flex flex-col items-end gap-2 shrink-0 w-40">
                      <div className="flex items-center gap-1.5 text-xs text-slate-500 font-semibold uppercase">
                        <CircleStackIcon className="w-3.5 h-3.5 opacity-50" />
                        <span className="truncate max-w-[120px]">{item.data_source_name}</span>
                      </div>
                      <div className="text-xs text-slate-400 font-medium">
                        {formatDate(item.created_at)}
                      </div>
                      <Button
                        variant="secondary"
                        size="sm"
                        className="opacity-0 group-hover:opacity-100 mt-1"
                        onClick={() => {
                          if (item.type === 'read') {
                            router.push(`/dashboard/query?id=${item.id}`);
                          } else {
                            router.push(`/dashboard/approvals?id=${item.id}`);
                          }
                        }}
                      >
                        <ArrowTopRightOnSquareIcon className="w-4 h-4 mr-1.5" />
                        Teleport
                      </Button>
                    </div>
                  </motion.div>
                ))}
              </motion.div>
            )}
          </AnimatePresence>
        </Card>
      </motion.div>

      {total > historyItems.length && (
        <motion.div
          className="flex justify-center pt-4"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
        >
          <Button variant="outline" className="rounded-full px-12" onClick={() => setPage(p => p + 1)} loading={loading}>
            Load More Streams
          </Button>
        </motion.div>
      )}
    </div>
  );
}
