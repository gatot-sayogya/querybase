import { ReactNode } from 'react';
import { cn } from '@/lib/utils';
import {
  ExclamationTriangleIcon,
  InformationCircleIcon,
  CheckCircleIcon,
  XCircleIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';

interface AlertProps {
  variant?: 'error' | 'warning' | 'success' | 'info';
  title?: string;
  children: ReactNode;
  dismissible?: boolean;
  onDismiss?: () => void;
  className?: string;
}

export function Alert({
  variant = 'info',
  title,
  children,
  dismissible = false,
  onDismiss,
  className,
}: AlertProps) {
  const variants = {
    error: {
      container: 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800',
      icon: 'text-red-400 dark:text-red-300',
      title: 'text-red-800 dark:text-red-400',
      body: 'text-red-700 dark:text-red-300',
      closeButton: 'text-red-400 hover:text-red-600 dark:hover:text-red-200 hover:bg-red-100 dark:hover:bg-red-900/30',
    },
    warning: {
      container: 'bg-yellow-50 border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800',
      icon: 'text-yellow-400 dark:text-yellow-300',
      title: 'text-yellow-800 dark:text-yellow-400',
      body: 'text-yellow-700 dark:text-yellow-300',
      closeButton: 'text-yellow-400 hover:text-yellow-600 dark:hover:text-yellow-200 hover:bg-yellow-100 dark:hover:bg-yellow-900/30',
    },
    success: {
      container: 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-800',
      icon: 'text-green-400 dark:text-green-300',
      title: 'text-green-800 dark:text-green-400',
      body: 'text-green-700 dark:text-green-300',
      closeButton: 'text-green-400 hover:text-green-600 dark:hover:text-green-200 hover:bg-green-100 dark:hover:bg-green-900/30',
    },
    info: {
      container: 'bg-blue-50 border-blue-200 dark:bg-blue-900/20 dark:border-blue-800',
      icon: 'text-blue-400 dark:text-blue-300',
      title: 'text-blue-800 dark:text-blue-400',
      body: 'text-blue-700 dark:text-blue-300',
      closeButton: 'text-blue-400 hover:text-blue-600 dark:hover:text-blue-200 hover:bg-blue-100 dark:hover:bg-blue-900/30',
    },
  };

  const icons = {
    error: ExclamationTriangleIcon,
    warning: ExclamationTriangleIcon,
    success: CheckCircleIcon,
    info: InformationCircleIcon,
  };

  const Icon = icons[variant];
  const styles = variants[variant];

  return (
    <div className={cn('border rounded-lg p-4', styles.container, className)}>
      <div className="flex">
        <div className="flex-shrink-0">
          <Icon className={cn('h-5 w-5', styles.icon)} aria-hidden="true" />
        </div>
        <div className="ml-3 flex-1">
          {title && (
            <h3 className={cn('text-sm font-medium', styles.title)}>
              {title}
            </h3>
          )}
          <div className={cn('text-sm', title ? 'mt-1' : '', styles.body)}>
            {children}
          </div>
        </div>
        {dismissible && onDismiss && (
          <div className="ml-auto pl-3">
            <button
              onClick={onDismiss}
              className={cn(
                'inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors',
                styles.closeButton
              )}
            >
              <XMarkIcon className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

interface QueryErrorProps {
  error: string;
  suggestions?: string[];
  onRetry?: () => void;
}

export function QueryError({ error, suggestions = [], onRetry }: QueryErrorProps) {
  const defaultSuggestions = [
    'Check if the data source connection is active',
    'Verify your SQL syntax is correct',
    'Ensure you have permission to access the tables',
  ];

  const allSuggestions = suggestions.length > 0 ? suggestions : defaultSuggestions;

  return (
    <Alert variant="error" title="Query Execution Failed">
      <div className="space-y-3">
        <p className="font-mono text-xs sm:text-sm bg-white/50 dark:bg-black/20 p-2 rounded overflow-x-auto">
          {error}
        </p>

        {allSuggestions.length > 0 && (
          <div>
            <p className="text-sm font-medium mb-2">Suggestions:</p>
            <ul className="list-disc list-inside text-sm space-y-1">
              {allSuggestions.map((suggestion, i) => (
                <li key={i}>{suggestion}</li>
              ))}
            </ul>
          </div>
        )}

        {onRetry && (
          <div className="pt-2">
            <button
              onClick={onRetry}
              className="text-sm font-medium underline hover:no-underline focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 rounded"
            >
              Try again
            </button>
          </div>
        )}
      </div>
    </Alert>
  );
}
