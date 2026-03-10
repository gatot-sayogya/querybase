'use client';

import { motion, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { staggerItem, defaultTransition } from '@/lib/animations';

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular' | 'rounded';
  width?: string | number;
  height?: string | number;
  animate?: boolean;
}

export function Skeleton({
  className,
  variant = 'rectangular',
  width,
  height,
  animate = true,
  ...props
}: SkeletonProps) {
  const variantStyles = {
    text: 'rounded max-w-full',
    circular: 'rounded-full',
    rectangular: 'rounded-none',
    rounded: 'rounded-md',
  };

  return (
    <div
      className={cn(
        'relative overflow-hidden bg-gray-200 dark:bg-gray-700',
        animate && 'animate-pulse',
        variantStyles[variant],
        className
      )}
      style={{ width, height }}
      {...props}
    >
      {animate && (
        <div
          className="absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/20 to-transparent animate-shimmer"
          style={{ animation: 'shimmer 2s infinite' }}
        />
      )}
    </div>
  );
}

interface TableSkeletonProps {
  rows?: number;
  columns?: number;
  animateRows?: boolean;
}

export function TableSkeleton({ rows = 5, columns = 4, animateRows = true }: TableSkeletonProps) {
  const shouldReduceMotion = useReducedMotion();

  const rowContent = (rowIndex: number) => (
    <div className="flex gap-4 py-2">
      {Array.from({ length: columns }).map((_, j) => (
        <Skeleton key={j} variant="text" width="20%" height={16} />
      ))}
    </div>
  );

  if (!animateRows || shouldReduceMotion) {
    return (
      <div className="space-y-3">
        <div className="flex gap-4">
          {Array.from({ length: columns }).map((_, i) => (
            <Skeleton key={i} variant="text" width="20%" height={20} />
          ))}
        </div>
        {Array.from({ length: rows }).map((_, i) => rowContent(i))}
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex gap-4">
        {Array.from({ length: columns }).map((_, i) => (
          <Skeleton key={i} variant="text" width="20%" height={20} />
        ))}
      </div>
      <motion.div
        initial="initial"
        animate="animate"
        variants={{
          initial: {},
          animate: {
            transition: {
              staggerChildren: 0.05,
            },
          },
        }}
      >
        {Array.from({ length: rows }).map((_, i) => (
          <motion.div
            key={i}
            variants={staggerItem}
            transition={defaultTransition}
          >
            {rowContent(i)}
          </motion.div>
        ))}
      </motion.div>
    </div>
  );
}

interface CardSkeletonProps {
  count?: number;
  animateCards?: boolean;
}

export function CardSkeleton({ count = 3, animateCards = true }: CardSkeletonProps) {
  const shouldReduceMotion = useReducedMotion();

  const cardContent = (index: number) => (
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-3">
      <div className="flex items-center space-x-3">
        <Skeleton variant="circular" width={40} height={40} />
        <div className="flex-1 space-y-2">
          <Skeleton variant="text" width="60%" height={16} />
          <Skeleton variant="text" width="40%" height={14} />
        </div>
      </div>
      <div className="space-y-2">
        <Skeleton variant="text" width="100%" height={14} />
        <Skeleton variant="text" width="80%" height={14} />
      </div>
    </div>
  );

  if (!animateCards || shouldReduceMotion) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: count }).map((_, i) => (
          <div key={i}>{cardContent(i)}</div>
        ))}
      </div>
    );
  }

  return (
    <motion.div
      className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
      initial="initial"
      animate="animate"
      variants={{
        initial: {},
        animate: {
          transition: {
            staggerChildren: 0.08,
          },
        },
      }}
    >
      {Array.from({ length: count }).map((_, i) => (
        <motion.div
          key={i}
          variants={staggerItem}
          transition={defaultTransition}
        >
          {cardContent(i)}
        </motion.div>
      ))}
    </motion.div>
  );
}
