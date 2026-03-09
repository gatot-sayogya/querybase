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
import PageTransition from '@/components/layout/PageTransition';
import { 
  MagnifyingGlassIcon, 
  FunnelIcon, 
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

  // Debounce search
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

  return (
    <PageTransition animation="fade">
      <div className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6">
        
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 pt-4">
          <div className="space-y-1">
            <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
              Execution History
            </h1>
            <p className="text-slate-500 dark:text-slate-400 font-medium">
              A comprehensive log of all system queries and state changes.
            </p>
          </div>
          
          <div className="relative w-full md:w-96 group">
            <MagnifyingGlassIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
            <input 
              type="text" 
              placeholder="Filter by query text or name..."
              className="w-full pl-12 pr-4 py-2.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-sm font-medium"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {/* Tab Control */}
        <div className="flex items-center gap-2 p-1.5 glass rounded-2xl w-fit sleek-shadow">
          {(['all', 'reads', 'writes'] as const).map(tab => (
            <button
              key={tab}
              onClick={() => { setActiveTab(tab); setPage(1); }}
              className={`px-6 py-2 text-sm font-bold rounded-xl transition-all duration-300 ${
                activeTab === tab 
                  ? 'bg-white dark:bg-slate-800 text-blue-600 shadow-sm' 
                  : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
              }`}
            >
              {tab === 'all' ? 'All Logs' : tab === 'reads' ? 'Reads' : 'Writes'}
            </button>
          ))}
        </div>

        {/* List Content */}
        <Card variant="default" className="border-none sleek-shadow overflow-hidden">
          {loading ? (
            <div className="p-20 flex justify-center"><Loading /></div>
          ) : historyItems.length === 0 ? (
            <div className="p-32 text-center space-y-4">
              <ArchiveBoxIcon className="w-16 h-16 text-slate-200 mx-auto" />
              <div className="space-y-1">
                <h3 className="text-lg font-bold text-slate-400">Log Archive Empty</h3>
                <p className="text-slate-400 text-sm font-medium">No results found for current telemetry filters.</p>
              </div>
            </div>
          ) : (
            <div className="divide-y divide-slate-50 dark:divide-slate-800/50">
               {historyItems.map((item) => (
                 <div 
                   key={`${item.type}-${item.id}`} 
                   className="p-6 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-all duration-300 group flex flex-col md:flex-row md:items-center justify-between gap-6"
                 >
                   <div className="space-y-4 flex-1 min-w-0">
                     <div className="flex items-center gap-4">
                        <div className={`p-2 rounded-xl border ${item.type === 'read' ? 'bg-blue-500/10 border-blue-500/20 text-blue-600' : 'bg-amber-500/10 border-amber-500/20 text-amber-600'}`}>
                           {item.type === 'read' ? <MagnifyingGlassIcon className="w-5 h-5" /> : <BoltIcon className="w-5 h-5" />}
                        </div>
                        <div>
                           <div className="font-bold text-slate-800 dark:text-gray-100 flex items-center gap-3">
                             {item.name}
                             <span className={`text-[10px] font-black uppercase tracking-tighter px-2 py-0.5 rounded-lg border ${getStatusStyle(item.status)}`}>
                               {item.status}
                             </span>
                           </div>
                           <div className="flex items-center gap-4 mt-1">
                              <div className="flex items-center gap-1.5 text-xs text-slate-500 font-semibold uppercase">
                                 <CircleStackIcon className="w-4 h-4 opacity-40" />
                                 {item.data_source_name}
                              </div>
                              <span className="text-slate-300 dark:text-slate-700">•</span>
                              <div className="text-xs text-slate-500 font-semibold uppercase">
                                 {formatDate(item.created_at)}
                              </div>
                           </div>
                        </div>
                     </div>
                     <div className="font-mono text-sm text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900/50 p-4 rounded-2xl border border-slate-100 dark:border-slate-800/50 truncate group-hover:bg-blue-500/5 transition-colors">
                        {item.query_text}
                     </div>
                   </div>

                   <div className="flex items-center gap-3 self-end md:self-center">
                     <Button 
                       variant="secondary" 
                       size="sm" 
                       className="opacity-0 group-hover:opacity-100"
                       onClick={() => {
                         if (item.type === 'read') {
                            router.push(`/dashboard/query?id=${item.id}`);
                         } else {
                            router.push(`/dashboard/approvals?id=${item.id}`);
                         }
                       }}
                     >
                       <ArrowTopRightOnSquareIcon className="w-4 h-4 mr-2" />
                       Teleport
                     </Button>
                   </div>
                 </div>
               ))}
            </div>
          )}
        </Card>

        {/* Pagination placeholder */}
        {total > historyItems.length && (
          <div className="flex justify-center pt-4">
             <Button variant="outline" className="rounded-full px-12" onClick={() => setPage(p => p + 1)} loading={loading}>
               Load More Streams
             </Button>
          </div>
        )}
      </div>
    </PageTransition>
  );
}
