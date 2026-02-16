'use client';

import Editor, { Monaco } from '@monaco-editor/react';
import { useEffect, useRef, useState } from 'react';
import { useSchemaStore } from '@/stores/schema-store';
import { useThemeStore } from '@/stores/theme-store';

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
  const [monaco, setMonaco] = useState<Monaco | null>(null);
  const [editor, setEditor] = useState<any>(null);
  const completionProviderRef = useRef<any>(null);
  const { getTableNames, getAllColumns, schemas } = useSchemaStore();
  const { getEffectiveTheme } = useThemeStore();
  const effectiveTheme = getEffectiveTheme();

  const createCompletionProvider = (monacoInstance: any, currentDataSourceId: string) => {
    const languages = monacoInstance.languages;
    const CompletionItemKind = languages.CompletionItemKind;

    return languages.registerCompletionItemProvider('sql', {
      provideCompletionItems: async (model: any, position: any) => {
        const suggestions: any[] = [];

        // Get text before cursor for context
        const textUntilPosition = model.getValueInRange({
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: position.lineNumber,
          endColumn: position.column,
        });

        const words = textUntilPosition.split(/\s+/);
        let lastWord = words[words.length - 1] || '';

        // Remove trailing special characters (semicolon, comma, parentheses, etc.)
        lastWord = lastWord.replace(/[;,\(\)\[\]\{\}]*$/, '').toUpperCase();

        // Handle empty last word (suggest all keywords)
        const shouldShowAll = !lastWord || lastWord === '';

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
          if (shouldShowAll || keyword.startsWith(lastWord)) {
            suggestions.push({
              label: keyword,
              kind: CompletionItemKind.Keyword,
              insertText: keyword,
              detail: 'SQL Keyword',
              sortText: `0_${keyword}`,
            });
          }
        });

        // Add schema-based suggestions if data source is selected and schema is loaded
        if (currentDataSourceId && schemas.has(currentDataSourceId)) {
          const tableNames = getTableNames(currentDataSourceId);
          const allColumns = getAllColumns(currentDataSourceId);

          console.log('Autocomplete:', {
            dataSourceId: currentDataSourceId,
            tableNames,
            columnCount: allColumns.size,
          });

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
              if (shouldShowAll || tableName.toUpperCase().startsWith(lastWord)) {
                suggestions.push({
                  label: tableName,
                  kind: CompletionItemKind.Class,
                  insertText: tableName,
                  detail: 'Table',
                  sortText: `1_${tableName}`,
                });
              }
            });
          } else {
            // Suggest columns and tables
            tableNames.forEach((tableName) => {
              // Add table name
              if (shouldShowAll || tableName.toUpperCase().startsWith(lastWord)) {
                suggestions.push({
                  label: tableName,
                  kind: CompletionItemKind.Class,
                  insertText: tableName,
                  detail: 'Table',
                  sortText: `1_${tableName}`,
                });
              }

              // Add columns for this table
              const columns = allColumns.get(tableName);
              if (columns && columns.length > 0) {
                columns.forEach((column: any) => {
                  // Add column with table prefix
                  const columnLabel = `${tableName}.${column.column_name}`;
                  if (
                    shouldShowAll ||
                    columnLabel.toUpperCase().startsWith(lastWord) ||
                    column.column_name.toUpperCase().startsWith(lastWord)
                  ) {
                    suggestions.push({
                      label: columnLabel,
                      kind: CompletionItemKind.Field,
                      insertText: columnLabel,
                      detail: `${column.data_type}${column.is_nullable ? '' : ' NOT NULL'}`,
                      documentation: `Column: ${column.column_name}\nType: ${column.data_type}\n${column.is_primary_key ? 'Primary Key\n' : ''}${column.is_foreign_key ? 'Foreign Key' : ''}`,
                      sortText: `2_${tableName}_${column.column_name}`,
                    });
                  }

                  // Also add column name alone for WHERE clauses etc
                  if (shouldShowAll || column.column_name.toUpperCase().startsWith(lastWord)) {
                    suggestions.push({
                      label: column.column_name,
                      kind: CompletionItemKind.Field,
                      insertText: column.column_name,
                      detail: `${column.data_type} (from ${tableName})`,
                      documentation: `Column: ${column.column_name}\nTable: ${tableName}\nType: ${column.data_type}`,
                      sortText: `3_${column.column_name}`,
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
  };

  const handleEditorWillMount = (monacoInstance: Monaco) => {
    setMonaco(monacoInstance);

    // Register signature help for functions
    monacoInstance.languages.registerSignatureHelpProvider('sql', {
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

  // Update completion provider when dataSourceId changes
  useEffect(() => {
    if (!monaco || !dataSourceId) return;

    console.log('Updating autocomplete for data source:', dataSourceId);

    // Dispose old provider if exists
    if (completionProviderRef.current) {
      completionProviderRef.current.dispose();
    }

    // Create new provider with current dataSourceId
    completionProviderRef.current = createCompletionProvider(monaco, dataSourceId);

    return () => {
      if (completionProviderRef.current) {
        completionProviderRef.current.dispose();
      }
    };
  }, [monaco, dataSourceId, schemas]);

  // Update editor theme when app theme changes
  useEffect(() => {
    if (!editor || !monaco) return;

    const monacoTheme = effectiveTheme === 'dark' ? 'vs-dark' : 'vs-light';
    monaco.editor.setTheme(monacoTheme);
  }, [editor, monaco, effectiveTheme]);

  const handleEditorChange = (value: string | undefined) => {
    onChange(value || '');
  };

  const handleEditorDidMount = (editorInstance: any, monacoInstance: Monaco) => {
    setEditor(editorInstance);
  };

  const monacoTheme = effectiveTheme === 'dark' ? 'vs-dark' : 'vs-light';

  return (
    <div className="border border-gray-300 dark:border-gray-700 rounded-lg overflow-hidden">
      <Editor
        height={editorHeight}
        defaultLanguage="sql"
        value={value}
        onChange={handleEditorChange}
        theme={monacoTheme}
        beforeMount={handleEditorWillMount}
        onMount={handleEditorDidMount}
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
