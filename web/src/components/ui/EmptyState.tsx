'use client';

import { ReactNode } from 'react';
import { motion, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { scaleInBounce, fadeIn, springConfig, duration, reducedMotionVariants } from '@/lib/animations';
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
  const shouldReduceMotion = useReducedMotion();

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

  const iconVariants = shouldReduceMotion ? reducedMotionVariants : scaleInBounce;
  const textVariants = shouldReduceMotion ? reducedMotionVariants : fadeIn;

  return (
    <div className={cn('text-center py-12 px-4', className)}>
      {illustration && (
        <motion.div
          variants={iconVariants}
          initial="initial"
          animate="animate"
          transition={{ ...springConfig.bouncy, duration: duration.slow }}
        >
          {illustrations[illustration]}
        </motion.div>
      )}
      {icon && (
        <motion.div
          className="mb-4"
          variants={iconVariants}
          initial="initial"
          animate="animate"
          transition={{ ...springConfig.bouncy, duration: duration.slow }}
        >
          {icon}
        </motion.div>
      )}
      <motion.h3
        className="text-lg font-medium text-gray-900 dark:text-white mt-4"
        variants={textVariants}
        initial="initial"
        animate="animate"
        transition={{ duration: duration.normal, delay: 0.1 }}
      >
        {title}
      </motion.h3>
      {description && (
        <motion.p
          className="mt-2 text-sm text-gray-500 dark:text-gray-400 max-w-sm mx-auto"
          variants={textVariants}
          initial="initial"
          animate="animate"
          transition={{ duration: duration.normal, delay: 0.15 }}
        >
          {description}
        </motion.p>
      )}
      {action && (
        <motion.div
          className="mt-6"
          variants={textVariants}
          initial="initial"
          animate="animate"
          transition={{ duration: duration.normal, delay: 0.2 }}
        >
          <motion.button
            onClick={action.onClick}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-blue-500 dark:hover:bg-blue-600"
            whileHover={shouldReduceMotion ? {} : { scale: 1.02 }}
            whileTap={shouldReduceMotion ? {} : { scale: 0.98 }}
            transition={springConfig.snappy}
          >
            {action.label}
          </motion.button>
        </motion.div>
      )}
    </div>
  );
}
