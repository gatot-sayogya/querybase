'use client';

import Editor from '@monaco-editor/react';
import { useState } from 'react';

interface SQLEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  readOnly?: boolean;
  height?: string;
}

export default function SQLEditor({
  value,
  onChange,
  placeholder = 'SELECT * FROM users LIMIT 10;',
  readOnly = false,
  height = '400px',
}: SQLEditorProps) {
  const [editorHeight] = useState(height);

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
        }}
      />
    </div>
  );
}
