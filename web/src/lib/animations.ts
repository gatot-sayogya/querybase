import { Variants, Transition } from 'framer-motion';

export const springConfig = {
  gentle: { stiffness: 200, damping: 25, mass: 1 },
  snappy: { stiffness: 400, damping: 30, mass: 0.8 },
  bouncy: { stiffness: 600, damping: 15, mass: 0.5 },
  stiff: { stiffness: 300, damping: 30, mass: 1 },
  micro: { stiffness: 500, damping: 25, mass: 0.5 },
  smooth: { stiffness: 200, damping: 22, mass: 1 },
} as const;

export const duration = {
  fast: 0.15,
  normal: 0.2,
  slow: 0.3,
  page: 0.4,
} as const;

export const easing = {
  easeOut: [0.16, 1, 0.3, 1] as const,
  easeIn: [0.4, 0, 1, 1] as const,
  easeInOut: [0.4, 0, 0.2, 1] as const,
} as const;

export const fadeIn: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
};

export const slideUp: Variants = {
  initial: { opacity: 0, y: 16 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
};

export const slideDown: Variants = {
  initial: { opacity: 0, y: -16 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: 8 },
};

export const slideLeft: Variants = {
  initial: { opacity: 0, x: 24 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: -16 },
};

export const slideRight: Variants = {
  initial: { opacity: 0, x: -24 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: 16 },
};

export const scaleIn: Variants = {
  initial: { opacity: 0, scale: 0.95 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 0.95 },
};

export const scaleInBounce: Variants = {
  initial: { opacity: 0, scale: 0.9 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 0.9 },
};

export const staggerContainer: Variants = {
  initial: {},
  animate: {
    transition: {
      staggerChildren: 0.04,
      delayChildren: 0.05,
    },
  },
  exit: {
    transition: {
      staggerChildren: 0.02,
      staggerDirection: -1,
    },
  },
};

export const staggerContainerSlow: Variants = {
  initial: {},
  animate: {
    transition: {
      staggerChildren: 0.08,
      delayChildren: 0.1,
    },
  },
  exit: {
    transition: {
      staggerChildren: 0.03,
      staggerDirection: -1,
    },
  },
};

export const staggerItem: Variants = {
  initial: { opacity: 0, y: 12 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
};

export const staggerItemScale: Variants = {
  initial: { opacity: 0, scale: 0.95 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 0.95 },
};

export const modalBackdrop: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
};

export const modalContent: Variants = {
  initial: { opacity: 0, scale: 0.95, y: 8 },
  animate: { opacity: 1, scale: 1, y: 0 },
  exit: { opacity: 0, scale: 0.95, y: 8 },
};

export const toastSlide: Variants = {
  initial: { opacity: 0, x: 100, scale: 0.9 },
  animate: { opacity: 1, x: 0, scale: 1 },
  exit: { opacity: 0, x: 100, scale: 0.9 },
};

export const shake: Variants = {
  initial: { x: 0 },
  animate: {
    x: [0, -8, 8, -6, 6, -4, 4, 0],
    transition: { duration: 0.4 },
  },
};

export const pulse: Variants = {
  initial: { scale: 1 },
  animate: {
    scale: [1, 1.05, 1],
    transition: {
      duration: 2,
      repeat: Infinity,
      ease: 'easeInOut',
    },
  },
};

export const countUp = (
  value: number,
  animDuration: number = 0.8
): { duration: number; ease: string } => ({
  duration: animDuration,
  ease: 'easeOut',
});

export const defaultTransition: Transition = {
  duration: duration.normal,
  ease: easing.easeOut,
};

export const pageTransition: Transition = {
  ...springConfig.gentle,
  duration: duration.page,
};

export const getStaggerDelay = (index: number, baseDelay: number = 0.05): number => {
  return baseDelay * index;
};

export const createStaggerVariants = (staggerDelay: number = 0.04): Variants => ({
  initial: {},
  animate: {
    transition: {
      staggerChildren: staggerDelay,
    },
  },
  exit: {
    transition: {
      staggerChildren: 0.02,
      staggerDirection: -1,
    },
  },
});

export const reducedMotionVariants: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
};
