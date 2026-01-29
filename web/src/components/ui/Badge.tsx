import { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface BadgeProps {
  children: ReactNode;
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'error' | 'info';
  size?: 'xs' | 'sm' | 'md';
  icon?: ReactNode;
  dot?: boolean;
  className?: string;
}

export function Badge({
  children,
  variant = 'default',
  size = 'sm',
  icon,
  dot = false,
  className,
}: BadgeProps) {
  const variants = {
    default: 'bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800 dark:text-gray-200 dark:border-gray-700',
    primary: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-700',
    success: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-700',
    warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-300 dark:border-yellow-700',
    error: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900/30 dark:text-red-300 dark:border-red-700',
    info: 'bg-cyan-100 text-cyan-800 border-cyan-200 dark:bg-cyan-900/30 dark:text-cyan-300 dark:border-cyan-700',
  };

  const sizes = {
    xs: 'px-1.5 py-0.5 text-xs',
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-sm',
  };

  const dotSizes = {
    xs: 'w-1 h-1',
    sm: 'w-1.5 h-1.5',
    md: 'w-2 h-2',
  };

  const dotColors = {
    default: 'bg-gray-400 dark:bg-gray-500',
    primary: 'bg-blue-400 dark:bg-blue-500',
    success: 'bg-green-400 dark:bg-green-500',
    warning: 'bg-yellow-400 dark:bg-yellow-500',
    error: 'bg-red-400 dark:bg-red-500',
    info: 'bg-cyan-400 dark:bg-cyan-500',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 border font-medium rounded-full transition-colors',
        variants[variant],
        sizes[size],
        className
      )}
    >
      {dot && (
        <span className={cn('rounded-full animate-pulse', dotSizes[size], dotColors[variant])} />
      )}
      {icon && <span className="flex-shrink-0">{icon}</span>}
      <span>{children}</span>
    </span>
  );
}

interface StatusBadgeProps {
  status: 'pending' | 'running' | 'completed' | 'failed';
  className?: string;
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const statusConfig = {
    pending: {
      variant: 'warning' as const,
      label: 'Pending',
      dot: true,
    },
    running: {
      variant: 'info' as const,
      label: 'Running',
      dot: true,
    },
    completed: {
      variant: 'success' as const,
      label: 'Completed',
      dot: false,
    },
    failed: {
      variant: 'error' as const,
      label: 'Failed',
      dot: false,
    },
  };

  const config = statusConfig[status];

  return (
    <Badge variant={config.variant} dot={config.dot} className={className}>
      {config.label}
    </Badge>
  );
}
