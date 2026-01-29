import { ReactNode } from 'react';
import { cn } from '@/lib/utils';
import {
  DocumentIcon,
  MagnifyingGlassIcon,
  ServerIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  illustration?: 'no-data' | 'no-results' | 'no-connections' | 'error';
  className?: string;
}

export function EmptyState({
  icon,
  title,
  description,
  action,
  illustration,
  className,
}: EmptyStateProps) {
  const illustrations = {
    'no-data': (
      <DocumentIcon className="w-16 h-16 mx-auto text-gray-400 dark:text-gray-500" />
    ),
    'no-results': (
      <MagnifyingGlassIcon className="w-16 h-16 mx-auto text-gray-400 dark:text-gray-500" />
    ),
    'no-connections': (
      <ServerIcon className="w-16 h-16 mx-auto text-gray-400 dark:text-gray-500" />
    ),
    'error': (
      <ExclamationTriangleIcon className="w-16 h-16 mx-auto text-red-400 dark:text-red-500" />
    ),
  };

  return (
    <div className={cn('text-center py-12 px-4', className)}>
      {illustration && illustrations[illustration]}
      {icon && <div className="mb-4">{icon}</div>}
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mt-4">
        {title}
      </h3>
      {description && (
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 max-w-sm mx-auto">
          {description}
        </p>
      )}
      {action && (
        <div className="mt-6">
          <button
            onClick={action.onClick}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-blue-500 dark:hover:bg-blue-600"
          >
            {action.label}
          </button>
        </div>
      )}
    </div>
  );
}
