import { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';
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

export default function Toast({ id, type, title, message, duration = 5000, onDismiss }: ToastProps) {
    const [isExiting, setIsExiting] = useState(false);
    const [progress, setProgress] = useState(100);

    useEffect(() => {
        const timer = setTimeout(() => {
            handleDismiss();
        }, duration);

        // Progress bar animation
        const interval = setInterval(() => {
            setProgress((prev) => Math.max(0, prev - (100 / (duration / 100))));
        }, 100);

        return () => {
            clearTimeout(timer);
            clearInterval(interval);
        };
    }, [duration]); // Removed handleDismiss from deps to avoid infinite loop

    const handleDismiss = () => {
        setIsExiting(true);
        setTimeout(() => {
            onDismiss(id);
        }, 300); // Wait for exit animation
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

    return (
        <div
            className={cn(
                "pointer-events-auto w-full max-w-sm overflow-hidden bg-white dark:bg-gray-800 rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 transition-all duration-300 transform",
                isExiting ? "translate-x-full opacity-0" : "animate-slide-left translate-x-0 opacity-100",
                styles[type]
            )}
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
        </div>
    );
}
