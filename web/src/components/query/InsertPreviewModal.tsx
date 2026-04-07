import { motion, AnimatePresence } from 'framer-motion';
import Button from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import type { InsertPreviewResult } from '@/lib/api/insert-preview';

interface InsertPreviewModalProps {
  preview: InsertPreviewResult;
  queryText: string;
  onConfirm: () => void;
  onCancel: () => void;
  loading: boolean;
}

export default function InsertPreviewModal({
  preview,
  queryText,
  onConfirm,
  onCancel,
  loading,
}: InsertPreviewModalProps) {
  const formatCellValue = (value: unknown): string => {
    if (value === null) return 'NULL';
    if (value === undefined) return '';
    if (typeof value === 'object') {
      const str = JSON.stringify(value);
      return str.length > 100 ? str.substring(0, 100) + '...' : str;
    }
    const str = String(value);
    return str.length > 100 ? str.substring(0, 100) + '...' : str;
  };

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
        onClick={onCancel}
      >
        <motion.div
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.95, opacity: 0 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Badge variant="success">INSERT</Badge>
                <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Preview Data to Insert
                </h2>
              </div>
              <button
                onClick={onCancel}
                className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              >
                <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
              INTO <span className="font-semibold">{preview.table_name}</span> — {preview.total_row_count} row{preview.total_row_count !== 1 ? 's' : ''} to insert
            </p>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-auto p-6">
            {/* SELECT indicator */}
            {preview.preview_type === 'select' && (
              <div className="mb-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3 text-sm text-blue-800 dark:text-blue-300">
                Preview from SELECT query. Showing up to 50 rows.
                {preview.total_row_count > 50 && (
                  <span> Total: {preview.total_row_count} rows.</span>
                )}
              </div>
            )}

            {/* Empty state */}
            {preview.rows.length === 0 && (
              <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-4 text-center text-yellow-800 dark:text-yellow-300">
                No rows to insert. The query will not insert any data.
              </div>
            )}

            {/* Data table */}
            {preview.rows.length > 0 && (
              <div className="border rounded-md overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-gray-100 dark:bg-gray-700">
                    <tr>
                      <th className="px-4 py-2 text-left font-medium text-gray-700 dark:text-gray-300 w-12 whitespace-nowrap">#</th>
                      {preview.columns.map((col) => (
                        <th key={col.name} className="px-4 py-2 text-left font-medium text-gray-700 dark:text-gray-300 whitespace-nowrap">
                          <div className="flex flex-col">
                            <span>{col.name}</span>
                            <span className="text-xs text-gray-500 dark:text-gray-400">{col.type}</span>
                          </div>
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {preview.rows.map((row, idx) => (
                      <tr key={idx} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                        <td className="px-4 py-2 text-gray-500 dark:text-gray-400 text-sm whitespace-nowrap">{idx + 1}</td>
                        {preview.columns.map((col) => (
                          <td key={col.name} className="px-4 py-2 font-mono text-xs text-gray-900 dark:text-gray-100 whitespace-nowrap">
                            {formatCellValue(row[col.name])}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
                
                {preview.rows.length < preview.total_row_count && (
                  <div className="px-4 py-2 text-sm text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
                    Showing {preview.rows.length} of {preview.total_row_count} rows
                  </div>
                )}
              </div>
            )}

            {/* Query text */}
            <div className="mt-6">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Query</h3>
              <pre className="bg-gray-100 dark:bg-gray-900 p-3 rounded text-sm text-gray-800 dark:text-gray-200 overflow-x-auto">
                {queryText}
              </pre>
            </div>
          </div>

          {/* Footer */}
          <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-between items-center">
            <Button
              variant="outline"
              onClick={onCancel}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button
              onClick={onConfirm}
              disabled={loading}
              className="bg-green-600 hover:bg-green-700 text-white"
            >
              {loading ? (
                <div className="flex items-center gap-2">
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                  Submitting...
                </div>
              ) : (
                'Submit for Approval'
              )}
            </Button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
