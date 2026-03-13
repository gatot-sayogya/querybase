'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown, ChevronUp, AlertTriangle, CheckCircle, Database } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { StatementPreview } from '@/lib/api/multi-query';

interface MultiQueryPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  statements: StatementPreview[];
  totalEstimatedRows: number;
  onApprove: () => void;
  onReject: () => void;
  loading?: boolean;
}

const operationTypeColors: Record<string, string> = {
  SELECT: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  INSERT: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  UPDATE: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  DELETE: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  CREATE_TABLE: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300',
  DROP_TABLE: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
  ALTER_TABLE: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
};

export function MultiQueryPreviewModal({
  isOpen,
  onClose,
  statements,
  totalEstimatedRows,
  onApprove,
  onReject,
  loading = false
}: MultiQueryPreviewModalProps) {
  const [expandedStatements, setExpandedStatements] = useState<Set<number>>(new Set());

  const toggleExpand = (sequence: number) => {
    const newExpanded = new Set(expandedStatements);
    if (newExpanded.has(sequence)) {
      newExpanded.delete(sequence);
    } else {
      newExpanded.add(sequence);
    }
    setExpandedStatements(newExpanded);
  };

  const writeOperations = statements.filter(
    s => ['INSERT', 'UPDATE', 'DELETE'].includes(s.operation_type)
  );

  const hasErrors = statements.some(s => s.error);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Database className="w-5 h-5" />
            Multi-Query Preview
          </DialogTitle>
        </DialogHeader>

        {/* Summary Section */}
        <div className="bg-muted/50 rounded-lg p-4 mb-4">
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold">{statements.length}</div>
              <div className="text-sm text-muted-foreground">Statements</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{totalEstimatedRows.toLocaleString()}</div>
              <div className="text-sm text-muted-foreground">Est. Affected Rows</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{writeOperations.length}</div>
              <div className="text-sm text-muted-foreground">Write Operations</div>
            </div>
          </div>

          {hasErrors && (
            <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-600 dark:text-red-400" />
              <span className="text-sm text-red-800 dark:text-red-300">
                Some statements have errors. Please review before proceeding.
              </span>
            </div>
          )}
        </div>

        {/* Statements List */}
        <ScrollArea className="flex-1 min-h-[300px]">
          <div className="space-y-3 pr-4">
            {statements.map((stmt) => {
              const isExpanded = expandedStatements.has(stmt.sequence);
              const hasError = !!stmt.error;

              return (
                <motion.div
                  key={stmt.sequence}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: stmt.sequence * 0.05 }}
                  className={cn(
                    'border rounded-lg overflow-hidden',
                    hasError 
                      ? 'border-red-200 dark:border-red-800 bg-red-50/50 dark:bg-red-900/10' 
                      : 'border-border'
                  )}
                >
                  {/* Header */}
                  <button
                    onClick={() => toggleExpand(stmt.sequence)}
                    className="w-full p-4 flex items-center justify-between hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-mono text-muted-foreground">
                        #{stmt.sequence + 1}
                      </span>
                      <Badge 
                        variant="secondary"
                        className={cn(
                          'font-medium',
                          operationTypeColors[stmt.operation_type] || 'bg-gray-100'
                        )}
                      >
                        {stmt.operation_type}
                      </Badge>
                      <span className="text-sm text-muted-foreground truncate max-w-md">
                        {stmt.query_text.substring(0, 80)}
                        {stmt.query_text.length > 80 ? '...' : ''}
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      {stmt.estimated_rows > 0 && (
                        <span className="text-sm text-muted-foreground">
                          ~{stmt.estimated_rows.toLocaleString()} rows
                        </span>
                      )}
                      {hasError && (
                        <AlertTriangle className="w-4 h-4 text-red-600 dark:text-red-400" />
                      )}
                      {isExpanded ? (
                        <ChevronUp className="w-4 h-4 text-muted-foreground" />
                      ) : (
                        <ChevronDown className="w-4 h-4 text-muted-foreground" />
                      )}
                    </div>
                  </button>

                  {/* Expanded Content */}
                  <AnimatePresence>
                    {isExpanded && (
                      <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        transition={{ duration: 0.2 }}
                        className="border-t border-border"
                      >
                        <div className="p-4 space-y-4">
                          {/* SQL Preview */}
                          <div>
                            <div className="text-sm font-medium mb-2">SQL</div>
                            <pre className="bg-muted p-3 rounded-md text-sm font-mono overflow-x-auto">
                              {stmt.query_text}
                            </pre>
                          </div>

                          {/* Error Message */}
                          {hasError && (
                            <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
                              <div className="text-sm font-medium text-red-800 dark:text-red-300 mb-1">
                                Error
                              </div>
                              <div className="text-sm text-red-700 dark:text-red-400">
                                {stmt.error}
                              </div>
                            </div>
                          )}

                          {/* Preview Data */}
                          {stmt.preview_rows && stmt.preview_rows.length > 0 && (
                            <div>
                              <div className="text-sm font-medium mb-2">
                                Preview ({stmt.preview_rows.length} rows)
                              </div>
                              <div className="overflow-x-auto">
                                <table className="w-full text-sm">
                                  <thead>
                                    <tr className="border-b">
                                      {stmt.columns?.map((col) => (
                                        <th key={col.name} className="text-left p-2 font-medium">
                                          {col.name}
                                        </th>
                                      ))}
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {stmt.preview_rows.slice(0, 5).map((row, idx) => (
                                      <tr key={idx} className="border-b last:border-0">
                                        {stmt.columns?.map((col) => (
                                          <td key={col.name} className="p-2 text-muted-foreground">
                                            {row[col.name] === null ? (
                                              <span className="italic">null</span>
                                            ) : (
                                              String(row[col.name]).substring(0, 50)
                                            )}
                                          </td>
                                        ))}
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                                {stmt.preview_rows.length > 5 && (
                                  <div className="text-center text-sm text-muted-foreground py-2">
                                    +{stmt.preview_rows.length - 5} more rows
                                  </div>
                                )}
                              </div>
                            </div>
                          )}

                          {/* Stats */}
                          <div className="flex gap-4 text-sm text-muted-foreground">
                            <span>Estimated rows: {stmt.estimated_rows.toLocaleString()}</span>
                          </div>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </motion.div>
              );
            })}
          </div>
        </ScrollArea>

        {/* Footer */}
        <DialogFooter className="flex justify-between items-center gap-4">
          <div className="text-sm text-muted-foreground">
            All statements will execute atomically
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onReject} disabled={loading}>
              Cancel
            </Button>
            <Button 
              onClick={onApprove} 
              disabled={loading || hasErrors}
              className="gap-2"
            >
              {loading ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <CheckCircle className="w-4 h-4" />
                  Execute All
                </>
              )}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
