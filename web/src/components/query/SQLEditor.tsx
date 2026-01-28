'use client';

import Editor, { Monaco } from '@monaco-editor/react';
import { useState } from 'react';
import { useSchemaStore } from '@/stores/schema-store';
import type { editor } from 'monaco-editor';

interface SQLEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  readOnly?: boolean;
  height?: string;
  dataSourceId?: string;
}

export default function SQLEditor({
  value,
  onChange,
  placeholder = 'SELECT * FROM users LIMIT 10;',
  readOnly = false,
  height = '400px',
  dataSourceId,
}: SQLEditorProps) {
  const [editorHeight] = useState(height);
  const { getTableNames, getAllColumns } = useSchemaStore();

  const handleEditorWillMount = (monaco: Monaco) => {
    // Register custom SQL language features if needed
    monaco.languages.registerCompletionItemProvider('sql', {
      provideCompletionItems: async (
        model: editor.ITextModel,
        position: editor.Position
      ) => {
        const suggestions: monaco.languages.CompletionItem[] = [];

        // Get text before cursor for context
        const textUntilPosition = model.getValueInRange({
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: position.lineNumber,
          endColumn: position.column,
        });

        const words = textUntilPosition.split(/\s+/);
        const lastWord = words[words.length - 1].toUpperCase();

        // SQL Keywords
        const keywords = [
          'SELECT', 'FROM', 'WHERE', 'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET',
          'DELETE', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'ON', 'AND', 'OR',
          'NOT', 'IN', 'LIKE', 'BETWEEN', 'ORDER', 'BY', 'GROUP', 'HAVING',
          'LIMIT', 'OFFSET', 'ASC', 'DESC', 'DISTINCT', 'COUNT', 'SUM', 'AVG',
          'MAX', 'MIN', 'CREATE', 'TABLE', 'DROP', 'ALTER', 'ADD', 'COLUMN',
          'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'UNIQUE', 'INDEX', 'VIEW',
        ];

        // Always suggest SQL keywords
        keywords.forEach((keyword) => {
          if (keyword.startsWith(lastWord)) {
            suggestions.push({
              label: keyword,
              kind: monaco.languages.CompletionItemKind.Keyword,
              insertText: keyword,
              detail: 'SQL Keyword',
              sortText: `0_${keyword}`,
            });
          }
        });

        // Add schema-based suggestions if data source is selected
        if (dataSourceId) {
          const tableNames = getTableNames(dataSourceId);
          const allColumns = getAllColumns(dataSourceId);

          // Detect if we're after FROM, JOIN, or INTO (suggest tables)
          const previousWords = words.slice(Math.max(0, words.length - 5)).join(' ');
          const suggestTables =
            /FROM\s*$/i.test(previousWords) ||
            /JOIN\s+(\w+\s+)?$/i.test(previousWords) ||
            /INTO\s*$/i.test(previousWords) ||
            /UPDATE\s*$/i.test(previousWords);

          if (suggestTables) {
            // Suggest table names
            tableNames.forEach((tableName) => {
              if (tableName.toUpperCase().startsWith(lastWord)) {
                suggestions.push({
                  label: tableName,
                  kind: monaco.languages.CompletionItemKind.Class,
                  insertText: tableName,
                  detail: 'Table',
                  sortText: `1_${tableName}`,
                });
              }
            });
          } else {
            // Suggest columns and tables
            tableNames.forEach((tableName) => {
              if (tableName.toUpperCase().startsWith(lastWord)) {
                suggestions.push({
                  label: tableName,
                  kind: monaco.languages.CompletionItemKind.Class,
                  insertText: tableName,
                  detail: 'Table',
                  sortText: `1_${tableName}`,
                });
              }

              // Add columns for this table
              const columns = allColumns.get(tableName);
              if (columns) {
                columns.forEach((column) => {
                  const columnLabel = `${tableName}.${column.column_name}`;
                  if (columnLabel.toUpperCase().startsWith(lastWord) || column.column_name.toUpperCase().startsWith(lastWord)) {
                    suggestions.push({
                      label: columnLabel,
                      kind: monaco.languages.CompletionItemKind.Field,
                      insertText: columnLabel,
                      detail: `${column.data_type}${column.is_nullable ? '' : ' NOT NULL'}`,
                      documentation: `Column: ${column.column_name}\nType: ${column.data_type}\n${column.is_primary_key ? 'Primary Key\n' : ''}${column.is_foreign_key ? 'Foreign Key' : ''}`,
                      sortText: `2_${tableName}_${column.column_name}`,
                    });
                  }
                });
              }
            });
          }
        }

        return { suggestions };
      },
    });

    // Register signature help for functions
    monaco.languages.registerSignatureHelpProvider('sql', {
      signatureHelpTriggerCharacters: ['(', ','],
      provideSignatureHelp: async () => {
        return {
          value: {
            activeParameter: 0,
            activeSignature: 0,
            signatures: [
              {
                label: 'COUNT(expression)',
                parameters: [
                  {
                    label: 'expression',
                    documentation: 'The column or expression to count',
                  },
                ],
              },
              {
                label: 'SUM(expression)',
                parameters: [
                  {
                    label: 'expression',
                    documentation: 'The column or expression to sum',
                  },
                ],
              },
              {
                label: 'AVG(expression)',
                parameters: [
                  {
                    label: 'expression',
                    documentation: 'The column or expression to average',
                  },
                ],
              },
              {
                label: 'MAX(expression)',
                parameters: [
                  {
                    label: 'expression',
                    documentation: 'The column or expression to find maximum value',
                  },
                ],
              },
              {
                label: 'MIN(expression)',
                parameters: [
                  {
                    label: 'expression',
                    documentation: 'The column or expression to find minimum value',
                  },
                ],
              },
            ],
          },
          dispose: () => {},
        };
      },
    });
  };

  const handleEditorChange = (value: string | undefined) => {
    onChange(value || '');
  };

  return (
    <div className="border border-gray-300 dark:border-gray-700 rounded-lg overflow-hidden">
      <Editor
        height={editorHeight}
        defaultLanguage="sql"
        value={value}
        onChange={handleEditorChange}
        theme="vs-dark"
        beforeMount={handleEditorWillMount}
        options={{
          minimap: { enabled: false },
          fontSize: 14,
          lineNumbers: 'on',
          roundedSelection: false,
          scrollBeyondLastLine: false,
          readOnly: readOnly,
          placeholder: placeholder,
          automaticLayout: true,
          tabSize: 2,
          wordWrap: 'on',
          formatOnPaste: true,
          formatOnType: true,
          suggestOnTriggerCharacters: true,
          quickSuggestions: {
            other: true,
            comments: false,
            strings: false,
          },
          parameterHints: {
            enabled: true,
          },
        }}
      />
    </div>
  );
}
