import { create } from 'zustand';
import type { DatabaseSchema, TableInfo, SchemaColumnInfo } from '@/types';
import { apiClient } from '@/lib/api-client';

interface SchemaState {
  // State
  schemas: Map<string, DatabaseSchema>; // data_source_id -> schema
  currentDataSourceId: string | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  loadSchema: (dataSourceId: string) => Promise<DatabaseSchema>;
  loadTables: (dataSourceId: string) => Promise<TableInfo[]>;
  loadTableDetails: (dataSourceId: string, tableName: string) => Promise<TableInfo>;
  searchTables: (dataSourceId: string, searchTerm: string) => Promise<TableInfo[]>;
  setCurrentDataSource: (dataSourceId: string) => void;
  clearError: () => void;
  clearSchema: (dataSourceId: string) => void;

  // Helper methods for autocomplete
  getTableNames: (dataSourceId: string) => string[];
  getColumns: (dataSourceId: string, tableName: string) => SchemaColumnInfo[];
  getAllColumns: (dataSourceId: string) => Map<string, SchemaColumnInfo[]>;
}

export const useSchemaStore = create<SchemaState>((set, get) => ({
  schemas: new Map(),
  currentDataSourceId: null,
  isLoading: false,
  error: null,

  loadSchema: async (dataSourceId: string) => {
    set({ isLoading: true, error: null, currentDataSourceId: dataSourceId });
    try {
      const schema = await apiClient.getDatabaseSchema(dataSourceId);
      set((state) => {
        const newSchemas = new Map(state.schemas);
        newSchemas.set(dataSourceId, schema);
        return {
          schemas: newSchemas,
          isLoading: false,
        };
      });
      return schema;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load schema';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  loadTables: async (dataSourceId: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getTables(dataSourceId);
      // Update partial schema info
      set((state) => {
        const newSchemas = new Map(state.schemas);
        const existingSchema = newSchemas.get(dataSourceId) || {
          data_source_id: dataSourceId,
          data_source_name: '',
          database_type: '',
          database_name: '',
          tables: response.tables,
        };
        existingSchema.tables = response.tables;
        newSchemas.set(dataSourceId, existingSchema);
        return {
          schemas: newSchemas,
          isLoading: false,
        };
      });
      return response.tables;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load tables';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  loadTableDetails: async (dataSourceId: string, tableName: string) => {
    set({ isLoading: true, error: null });
    try {
      const tableInfo = await apiClient.getTableDetails(dataSourceId, tableName);
      // Update schema with detailed table info
      set((state) => {
        const newSchemas = new Map(state.schemas);
        const schema = newSchemas.get(dataSourceId);
        if (schema) {
          const tableIndex = schema.tables.findIndex((t) => t.table_name === tableName);
          if (tableIndex >= 0) {
            schema.tables[tableIndex] = tableInfo;
          } else {
            schema.tables.push(tableInfo);
          }
          newSchemas.set(dataSourceId, schema);
        }
        return {
          schemas: newSchemas,
          isLoading: false,
        };
      });
      return tableInfo;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load table details';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  searchTables: async (dataSourceId: string, searchTerm: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.searchTables(dataSourceId, searchTerm);
      set({ isLoading: false });
      return response.tables;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to search tables';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  setCurrentDataSource: (dataSourceId: string) => {
    set({ currentDataSourceId: dataSourceId });
  },

  clearError: () => set({ error: null }),

  clearSchema: (dataSourceId: string) => {
    set((state) => {
      const newSchemas = new Map(state.schemas);
      newSchemas.delete(dataSourceId);
      return { schemas: newSchemas };
    });
  },

  // Helper methods for autocomplete
  getTableNames: (dataSourceId: string) => {
    const schema = get().schemas.get(dataSourceId);
    if (!schema) return [];
    return schema.tables.map((t) => t.table_name);
  },

  getColumns: (dataSourceId: string, tableName: string) => {
    const schema = get().schemas.get(dataSourceId);
    if (!schema) return [];
    const table = schema.tables.find((t) => t.table_name === tableName);
    return table?.columns || [];
  },

  getAllColumns: (dataSourceId: string) => {
    const schema = get().schemas.get(dataSourceId);
    if (!schema) return new Map();
    const columnMap = new Map<string, SchemaColumnInfo[]>();
    schema.tables.forEach((table) => {
      columnMap.set(table.table_name, table.columns);
    });
    return columnMap;
  },
}));
