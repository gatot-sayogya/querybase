import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import type { InsertPreviewResult } from '@/lib/api/insert-preview';

interface InsertPreviewPanelProps {
  preview: InsertPreviewResult;
  onProceed: () => void;
  onCancel: () => void;
}

export default function InsertPreviewPanel({
  preview,
  onProceed,
  onCancel,
}: InsertPreviewPanelProps) {
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
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Badge variant="info">INSERT</Badge>
          <span className="text-lg font-semibold">INTO {preview.table_name}</span>
          <span className="text-muted-foreground">
            {preview.total_row_count} row{preview.total_row_count !== 1 ? 's' : ''} to insert
          </span>
        </div>
      </div>

      {/* SELECT indicator */}
      {preview.preview_type === 'select' && (
        <div className="bg-blue-50 border border-blue-200 rounded-md p-3 text-sm text-blue-800">
          Preview from SELECT query. Showing up to 50 rows.
          {preview.total_row_count > 50 && (
            <span> Total: {preview.total_row_count} rows.</span>
          )}
        </div>
      )}

      {/* Empty state */}
      {preview.rows.length === 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-md p-3 text-sm text-yellow-800">
          No rows to insert. The query will not insert any data.
        </div>
      )}

      {/* Data table */}
      {preview.rows.length > 0 && (
        <div className="border rounded-md overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-2 text-left font-medium w-12">#</th>
                {preview.columns.map((col) => (
                  <th key={col.name} className="px-4 py-2 text-left font-medium">
                    <div className="flex flex-col">
                      <span>{col.name}</span>
                      <span className="text-xs text-muted-foreground">{col.type}</span>
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {preview.rows.map((row, idx) => (
                <tr key={idx} className="border-t hover:bg-muted/50">
                  <td className="px-4 py-2 text-muted-foreground">{idx + 1}</td>
                  {preview.columns.map((col) => (
                    <td key={col.name} className="px-4 py-2 font-mono text-xs">
                      {formatCellValue(row[col.name])}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          
          {preview.rows.length < preview.total_row_count && (
            <div className="px-4 py-2 text-sm text-muted-foreground border-t bg-muted/30">
              Showing {preview.rows.length} of {preview.total_row_count} rows
            </div>
          )}
        </div>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={onProceed}>
          Proceed to Transaction
        </Button>
      </div>
    </div>
  );
}
