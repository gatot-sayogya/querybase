'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import toast from 'react-hot-toast';
import { motion, AnimatePresence } from 'framer-motion';
import { useAuthStore } from '@/stores/auth-store';
import { apiClient } from '@/lib/api-client';
import DataSourceSchemaSelector from './DataSourceSchemaSelector';
import QueryResults from './QueryResults';
import Button from '@/components/ui/Button';
import Loading from '@/components/ui/Loading';
import { QueryError } from '@/components/ui/Alert';
import type { QueryResult, WriteQueryPreview } from '@/types';
import dynamic from 'next/dynamic';
import WritePreviewModal from './WritePreviewModal';
import { springConfig, staggerContainer, staggerItem } from '@/lib/animations';

const SQLEditor = dynamic(() => import('./SQLEditor'), { ssr: false });

export default function QueryExecutor() {
  const router = useRouter();
  const { user, isAuthenticated } = useAuthStore();

  const [dataSourceId, setDataSourceId] = useState('');
  const [queryText, setQueryText] = useState('');
  const [queryId, setQueryId] = useState<string | null>(null);
  const [results, setResults] = useState<QueryResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [rowLimit, setRowLimit] = useState(1000);
  const [canWrite, setCanWrite] = useState(false);
  const [isWriteQuery, setIsWriteQuery] = useState(false);
  const [permissionError, setPermissionError] = useState<{ message: string; hint: string; dataSource: string } | null>(null);
  const [writePreview, setWritePreview] = useState<WriteQueryPreview | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [pendingWriteQuery, setPendingWriteQuery] = useState<string>('');

  // Resizing state
  const [isSidebarResizing, setIsSidebarResizing] = useState(false);
  const [sidebarWidth, setSidebarWidth] = useState(260);
  const [isEditorResizing, setIsEditorResizing] = useState(false);
  const [editorHeight, setEditorHeight] = useState(280);
  const [isFullscreenResults, setIsFullscreenResults] = useState(false);
  
  const containerRef = useRef<HTMLDivElement>(null);
  const resizerRef = useRef<HTMLDivElement>(null);

  const startSidebarResizing = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsSidebarResizing(true);
  }, []);

  const stopResizing = useCallback(() => {
    setIsSidebarResizing(false);
    setIsEditorResizing(false);
  }, []);

  const startEditorResizing = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsEditorResizing(true);
  }, []);

  const resize = useCallback((e: MouseEvent) => {
    if (isSidebarResizing) {
      const newWidth = e.clientX - 16;
      if (newWidth > 180 && newWidth < 600) {
        setSidebarWidth(newWidth);
      }
    } else if (isEditorResizing && containerRef.current) {
      const containerRect = containerRef.current.getBoundingClientRect();
      const newHeight = e.clientY - containerRect.top - 100; // Offset for header + padding
      if (newHeight > 100 && newHeight < containerRect.height - 100) {
        setEditorHeight(newHeight);
      }
    }
  }, [isSidebarResizing, isEditorResizing]);

  useEffect(() => {
    if (isSidebarResizing || isEditorResizing) {
      window.addEventListener('mousemove', resize);
      window.addEventListener('mouseup', stopResizing);
    } else {
      window.removeEventListener('mousemove', resize);
      window.removeEventListener('mouseup', stopResizing);
    }
    return () => {
      window.removeEventListener('mousemove', resize);
      window.removeEventListener('mouseup', stopResizing);
    };
  }, [isSidebarResizing, isEditorResizing, resize, stopResizing]);

  // Redirect if not authenticated
  if (!isAuthenticated) {
    router.push('/login');
    return null;
  }

  const handleExecuteQuery = async () => {
    if (!dataSourceId) {
      setError('Please select a data source');
      return;
    }

    if (!queryText.trim()) {
      setError('Please enter a SQL query');
      return;
    }

    setLoading(true);
    setError(null);
    setPermissionError(null);
    setResults(null);
    setQueryId(null);

    try {
      // Add LIMIT if not present and query is SELECT
      let finalQuery = queryText.trim();
      const isSelectQuery = /^\s*SELECT\s/i.test(finalQuery);
      const isDeleteQuery = /^\s*DELETE\s/i.test(finalQuery);
      const isUpdateQuery = /^\s*UPDATE\s/i.test(finalQuery);
      const hasLimit = /\bLIMIT\s+\d+\s*$/i.test(finalQuery);

      // For DELETE/UPDATE queries with write permission, show preview first
      if ((isDeleteQuery || isUpdateQuery) && canWrite) {
        setPreviewLoading(true);
        try {
          const preview = await apiClient.previewWriteQuery(dataSourceId, finalQuery);
          setWritePreview(preview);
          setPendingWriteQuery(finalQuery);
        } catch (previewErr: any) {
          // If preview fails, fall through to normal execution (which will create approval)
          setError(previewErr.response?.data?.error || previewErr.message || 'Failed to preview query');
        } finally {
          setPreviewLoading(false);
          setLoading(false);
        }
        return;
      }

      if (isSelectQuery && !hasLimit && rowLimit > 0) {
        finalQuery += ` LIMIT ${rowLimit}`;
      }

      // Execute query
      const response = await apiClient.executeQuery({
        data_source_id: dataSourceId,
        query_text: finalQuery,
      });

      console.log('Query response:', response);

      // Backend returns query_id (not id) in ExecuteQueryResponse
      const qid = (response as any).query_id || response.id;

      if (!qid) {
        console.error('Invalid query response - missing ID:', response);
        throw new Error('Server did not return a query ID. Please check the backend logs.');
      }

      setQueryId(qid);

      // Check if results are included in the response
      const hasData = (response as any).data && Array.isArray((response as any).data);
      const hasColumns = (response as any).columns && Array.isArray((response as any).columns);

      // Poll for results if query is still running
      if (response.status === 'running' || response.status === 'pending') {
        pollForResult(qid);
      } else if (response.status === 'completed' && hasData && hasColumns) {
        // Use the results from the response directly
        setResults({
          query_id: qid,
          row_count: (response as any).row_count || 0,
          columns: (response as any).columns,
          data: (response as any).data,
        });
        setLoading(false);
      } else if (response.status === 'completed') {
        // Fetch results from server
        const queryWithResults = await apiClient.getQuery(qid);
        if (queryWithResults.results) {
          setResults(queryWithResults.results);
        }
        setLoading(false);
      } else if (response.status === 'failed') {
        setError(response.error_message || 'Query execution failed');
        setLoading(false);
      } else {
        setLoading(false);
      }
    } catch (err: any) {
      console.error('Query execution error:', err);
      if (err.response?.data?.code === 'PERMISSION_DENIED_WRITE') {
        const data = err.response.data;
        setPermissionError({
          message: data.error,
          hint: data.hint,
          dataSource: data.data_source
        });
        setError(null);
      } else {
        setError(err.response?.data?.error || err.message || 'Failed to execute query');
        setPermissionError(null);
      }
      setLoading(false);
    }
  };

  const pollForResult = async (id: string) => {
    if (!id) {
      console.error('pollForResult called with invalid ID:', id);
      setError('Invalid query ID');
      setLoading(false);
      return;
    }

    const maxAttempts = 30; // 30 seconds max
    let attempts = 0;

    const interval = setInterval(async () => {
      attempts++;

      try {
        const query = await apiClient.getQuery(id);

        if (query.status === 'completed') {
          clearInterval(interval);
          if (query.results) {
            setResults(query.results);
          }
          setLoading(false);
        } else if (query.status === 'failed') {
          clearInterval(interval);
          setError(query.error_message || 'Query execution failed');
          setLoading(false);
        } else if (attempts >= maxAttempts) {
          clearInterval(interval);
          setError('Query execution timed out');
          setLoading(false);
        }
      } catch (err) {
        clearInterval(interval);
        console.error('Poll result error:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch query status');
        setLoading(false);
      }
    }, 1000);
  };

  const handleSaveQuery = async () => {
    if (!dataSourceId || !queryText.trim()) {
      setError('Please select a data source and enter a query');
      return;
    }

    try {
      await apiClient.saveQuery({
        data_source_id: dataSourceId,
        query_text: queryText.trim(),
        name: 'Saved Query',
        description: `Query executed on ${new Date().toLocaleString()}`,
      });
      toast.success('Query saved successfully!', { duration: 5000 });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save query');
    }
  };

  const handleExportCSV = async () => {
    if (!queryId) return;

    try {
      const blob = await apiClient.exportQuery(queryId, 'csv');
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `query-${queryId}-${Date.now()}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export query');
    }
  };

  const handleExportJSON = async () => {
    if (!queryId) return;

    try {
      const blob = await apiClient.exportQuery(queryId, 'json');
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `query-${queryId}-${Date.now()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export query');
    }
  };

  const handleTableSelect = async (tableName: string) => {
    if (!dataSourceId) {
      setError('Please select a data source first');
      return;
    }

    // Set the query text and execute it
    const query = `SELECT * FROM ${tableName} LIMIT 100`;
    setQueryText(query);

    // Execute the query
    setLoading(true);
    setError(null);
    setPermissionError(null);
    setResults(null);
    setQueryId(null);

    try {
      const response = await apiClient.executeQuery({
        data_source_id: dataSourceId,
        query_text: query,
      });

      console.log('Query response:', response);

      const qid = (response as any).query_id || response.id;

      if (!qid) {
        console.error('Invalid query response - missing ID:', response);
        throw new Error('Server did not return a query ID. Please check the backend logs.');
      }

      setQueryId(qid);

      const hasData = (response as any).data && Array.isArray((response as any).data);
      const hasColumns = (response as any).columns && Array.isArray((response as any).columns);

      if (response.status === 'running' || response.status === 'pending') {
        pollForResult(qid);
      } else if (response.status === 'completed' && hasData && hasColumns) {
        setResults({
          query_id: qid,
          row_count: (response as any).row_count || 0,
          columns: (response as any).columns,
          data: (response as any).data,
        });
        setLoading(false);
      } else if (response.status === 'completed') {
        const queryWithResults = await apiClient.getQuery(qid);
        if (queryWithResults.results) {
          setResults(queryWithResults.results);
        }
        setLoading(false);
      } else if (response.status === 'failed') {
        setError(response.error_message || 'Query execution failed');
        setLoading(false);
      } else {
        setLoading(false);
      }
    } catch (err) {
      console.error('Query execution error:', err);
      setError(err instanceof Error ? err.message : 'Failed to execute query');
      setLoading(false);
    }
  };

  const handleConfirmWriteQuery = async () => {
    if (!pendingWriteQuery || !dataSourceId) return;

    setLoading(true);
    setWritePreview(null);
    setError(null);
    setPermissionError(null);
    setResults(null);
    setQueryId(null);

    try {
      const response = await apiClient.executeQuery({
        data_source_id: dataSourceId,
        query_text: pendingWriteQuery,
      });

      const qid = (response as any).query_id || response.id;

      if (!qid) {
        throw new Error('Server did not return a query ID.');
      }

      setQueryId(qid);

      if (response.status === 'running' || response.status === 'pending') {
        pollForResult(qid);
      } else {
        setLoading(false);
      }
    } catch (err: any) {
      if (err.response?.data?.code === 'PERMISSION_DENIED_WRITE') {
        const data = err.response.data;
        setPermissionError({ message: data.error, hint: data.hint, dataSource: data.data_source });
        setError(null);
      } else {
        setError(err.response?.data?.error || err.message || 'Failed to execute query');
        setPermissionError(null);
      }
      setLoading(false);
    }
  };

  return (
    <>
      <AnimatePresence>
        {writePreview && (
          <WritePreviewModal
            preview={writePreview}
            queryText={pendingWriteQuery}
            onConfirm={handleConfirmWriteQuery}
            onCancel={() => { setWritePreview(null); setPendingWriteQuery(''); }}
            loading={loading}
          />
        )}
      </AnimatePresence>
      <div className={`flex h-full overflow-hidden bg-gray-50 dark:bg-gray-900/50 p-2 gap-0 ${isSidebarResizing || isEditorResizing ? 'select-none' : ''} ${isSidebarResizing ? 'cursor-col-resize' : ''} ${isEditorResizing ? 'cursor-row-resize' : ''}`}>
        {/* Data Source & Schema Sidebar */}
        <motion.div 
          className="flex-shrink-0 glass rounded-3xl sleek-shadow border border-white/20 dark:border-white/5 flex flex-col overflow-hidden"
          style={{ width: `${sidebarWidth}px` }}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
        >
          <div className="p-3 flex flex-col flex-1 overflow-hidden">
            <DataSourceSchemaSelector
              value={dataSourceId}
              onChange={setDataSourceId}
              onTableSelect={handleTableSelect}
              disabled={loading}
              onWritePermissionChange={setCanWrite}
            />
          </div>
        </motion.div>

        {/* Resizer Handle */}
        <div
          ref={resizerRef}
          onMouseDown={startSidebarResizing}
          className={`w-2 h-full flex items-center justify-center cursor-col-resize group flex-shrink-0 z-10`}
          title="Drag to resize workspace"
        >
          <div className={`w-0.5 h-12 rounded-full bg-slate-200 dark:bg-slate-700 group-hover:bg-blue-400 group-hover:h-24 transition-all duration-300 ${isSidebarResizing ? 'bg-blue-500 h-full' : ''}`}></div>
        </div>

        {/* Main Content */}
        <motion.div 
          className="flex-1 flex flex-col overflow-hidden pr-1" 
          ref={containerRef}
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, delay: 0.1, ease: 'easeOut' }}
        >

          {/* Content Area */}
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex-1 flex flex-col w-full h-full gap-3 overflow-hidden">
              {/* Show query editor only after data source is selected */}
              <AnimatePresence mode="wait">
                {!dataSourceId ? (
                  <motion.div 
                    key="placeholder"
                    className="flex flex-col items-center justify-center flex-1 border-2 border-dashed border-slate-300 dark:border-slate-700/50 rounded-3xl bg-white/50 dark:bg-slate-800/20 glass sleek-shadow m-2"
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.4, ease: 'easeOut' }}
                  >
                    <div className="text-center">
                      <motion.span 
                        className="inline-flex items-center justify-center w-20 h-20 rounded-3xl bg-blue-500/10 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 mb-6 shadow-inner"
                        animate={{ y: [0, -5, 0] }}
                        transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
                      >
                        <svg
                          className="h-10 w-10"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={1.5}
                            d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"
                          />
                        </svg>
                      </motion.span>
                      <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-2 tracking-tight">
                        Select a Data Source
                      </h3>
                      <p className="text-sm font-medium text-slate-500 dark:text-slate-400 max-w-sm mx-auto leading-relaxed">
                        Choose a database from the sidebar to start writing queries and exploring your telemetry data.
                      </p>
                    </div>
                  </motion.div>
                ) : (
                  <motion.div 
                    key="editor"
                    className="flex flex-col flex-1 overflow-hidden gap-1.5"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.3 }}
                  >

                    {/* SQL Editor - Glassy container */}
                    <motion.div 
                      className="glass rounded-3xl sleek-shadow flex flex-col flex-shrink-0 overflow-hidden"
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.4, delay: 0.1 }}
                    >
                    <div className="flex items-center justify-between px-4 py-2 bg-white/40 dark:bg-slate-800/40 border-b border-slate-200 dark:border-white/10 backdrop-blur-sm">
                      <div className="flex items-center gap-3">
                        <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest shrink-0">
                          Query Editor
                        </label>
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-lg text-[9px] font-bold uppercase tracking-widest shrink-0 ${
                            canWrite
                              ? 'bg-blue-500/10 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300'
                              : 'bg-slate-500/10 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
                          }`}
                        >
                          {canWrite ? 'Read + Write' : 'Read Only'}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="flex items-center gap-1.5 bg-white/80 dark:bg-slate-800/80 rounded-xl px-3 py-1 border border-slate-200 dark:border-white/10 hover:border-blue-400 dark:hover:border-blue-500 transition-colors sleek-shadow-sm">
                          <span className="text-[10px] text-slate-500 dark:text-slate-400 font-bold uppercase tracking-wider">Limit</span>
                          <select
                            value={rowLimit}
                            onChange={(e) => setRowLimit(Number(e.target.value))}
                            disabled={loading}
                            className="bg-transparent text-xs text-slate-900 dark:text-slate-100 focus:outline-none border-none p-0 pr-4 cursor-pointer font-semibold appearance-none w-14 text-right"
                            style={{ backgroundImage: 'none' }}
                          >
                            <option value={0}>None</option>
                            <option value={100}>100</option>
                            <option value={500}>500</option>
                            <option value={1000}>1000</option>
                            <option value={5000}>5000</option>
                          </select>
                        </div>
                        
                        <div className="h-4 w-px bg-slate-300 dark:bg-slate-600 mx-1"></div>
                        <div className="flex items-center gap-2">
                          <Button
                            onClick={handleSaveQuery}
                            disabled={!queryText.trim()}
                            variant="secondary"
                            size="sm"
                            className="rounded-xl font-bold hover:-translate-y-0.5 transition-all ease-spring h-8"
                          >
                            Save
                          </Button>
                          <Button
                            onClick={handleExecuteQuery}
                            disabled={!queryText.trim() || (isWriteQuery && !canWrite)}
                            loading={loading || previewLoading}
                            variant="primary"
                            size="sm"
                            title={isWriteQuery && !canWrite ? "Write permission required" : ""}
                            className="rounded-xl font-bold shadow-lg shadow-blue-500/20 hover:scale-105 transition-all ease-spring h-8"
                          >
                            {loading || previewLoading ? 'Running...' : 'Execute'}
                          </Button>
                        </div>
                      </div>
                    </div>
                    <div className="bg-transparent flex-shrink-0">
                      <SQLEditor
                        value={queryText}
                        onChange={setQueryText}
                        placeholder="SELECT * FROM users LIMIT 10;"
                        readOnly={loading}
                        height={`${editorHeight}px`}
                        dataSourceId={dataSourceId}
                        canWrite={canWrite}
                        onWriteDetected={setIsWriteQuery}
                        onExecute={handleExecuteQuery}
                      />
                    </div>
                  </motion.div>

                  {/* Vertical Resize Handle */}
                  <div 
                    className={`h-1.5 w-full flex items-center justify-center cursor-row-resize group/handle -my-0.5 z-10 transition-colors ${isEditorResizing ? 'bg-blue-500/20' : 'bg-transparent hover:bg-blue-500/10'}`}
                    onMouseDown={startEditorResizing}
                  >
                    <div className={`h-0.5 w-12 rounded-full transition-all duration-300 ${isEditorResizing ? 'bg-blue-500 w-24' : 'bg-slate-300 dark:bg-white/10 group-hover/handle:bg-blue-400 group-hover/handle:w-16'}`} />
                  </div>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* Error Display */}
              <AnimatePresence>
                {error && (
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    transition={{ duration: 0.3 }}
                  >
                    <QueryError
                      error={error}
                      onRetry={() => {
                        setError(null);
                        handleExecuteQuery();
                      }}
                    />
                  </motion.div>
                )}
              </AnimatePresence>

              {/* Permission Error Display */}
              <AnimatePresence>
                {permissionError && (
                  <motion.div 
                    className="glass rounded-3xl sleek-shadow p-6 flex flex-col items-center justify-center text-center gap-4 m-2 border border-red-200 dark:border-red-900/30"
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.4 }}
                  >
                    <span className="p-4 bg-red-100 dark:bg-red-500/10 text-red-600 dark:text-red-400 rounded-2xl shadow-inner">
                      <svg className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                      </svg>
                    </span>
                    <div>
                      <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">
                          Write Access Required
                      </h3>
                      <p className="text-sm font-medium text-slate-500 dark:text-slate-400 max-w-md mx-auto leading-relaxed">
                          Your groups don&apos;t have write permission on <strong className="text-slate-700 dark:text-slate-300">{permissionError.dataSource}</strong>. {permissionError.hint}
                      </p>
                    </div>
                    <div className="mt-2 flex gap-3">
                      <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} transition={springConfig.micro}>
                        <Button
                          onClick={() => router.push('/profile')}
                          variant="primary"
                          className="rounded-xl font-bold"
                        >
                          View My Groups
                        </Button>
                      </motion.div>
                      <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} transition={springConfig.micro}>
                        <Button
                          onClick={() => setPermissionError(null)}
                          variant="outline"
                          className="rounded-xl font-bold hover:bg-slate-100 dark:hover:bg-white/5 transition-colors"
                        >
                          Dismiss
                        </Button>
                      </motion.div>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* Results Container */}
              <AnimatePresence>
                {results && queryId && (
                  <motion.div 
                    className={`flex-1 flex flex-col overflow-hidden glass rounded-3xl sleek-shadow ${isFullscreenResults ? 'fullscreen-results' : ''}`}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    transition={{ duration: 0.5, delay: 0.15 }}
                  >
                    <div className="flex items-center justify-between px-4 py-2 bg-white/40 dark:bg-slate-800/40 border-b border-slate-200 dark:border-white/10 flex-shrink-0 backdrop-blur-sm">
                    <div className="flex items-center gap-3">
                      <span className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-widest">Results</span>
                      <span className="px-2 py-0.5 bg-blue-500/10 dark:bg-blue-500/20 text-blue-700 dark:text-blue-300 rounded-lg text-[10px] font-bold tracking-widest uppercase">
                        {results.row_count} rows
                      </span>
                    </div>
                    <div className="flex gap-2">
                      <motion.div whileHover={{ scale: 1.02, y: -2 }} whileTap={{ scale: 0.98 }} transition={springConfig.micro}>
                        <Button
                          onClick={handleExportCSV}
                          variant="outline"
                          size="sm"
                          className="rounded-xl font-bold text-[10px] uppercase tracking-wider py-1"
                        >
                          CSV
                        </Button>
                      </motion.div>
                      <motion.div whileHover={{ scale: 1.02, y: -2 }} whileTap={{ scale: 0.98 }} transition={springConfig.micro}>
                        <Button
                          onClick={handleExportJSON}
                          variant="outline"
                          size="sm"
                          className="rounded-xl font-bold text-[10px] uppercase tracking-wider py-1"
                        >
                          JSON
                        </Button>
                      </motion.div>
                    </div>
                  </div>
                  <div className="flex-1 overflow-hidden flex flex-col bg-transparent">
                    <QueryResults
                      queryId={queryId}
                      results={results}
                      loading={loading}
                      error={error}
                      isFullscreen={isFullscreenResults}
                      onToggleFullscreen={() => setIsFullscreenResults(!isFullscreenResults)}
                    />
                  </div>
                </motion.div>
                )}
              </AnimatePresence>

              {/* Full-Screen Overlay Implementation using CSS for the container above */}
              <style jsx global>{`
                .fullscreen-results {
                  position: fixed !important;
                  top: 1rem !important;
                  left: 1rem !important;
                  right: 1rem !important;
                  bottom: 1rem !important;
                  z-index: 9999 !important;
                  margin: 0 !important;
                  background: rgba(255, 255, 255, 0.95);
                }
                .dark .fullscreen-results {
                  background: rgba(15, 23, 42, 0.95);
                }
              `}</style>

              {/* Loading State */}
              <AnimatePresence>
                {loading && !results && (
                  <motion.div 
                    className="flex items-center justify-center flex-1 glass rounded-3xl mx-2"
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    transition={{ duration: 0.3 }}
                  >
                    <Loading variant="bars" size="lg" text="Executing query..." />
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        </motion.div>
      </div>
    </>
  );
}

