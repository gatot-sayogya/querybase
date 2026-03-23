export interface ColumnInfo {
  name: string;
  type: string;
}

export interface InsertPreviewResult {
  table_name: string;
  columns: ColumnInfo[];
  rows: Record<string, unknown>[];
  total_row_count: number;
  preview_type: 'values' | 'select';
  select_query?: string;
}

export interface InsertPreviewRequest {
  data_source_id: string;
  query_text: string;
}
