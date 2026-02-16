import { cn } from '@/lib/utils';

export interface LoadingProps {
    variant?: 'spinner' | 'dots' | 'bars' | 'skeleton';
    size?: 'sm' | 'md' | 'lg' | 'xl';
    className?: string;
    text?: string;
}

export default function Loading({ variant = 'spinner', size = 'md', className, text }: LoadingProps) {
    const sizeClasses = {
        sm: 'w-4 h-4',
        md: 'w-8 h-8',
        lg: 'w-12 h-12',
        xl: 'w-16 h-16',
    };

    const renderVariant = () => {
        switch (variant) {
            case 'dots':
                return (
                    <div className={cn("flex space-x-1", className)}>
                        <div className={cn("bg-blue-600 rounded-full animate-bounce [animation-delay:-0.3s]", sizeClasses[size])}></div>
                        <div className={cn("bg-blue-600 rounded-full animate-bounce [animation-delay:-0.15s]", sizeClasses[size])}></div>
                        <div className={cn("bg-blue-600 rounded-full animate-bounce", sizeClasses[size])}></div>
                    </div>
                );
            case 'bars':
                return (
                    <div className={cn("flex space-x-1 items-end", className)} style={{ height: size === 'sm' ? 16 : size === 'md' ? 32 : size === 'lg' ? 48 : 64 }}>
                        <div className={cn("bg-blue-600 w-1/4 animate-subtle-bounce", "h-1/2")}></div>
                        <div className={cn("bg-blue-600 w-1/4 animate-subtle-bounce [animation-delay:0.1s]", "h-3/4")}></div>
                        <div className={cn("bg-blue-600 w-1/4 animate-subtle-bounce [animation-delay:0.2s]", "h-full")}></div>
                    </div>
                );
            case 'skeleton':
                return (
                    <div className={cn("animate-shimmer bg-gradient-to-r from-gray-200 via-gray-100 to-gray-200 dark:from-gray-800 dark:via-gray-700 dark:to-gray-800 bg-[length:1000px_100%] rounded", sizeClasses[size], className)} />
                );
            case 'spinner':
            default:
                return (
                    <svg
                        className={cn("animate-spin text-blue-600", sizeClasses[size], className)}
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                    >
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                );
        }
    };

    return (
        <div className="flex flex-col items-center justify-center p-4">
            {renderVariant()}
            {text && (
                <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 font-medium animate-pulse">
                    {text}
                </p>
            )}
        </div>
    );
}
