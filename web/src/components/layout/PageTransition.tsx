'use client';

import { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface PageTransitionProps {
    children: ReactNode;
    className?: string;
    animation?: 'fade' | 'slide-up' | 'slide-down' | 'scale';
}

export default function PageTransition({
    children,
    className,
    animation = 'fade'
}: PageTransitionProps) {
    const animations = {
        fade: 'animate-fade-in',
        'slide-up': 'animate-slide-up',
        'slide-down': 'animate-slide-down',
        scale: 'animate-scale-in',
    };

    return (
        <div className={cn(animations[animation], className)}>
            {children}
        </div>
    );
}
