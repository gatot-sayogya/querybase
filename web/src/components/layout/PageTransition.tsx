'use client';

import { ReactNode, ElementType } from 'react';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import {
  fadeIn,
  slideUp,
  slideDown,
  scaleIn,
  staggerContainer,
  defaultTransition,
  reducedMotionVariants,
} from '@/lib/animations';

interface PageTransitionProps {
  children: ReactNode;
  className?: string;
  animation?: 'fade' | 'slide-up' | 'slide-down' | 'scale';
  stagger?: boolean;
  delay?: number;
  as?: ElementType;
}

const animationVariants = {
  fade: fadeIn,
  'slide-up': slideUp,
  'slide-down': slideDown,
  scale: scaleIn,
};

export default function PageTransition({
  children,
  className,
  animation = 'fade',
  stagger = false,
  delay = 0,
  as: Component = 'div',
}: PageTransitionProps) {
  const shouldReduceMotion = useReducedMotion();
  const variants = shouldReduceMotion ? reducedMotionVariants : animationVariants[animation];
  const MotionComponent = motion(Component);

  return (
    <AnimatePresence mode="wait">
      <MotionComponent
        className={cn(className)}
        variants={stagger ? staggerContainer : variants}
        initial="initial"
        animate="animate"
        exit="exit"
        transition={stagger ? {} : { ...defaultTransition, delay }}
      >
        {children}
      </MotionComponent>
    </AnimatePresence>
  );
}

export function PageTransitionWrapper({
  children,
  className,
  animation = 'fade',
  stagger = false,
  delay = 0,
}: PageTransitionProps) {
  const shouldReduceMotion = useReducedMotion();
  const variants = shouldReduceMotion ? reducedMotionVariants : animationVariants[animation];

  return (
    <motion.div
      className={cn(className)}
      variants={stagger ? staggerContainer : variants}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={stagger ? {} : { ...defaultTransition, delay }}
    >
      {children}
    </motion.div>
  );
}

export function StaggerContainer({
  children,
  className,
  staggerDelay = 0.04,
}: {
  children: ReactNode;
  className?: string;
  staggerDelay?: number;
}) {
  const shouldReduceMotion = useReducedMotion();

  if (shouldReduceMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      className={className}
      initial="initial"
      animate="animate"
      exit="exit"
      variants={{
        initial: {},
        animate: {
          transition: {
            staggerChildren: staggerDelay,
            delayChildren: 0.05,
          },
        },
        exit: {
          transition: {
            staggerChildren: 0.02,
            staggerDirection: -1,
          },
        },
      }}
    >
      {children}
    </motion.div>
  );
}

export function StaggerItem({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const shouldReduceMotion = useReducedMotion();

  if (shouldReduceMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      className={className}
      variants={{
        initial: { opacity: 0, y: 12 },
        animate: { opacity: 1, y: 0 },
        exit: { opacity: 0, y: -8 },
      }}
      transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
    >
      {children}
    </motion.div>
  );
}
