'use client';

import { motion, AnimatePresence } from 'framer-motion';
import Button from '@/components/ui/Button';

interface ValidationResult {
  valid: boolean;
  status: string;
  message: string;
  affected_rows: number;
  preview_rows?: Record<string, unknown>[];
  columns?: string[];
  suggestion?: string;
}

interface QueryValidationModalProps {
  isOpen: boolean;
  validation: ValidationResult | null;
  queryText?: string;
  onEdit: () => void;
  onCancel: () => void;
}

export default function QueryValidationModal({
  isOpen,
  validation,
  queryText,
  onEdit,
  onCancel,
}: QueryValidationModalProps) {
  if (!isOpen || !validation) return null;

  const isNoMatch = validation.status === 'no_match';

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
          onClick={onCancel}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 20 }}
            transition={{ type: 'spring', damping: 25, stiffness: 300 }}
            className="glass rounded-2xl sleek-shadow max-w-lg w-full overflow-hidden border border-white/20 dark:border-white/10"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="px-6 py-4 border-b border-slate-200/50 dark:border-white/10">
              <div className="flex items-center gap-3">
                <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                  isNoMatch 
                    ? 'bg-blue-100 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400' 
                    : 'bg-green-100 dark:bg-green-500/20 text-green-600 dark:text-green-400'
                }`}>
                  {isNoMatch ? (
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  ) : (
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  )}
                </div>
                <div>
                  <h3 className="text-lg font-bold text-slate-900 dark:text-white">
                    {isNoMatch ? 'Query Would Affect 0 Rows' : 'Query Validation'}
                  </h3>
                  <p className="text-sm text-slate-500 dark:text-slate-400">
                    {isNoMatch ? 'No matching data found' : 'Validation successful'}
                  </p>
                </div>
              </div>
            </div>

            {/* Content */}
            <div className="px-6 py-5">
              {/* Message */}
              <div className={`p-4 rounded-xl mb-4 ${
                isNoMatch 
                  ? 'bg-blue-50/50 dark:bg-blue-500/10 border border-blue-200 dark:border-blue-500/20' 
                  : 'bg-green-50/50 dark:bg-green-500/10 border border-green-200 dark:border-green-500/20'
              }`}>
                <p className={`text-sm font-medium ${
                  isNoMatch 
                    ? 'text-blue-800 dark:text-blue-300' 
                    : 'text-green-800 dark:text-green-300'
                }`}>
                  {validation.message}
                </p>
              </div>

              {/* Query Preview */}
              {queryText && (
                <div className="mb-4">
                  <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2 block">
                    Your Query
                  </label>
                  <div className="bg-slate-100 dark:bg-slate-900 rounded-lg p-3 overflow-x-auto">
                    <code className="text-xs font-mono text-slate-700 dark:text-slate-300 whitespace-pre-wrap break-all">
                      {queryText.length > 200 ? queryText.substring(0, 200) + '...' : queryText}
                    </code>
                  </div>
                </div>
              )}

              {/* Suggestion */}
              {validation.suggestion && (
                <div className="flex items-start gap-2 text-sm text-slate-600 dark:text-slate-400">
                  <svg className="w-4 h-4 mt-0.5 flex-shrink-0 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                  </svg>
                  <p>{validation.suggestion}</p>
                </div>
              )}

              {/* Preview rows if available and not no_match */}
              {!isNoMatch && validation.preview_rows && validation.preview_rows.length > 0 && (
                <div className="mt-4">
                  <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2 block">
                    Preview ({validation.preview_rows.length} of {validation.affected_rows} rows)
                  </label>
                  <div className="bg-slate-100 dark:bg-slate-900 rounded-lg p-3 overflow-x-auto max-h-40">
                    <table className="min-w-full text-xs">
                      <thead>
                        <tr>
                          {validation.columns?.map((col) => (
                            <th key={col} className="text-left font-semibold text-slate-600 dark:text-slate-400 pb-2 pr-4">
                              {col}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {validation.preview_rows.map((row, idx) => (
                          <tr key={idx} className="border-t border-slate-200 dark:border-white/5">
                            {validation.columns?.map((col) => (
                              <td key={col} className="py-2 pr-4 text-slate-700 dark:text-slate-300">
                                {String(row[col] ?? 'NULL')}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>

            {/* Footer */}
            <div className="px-6 py-4 border-t border-slate-200/50 dark:border-white/10 flex justify-end gap-3">
              <Button variant="secondary" onClick={onCancel}>
                Cancel
              </Button>
              <Button onClick={onEdit}>
                {isNoMatch ? 'Edit Query' : 'Continue'}
              </Button>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
