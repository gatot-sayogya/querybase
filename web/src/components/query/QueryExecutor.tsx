'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import toast from 'react-hot-toast';
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
      {writePreview && (
        <WritePreviewModal
          preview={writePreview}
          queryText={pendingWriteQuery}
          onConfirm={handleConfirmWriteQuery}
          onCancel={() => { setWritePreview(null); setPendingWriteQuery(''); }}
          loading={loading}
        />
      )}
      <div className="flex h-full overflow-hidden">
        {/* Data Source & Schema Sidebar */}
        <div className="w-64 flex-shrink-0 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 flex flex-col overflow-hidden">
          <div className="p-2 flex flex-col flex-1 overflow-hidden">
            <DataSourceSchemaSelector
              value={dataSourceId}
              onChange={setDataSourceId}
              onTableSelect={handleTableSelect}
              disabled={loading}
              onWritePermissionChange={setCanWrite}
            />
          </div>
        </div>

        {/* Main Content */}
        <div className="flex-1 flex flex-col overflow-hidden">

          {/* Content Area */}
          <div className="flex-1 flex flex-col bg-gray-50 dark:bg-gray-900 overflow-hidden">
            <div className="flex-1 flex flex-col w-full h-full p-2 gap-2 overflow-hidden">
              {/* Show query editor only after data source is selected */}
              {!dataSourceId ? (
                <div className="flex items-center justify-center h-96 border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-xl bg-white dark:bg-gray-800">
                  <div className="text-center animate-fade-in">
                    <span className="inline-block p-4 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-500 mb-4">
                      <svg
                        className="h-12 w-12"
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
                    </span>
                    <h3 className="mt-2 text-lg font-medium text-gray-900 dark:text-white">
                      Select a Data Source
                    </h3>
                    <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 max-w-xs mx-auto">
                      Choose a database from the sidebar to start writing queries and exploring your data
                    </p>
                  </div>
                </div>
              ) : (
                <div className="animate-slide-up flex flex-col flex-1 overflow-hidden gap-3">


                  {/* SQL Editor - Now with flexible height */}
                  <div className="space-y-1">
                    <div className="flex items-center justify-between px-2 py-1 bg-gray-50 dark:bg-gray-800/50">
                      <label className="text-xs text-gray-500 dark:text-gray-400">
                        Query
                      </label>
                      <div className="flex items-center gap-2">
                        <div className="flex items-center gap-1 bg-white dark:bg-gray-700 rounded px-2 py-0.5 border border-gray-200 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 transition-colors">
                          <span className="text-[10px] text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wider">Limit</span>
                          <select
                            value={rowLimit}
                            onChange={(e) => setRowLimit(Number(e.target.value))}
                            disabled={loading}
                            className="bg-transparent text-xs text-gray-900 dark:text-gray-100 focus:outline-none border-none p-0 pr-4 cursor-pointer font-medium appearance-none w-16 text-right"
                            style={{ backgroundImage: 'none' }}
                          >
                            <option value={0}>None</option>
                            <option value={100}>100</option>
                            <option value={500}>500</option>
                            <option value={1000}>1000</option>
                            <option value={5000}>5000</option>
                          </select>
                        </div>
                        
                        <div className="h-4 w-px bg-gray-300 dark:bg-gray-600 mx-1"></div>
                        <Button
                          onClick={handleSaveQuery}
                          disabled={!queryText.trim()}
                          variant="secondary"
                          size="sm"
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
                        >
                          {loading || previewLoading ? 'Running...' : 'Run'}
                        </Button>
                      </div>
                    </div>
                    <div className="border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex-shrink-0">
                      <SQLEditor
                        value={queryText}
                        onChange={setQueryText}
                        placeholder="SELECT * FROM users LIMIT 10;"
                        readOnly={loading}
                        height="250px"
                        dataSourceId={dataSourceId}
                        canWrite={canWrite}
                        onWriteDetected={setIsWriteQuery}
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Error Display */}
              {error && (
                <div className="animate-shake">
                  <QueryError
                    error={error}
                    onRetry={() => {
                      setError(null);
                      handleExecuteQuery();
                    }}
                  />
                </div>
              )}

              {/* Permission Error Display */}
              {permissionError && (
                <div className="animate-slide-up bg-white dark:bg-gray-800 border border-red-200 dark:border-red-900/50 rounded-lg shadow-sm p-4 flex flex-col items-center justify-center text-center gap-2 m-4">
                  <span className="p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded-full mb-2">
                    <svg className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                  </span>
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                    Write Access Required
                  </h3>
                  <p className="text-sm text-gray-500 dark:text-gray-400 max-w-md">
                    Your groups don&apos;t have write permission on <strong>{permissionError.dataSource}</strong>. {permissionError.hint}
                  </p>
                  <div className="mt-4 flex gap-3">
                    <Button
                      onClick={() => router.push('/profile')}
                      variant="primary"
                    >
                      View My Groups
                    </Button>
                    <Button
                      onClick={() => setPermissionError(null)}
                      variant="outline"
                    >
                      Dismiss
                    </Button>
                  </div>
                </div>
              )}

              {/* Results - Now with flexible height */}
              {results && queryId && (
                <div className="flex-1 flex flex-col animate-slide-up overflow-hidden bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-sm">
                  <div className="flex items-center justify-between px-3 py-2 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-gray-900 dark:text-white">Results</span>
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {results.row_count} rows
                      </span>
                    </div>
                    <div className="flex gap-1">
                      <Button
                        onClick={handleExportCSV}
                        variant="outline"
                        size="sm"
                      >
                        CSV
                      </Button>
                      <Button
                        onClick={handleExportJSON}
                        variant="outline"
                        size="sm"
                      >
                        JSON
                      </Button>
                    </div>
                  </div>
                  <div className="flex-1 overflow-hidden bg-white dark:bg-gray-800 flex flex-col">
                    <QueryResults
                      queryId={queryId}
                      results={results}
                      loading={loading}
                      error={error}
                    />
                  </div>
                </div>
              )}

              {/* Loading State */}
              {loading && !results && (
                <div className="flex items-center justify-center h-64 animate-fade-in">
                  <Loading variant="bars" size="lg" text="Executing query..." />
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

