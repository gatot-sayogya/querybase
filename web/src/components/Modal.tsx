import { ReactNode, useEffect } from 'react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export default function Modal({ isOpen, onClose, title, children }: ModalProps) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6">
      <div 
        className="absolute inset-0 bg-[var(--bg-dark)]/80 backdrop-blur-sm transition-opacity" 
        onClick={onClose}
      />
      
      <div className="relative w-full max-w-4xl max-h-[90vh] flex flex-col bg-[var(--bg-page)] shadow-2xl overflow-hidden border border-[var(--border-light)]" style={{ borderRadius: '2px' }}>
        <div className="flex items-center justify-between px-8 py-6 border-b border-[var(--border-light)] bg-transparent">
          <h2 className="text-xl font-black tracking-widest uppercase text-[var(--text-primary)]">
            {title}
          </h2>
          <button 
            onClick={onClose}
            className="p-2 text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>
        
        <div className="p-8 overflow-y-auto hide-scrollbar">
          {children}
        </div>
      </div>
    </div>
  );
}
