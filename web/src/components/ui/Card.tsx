'use client';

import { forwardRef, HTMLAttributes } from 'react';
import { cn } from '@/lib/utils';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'glass' | 'elevated' | 'interactive';
  padding?: 'none' | 'sm' | 'md' | 'lg';
}

const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = 'default', padding = 'md', children, ...props }, ref) => {
    const variants = {
      default: 'bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 sleek-shadow',
      glass: 'glass sleek-shadow',
      elevated: 'bg-white dark:bg-slate-900 border-none shadow-glass float-shadow',
      interactive: 'bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 sleek-shadow cursor-pointer hover:shadow-glass hover:-translate-y-1 transition-all duration-300 ease-spring',
    };

    const paddings = {
      none: '',
      sm: 'p-3',
      md: 'p-6',
      lg: 'p-8',
    };

    return (
      <div
        ref={ref}
        className={cn('rounded-3xl overflow-hidden', variants[variant], paddings[padding], className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Card.displayName = 'Card';

export default Card;
