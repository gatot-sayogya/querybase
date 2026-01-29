'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import { apiClient } from '@/lib/api-client';
import SQLEditor from './SQLEditor';
import DataSourceSchemaSelector from './DataSourceSchemaSelector';
import QueryResults from './QueryResults';
import type { QueryResult } from '@/types';

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
    setResults(null);
    setQueryId(null);

    try {
      // Add LIMIT if not present and query is SELECT
      let finalQuery = queryText.trim();
      const isSelectQuery = /^\s*SELECT\s/i.test(finalQuery);
      const hasLimit = /\bLIMIT\s+\d+\s*$/i.test(finalQuery);

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
    } catch (err) {
      console.error('Query execution error:', err);
      setError(err instanceof Error ? err.message : 'Failed to execute query');
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
      alert('Query saved successfully!');
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

  return (
    <div className="flex h-[calc(100vh-120px)] gap-4">
      {/* Data Source & Schema Sidebar */}
      <div className="w-80 flex-shrink-0 space-y-4">
        <DataSourceSchemaSelector
          value={dataSourceId}
          onChange={setDataSourceId}
          onTableSelect={handleTableSelect}
          disabled={loading}
        />
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Query Editor</h1>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Execute SQL queries on your data sources
            </p>
          </div>
          <div className="text-sm text-gray-600 dark:text-gray-400">
            Logged in as <strong>{user?.username}</strong>
          </div>
        </div>

        {/* Show query editor only after data source is selected */}
        {!dataSourceId ? (
          <div className="flex items-center justify-center h-96 border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg">
            <div className="text-center">
              <svg
                className="mx-auto h-16 w-16 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"
                />
              </svg>
              <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">
                Select a Data Source
              </h3>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                Choose a database from the selector above to start writing queries
              </p>
            </div>
          </div>
        ) : (
          <>
            {/* Row Limit Selector */}
            <div className="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Row Limit:
              </label>
              <select
                value={rowLimit}
                onChange={(e) => setRowLimit(Number(e.target.value))}
                disabled={loading}
                className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value={0}>No Limit</option>
                <option value={100}>100 rows</option>
                <option value={500}>500 rows</option>
                <option value={1000}>1000 rows (default)</option>
                <option value={5000}>5000 rows</option>
                <option value={10000}>10000 rows</option>
              </select>
              <span className="text-xs text-gray-500 dark:text-gray-400">
                Automatically added to SELECT queries without LIMIT
              </span>
            </div>

            {/* SQL Editor */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  SQL Query
                </label>
                <div className="flex space-x-2">
                  <button
                    onClick={handleSaveQuery}
                    disabled={loading || !queryText.trim()}
                    className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Save Query
                  </button>
                  <button
                    onClick={handleExecuteQuery}
                    disabled={loading || !queryText.trim()}
                    className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {loading ? 'Executing...' : 'Run Query'}
                  </button>
                </div>
              </div>
              <SQLEditor
                value={queryText}
                onChange={setQueryText}
                placeholder="SELECT * FROM users LIMIT 10;"
                readOnly={loading}
                height="400px"
                dataSourceId={dataSourceId}
              />
            </div>
          </>
        )}

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-red-400"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800 dark:text-red-400">
                Error
              </h3>
              <p className="mt-1 text-sm text-red-700 dark:text-red-300 whitespace-pre-wrap">
                {error}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Results */}
      {results && queryId && (
        <div className="space-y-4">
          <div className="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 pb-4">
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                Query Results
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {results.row_count} rows returned
              </p>
            </div>
            <div className="flex space-x-2">
              <button
                onClick={handleExportCSV}
                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                Export CSV
              </button>
              <button
                onClick={handleExportJSON}
                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                Export JSON
              </button>
            </div>
          </div>
          <QueryResults
            queryId={queryId}
            results={results}
            loading={loading}
            error={error}
          />
        </div>
      )}

      {/* Loading State */}
      {loading && !results && (
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">Executing query...</p>
          </div>
        </div>
      )}
      </div>
    </div>
  );
}
