'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { apiClient } from '@/lib/api-client';
import type { DataSource } from '@/types';

interface DataSourceListProps {
  onSelectDataSource?: (dataSourceId: string) => void;
  onEditDataSource?: (dataSource: DataSource) => void;
  selectedId: string | null;
}

export default function DataSourceList({
  onSelectDataSource,
  onEditDataSource,
  selectedId,
}: DataSourceListProps) {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

  useEffect(() => {
    fetchDataSources();
  }, []);

  const fetchDataSources = async () => {
    try {
      setLoading(true);
      setError(null);
      const sources = await apiClient.getDataSources();
      setDataSources(sources);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data sources');
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async (id: string) => {
    try {
      setTestingId(id);
      await apiClient.testDataSourceConnection(id, {});
      toast.success('Connection successful! ✓', { duration: 5000 });
    } catch (err) {
      toast.error(`Connection failed: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    } finally {
      setTestingId(null);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete data source "${name}"?`)) {
      return;
    }

    try {
      await apiClient.deleteDataSource(id);
      setDataSources(dataSources.filter((ds) => ds.id !== id));
    } catch (err) {
      toast.error(`Failed to delete: ${err instanceof Error ? err.message : 'Unknown error'}`, { duration: 7000 });
    }
  };

  const getTypeBadgeColor = (type: string) => {
    return 'badge-slate';
  };

  const getHealthStatusColor = (ds: DataSource) => {
    if (!ds.is_active) {
      return 'badge-red text-red-700 bg-red-50';
    }
    return 'badge-green';
  };

  if (loading) {
    return (
      <div className="p-8 grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="animate-pulse bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-6 h-[200px]">
            <div className="flex gap-4">
              <div className="w-11 h-11 bg-gray-200 dark:bg-gray-700 rounded-xl"></div>
              <div className="flex-1 space-y-2 py-1">
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
                <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
              </div>
            </div>
            <div className="mt-6 h-3 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
            <div className="mt-4 h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 m-4">
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        <button
          onClick={fetchDataSources}
          className="mt-2 text-sm text-red-600 dark:text-red-400 underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        {dataSources.length === 0 ? (
          <div className="p-20 text-center glass rounded-3xl border border-slate-100 dark:border-slate-800/50">
            <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4 text-slate-400">
              <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4M0 12h18M0 12h18" />
              </svg>
            </div>
            <h3 className="text-lg font-bold text-slate-800 dark:text-white">No Infrastructure Bridges</h3>
            <p className="text-slate-500 text-sm mt-1">Connect your first database to enable query execution.</p>
          </div>
        ) : (
          <div className="grid gap-3">
            {dataSources.map((dataSource) => {
              const isPg = dataSource.type === 'postgresql';
              const isMysql = dataSource.type === 'mysql';
              
              let iconClass = 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300';
              if (isPg) iconClass = 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20';
              if (isMysql) iconClass = 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20';

              return (
                <div 
                  key={dataSource.id} 
                  className="group p-5 glass rounded-3xl border border-white/50 dark:border-slate-800/50 hover:border-blue-500/30 hover:bg-white dark:hover:bg-slate-800/50 transition-all duration-300 sleek-shadow flex flex-col md:flex-row md:items-center justify-between gap-4"
                >
                  <div className="flex items-center gap-4">
                    <div className={`w-14 h-14 rounded-2xl flex items-center justify-center text-2xl border shadow-sm transition-transform group-hover:scale-105 ${iconClass}`}>
                      {isPg ? <PgIcon /> : isMysql ? <MysqlIcon /> : <DbIcon />}
                    </div>
                    <div>
                      <div className="flex items-center gap-3">
                        <span className="font-bold text-slate-900 dark:text-white text-lg">
                          {dataSource.name}
                        </span>
                        <span className={`text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border ${iconClass}`}>
                          {dataSource.type}
                        </span>
                        <span className={`text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-lg border ${dataSource.is_active ? 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20' : 'bg-slate-500/10 text-slate-500'}`}>
                          {dataSource.is_active ? 'Online' : 'Offline'}
                        </span>
                      </div>
                      <div className="text-sm text-slate-500 dark:text-slate-400 font-mono mt-0.5 opacity-80">
                         {dataSource.host}:{dataSource.port}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300">
                    <button
                      onClick={() => handleTestConnection(dataSource.id)}
                      disabled={testingId === dataSource.id}
                      className="h-10 px-6 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-bold text-xs hover:bg-emerald-500 hover:text-white transition-all shadow-sm disabled:opacity-50"
                    >
                      {testingId === dataSource.id ? 'Pinging...' : 'Pulse Test'}
                    </button>
                    {onEditDataSource && (
                      <button
                        onClick={() => onEditDataSource(dataSource)}
                        className="h-10 px-6 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 font-bold text-xs hover:bg-blue-500 hover:text-white transition-all shadow-sm"
                      >
                        Modify
                      </button>
                    )}
                    <button
                      onClick={() => handleDelete(dataSource.id, dataSource.name)}
                      className="h-10 px-10 rounded-xl bg-rose-500/10 text-rose-600 dark:text-rose-400 font-bold text-xs hover:bg-rose-500 hover:text-white transition-all shadow-sm ml-2"
                      title="Decommission Source"
                    >
                       Purge
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

// Minimal SVG Icons for Databases
function PgIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12.022 2.1c-5.467 0-9.878 4.41-9.878 9.88 0 5.467 4.411 9.878 9.878 9.878 5.467 0 9.878-4.411 9.878-9.878 0-5.47-4.411-9.88-9.878-9.88zm3.623 14.88c-.5.441-1.294.67-2.323.67H9.255v1.233h-1.97v-8.498h6.143c.97 0 1.705.235 2.176.64.441.41.676.97.676 1.734.025.794-.236 1.352-.647 1.734-.383.353-.941.529-1.646.529h-3.41v1.94h4.41v-3.41h1.56v3.438zM8.344 8.784h4.086c.764 0 1.352.176 1.735.47.382.264.588.675.588 1.146 0 .47-.206.882-.588 1.205-.383.294-.971.47-1.735.47H8.344v-3.29z"/>
    </svg>
  );
}

function MysqlIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12.002 1.357c-5.836 0-10.56 4.721-10.56 10.551 0 5.828 4.724 10.55 10.56 10.55 5.833 0 10.554-4.722 10.554-10.55 0-5.83-4.721-10.551-10.554-10.551zm1.264 6.646c1.373 0 1.956.892 1.956 1.94 0 1.135-.615 1.983-1.897 1.983-.756 0-1.284-.336-1.574-.75h-.06v3.593H9.863V8.11h1.616v.612h.063c.27-.406.804-.719 1.724-.719zm-1.028 1.48c-.68 0-1.042.47-1.042 1.055 0 .564.364 1.026 1.053 1.026.685 0 1.05-.44 1.05-1.042 0-.616-.365-1.038-1.06-1.038zM5.385 15.65h1.83V9.752H5.385v5.898zm.9-8c-.658 0-1.127.47-1.127 1.114 0 .647.469 1.116 1.127 1.116.634 0 1.077-.47 1.077-1.116 0-.645-.443-1.114-1.076-1.114zm11.332 7.822c-1.375 0-1.958-.893-1.958-1.94 0-1.133.616-1.984 1.898-1.984.757 0 1.285.337 1.575.75h.06V8.11h1.828v7.362h-1.616v-.613h-.063c-.27.406-.804.72-1.724.72zm1.028-1.48c.682 0 1.043-.466 1.043-1.053 0-.564-.361-1.027-1.05-1.027-.685 0-1.05.441-1.05 1.043 0 .615.365 1.037 1.058 1.037z"/>
    </svg>
  );
}

function DbIcon() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <ellipse cx="12" cy="5" rx="9" ry="3"/>
      <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
      <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
    </svg>
  );
}
