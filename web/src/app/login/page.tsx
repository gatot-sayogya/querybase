'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { staggerContainer, staggerItem, shake, springConfig, duration, reducedMotionVariants } from '@/lib/animations';

export default function LoginPage() {
  const router = useRouter();
  const { login, isLoading, error } = useAuthStore();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const shouldReduceMotion = useReducedMotion();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(username, password);
      router.push('/dashboard');
    } catch (err) {
      // Error is handled by the store
    }
  };

  const containerVariants = shouldReduceMotion ? reducedMotionVariants : staggerContainer;
  const itemVariants = shouldReduceMotion ? reducedMotionVariants : staggerItem;
  const errorVariants = shouldReduceMotion ? reducedMotionVariants : shake;

  return (
    <div className="min-h-screen flex items-center justify-center p-6 relative overflow-hidden bg-slate-50 dark:bg-slate-900">
      <div
        style={{
          position: 'fixed',
          top: '-120px',
          left: '50%',
          transform: 'translateX(-20%)',
          width: '640px',
          height: '640px',
          background: 'radial-gradient(circle, rgba(37, 99, 235, 0.25) 0%, transparent 70%)',
          pointerEvents: 'none',
        }}
      />

      <motion.div
        className="login-card"
        variants={containerVariants}
        initial="initial"
        animate="animate"
        transition={{ duration: duration.page }}
      >
        <motion.div className="logo-row" variants={itemVariants}>
          <motion.div
            className="logo-mark"
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ ...springConfig.bouncy, delay: 0.1 }}
          >
            Q
          </motion.div>
          <motion.span
            className="logo-name"
            initial={{ x: -10, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            transition={{ delay: 0.2, duration: duration.normal }}
          >
            QueryBase
          </motion.span>
        </motion.div>

        <motion.div className="title-block" variants={itemVariants}>
          <h1>Welcome back</h1>
          <p>Sign in to your account to continue</p>
        </motion.div>

        <form className="form-fields" onSubmit={handleSubmit}>
          <motion.div className="form-group" variants={itemVariants}>
            <label className="form-label" htmlFor="username">
              Username
            </label>
            <input
              id="username"
              name="username"
              className="form-input"
              type="text"
              required
              placeholder="Enter username"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </motion.div>

          <motion.div className="form-group" variants={itemVariants}>
            <label className="form-label" htmlFor="password">
              Password
            </label>
            <input
              id="password"
              name="password"
              className="form-input"
              type="password"
              required
              placeholder="••••••••"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </motion.div>

          <AnimatePresence mode="wait">
            {error && (
              <motion.div
                className="error-msg show"
                variants={errorVariants}
                initial="initial"
                animate="animate"
                exit={{ opacity: 0, x: 20 }}
                transition={{ duration: 0.4 }}
              >
                <svg className="w-5 h-5 inline mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                {error}
              </motion.div>
            )}
          </AnimatePresence>

          <motion.div variants={itemVariants}>
            <motion.button
              type="submit"
              className="signin-btn"
              disabled={isLoading}
              whileHover={!shouldReduceMotion && !isLoading ? { scale: 1.02 } : {}}
              whileTap={!shouldReduceMotion && !isLoading ? { scale: 0.98 } : {}}
              transition={springConfig.snappy}
            >
              {isLoading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Signing in...
                </span>
              ) : (
                'Sign in'
              )}
            </motion.button>
          </motion.div>

          <motion.div className="mt-2 text-center" variants={itemVariants}>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Demo Credentials: <span className="font-mono">admin / admin123</span>
            </p>
          </motion.div>
        </form>

        <motion.p
          className="card-footer"
          variants={itemVariants}
        >
          Secure database access for your team
        </motion.p>
      </motion.div>
    </div>
  );
}
