'use client';

import { useEffect, useState } from 'react';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { toastSlide, springConfig, duration, reducedMotionVariants } from '@/lib/animations';
import { CheckCircleIcon, XCircleIcon, InformationCircleIcon, ExclamationTriangleIcon, XMarkIcon } from '@heroicons/react/24/outline';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface ToastProps {
    id: string;
    type: ToastType;
    title: string;
    message?: string;
    duration?: number;
    onDismiss: (id: string) => void;
}

export default function Toast({ id, type, title, message, duration: toastDuration = 5000, onDismiss }: ToastProps) {
    const [isExiting, setIsExiting] = useState(false);
    const [progress, setProgress] = useState(100);
    const shouldReduceMotion = useReducedMotion();

    useEffect(() => {
        const timer = setTimeout(() => {
            handleDismiss();
        }, toastDuration);

        const interval = setInterval(() => {
            setProgress((prev) => Math.max(0, prev - (100 / (toastDuration / 100))));
        }, 100);

        return () => {
            clearTimeout(timer);
            clearInterval(interval);
        };
    }, [toastDuration]);

    const handleDismiss = () => {
        setIsExiting(true);
        setTimeout(() => {
            onDismiss(id);
        }, 200);
    };

    const icons = {
        success: <CheckCircleIcon className="w-6 h-6 text-green-500" />,
        error: <XCircleIcon className="w-6 h-6 text-red-500" />,
        info: <InformationCircleIcon className="w-6 h-6 text-blue-500" />,
        warning: <ExclamationTriangleIcon className="w-6 h-6 text-amber-500" />,
    };

    const styles = {
        success: 'border-l-4 border-l-green-500',
        error: 'border-l-4 border-l-red-500',
        info: 'border-l-4 border-l-blue-500',
        warning: 'border-l-4 border-l-amber-500',
    };

    const variants = shouldReduceMotion ? reducedMotionVariants : toastSlide;
    const transition = shouldReduceMotion
        ? { duration: duration.fast }
        : { ...springConfig.snappy, duration: duration.slow };

    return (
        <motion.div
            className={cn(
                "pointer-events-auto w-full max-w-sm overflow-hidden bg-white dark:bg-gray-800 rounded-lg shadow-lg ring-1 ring-black ring-opacity-5",
                styles[type]
            )}
            variants={variants}
            initial="initial"
            animate={isExiting ? "exit" : "animate"}
            transition={transition}
            role="alert"
        >
            <div className="p-4">
                <div className="flex items-start">
                    <div className="flex-shrink-0">
                        {icons[type]}
                    </div>
                    <div className="ml-3 w-0 flex-1 pt-0.5">
                        <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                            {title}
                        </p>
                        {message && (
                            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                                {message}
                            </p>
                        )}
                    </div>
                    <div className="ml-4 flex flex-shrink-0">
                        <button
                            type="button"
                            className="inline-flex rounded-md bg-white dark:bg-gray-800 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                            onClick={handleDismiss}
                        >
                            <span className="sr-only">Close</span>
                            <XMarkIcon className="h-5 w-5" aria-hidden="true" />
                        </button>
                    </div>
                </div>
            </div>
            <div
                className={cn(
                    "h-1 transition-all duration-100 linear",
                    type === 'success' && "bg-green-500",
                    type === 'error' && "bg-red-500",
                    type === 'info' && "bg-blue-500",
                    type === 'warning' && "bg-amber-500"
                )}
                style={{ width: `${progress}%` }}
            />
        </motion.div>
    );
}

export function ToastContainer({ children }: { children: React.ReactNode }) {
    return (
        <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-3 pointer-events-none">
            <AnimatePresence mode="popLayout">
                {children}
            </AnimatePresence>
        </div>
    );
}
