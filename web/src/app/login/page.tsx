'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import Button from '@/components/ui/Button';
import Card from '@/components/ui/Card';
import PageTransition from '@/components/layout/PageTransition';

export default function LoginPage() {
  const router = useRouter();
  const { login, isLoading, error } = useAuthStore();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(username, password);
      router.push('/dashboard');
    } catch (err) {
      // Error is handled by the store
    }
  };

  return (
    <PageTransition>
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 relative overflow-hidden py-12 px-4 sm:px-6 lg:px-8">
        {/* Background decorations */}
        <div className="absolute top-[-10%] left-[-10%] w-[500px] h-[500px] rounded-full bg-blue-500/5 blur-[80px] pointer-events-none" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[500px] h-[500px] rounded-full bg-indigo-500/5 blur-[80px] pointer-events-none" />

        <div className="w-full max-w-md space-y-8 relative z-10">
          <div className="text-center animate-slide-down">
            <h2 className="mt-6 text-3xl font-extrabold text-gray-900 dark:text-white tracking-tight">
              Welcome back
            </h2>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
              Sign in to manage your databases
            </p>
          </div>

          <Card variant="glass" className="p-8 backdrop-blur-xl bg-white/70 dark:bg-gray-800/70 shadow-2xl animate-scale-in border-white/50 dark:border-gray-700/50">
            <form className="space-y-6" onSubmit={handleSubmit}>
              {error && (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 px-4 py-3 rounded-md text-sm animate-shake">
                  {error}
                </div>
              )}
              <div className="space-y-4">
                <div className="group">
                  <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-400 transition-colors">
                    Username
                  </label>
                  <input
                    id="username"
                    name="username"
                    type="text"
                    required
                    className="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-900/50 dark:text-white transition-all duration-200"
                    placeholder="Enter your username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                  />
                </div>
                <div className="group">
                  <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-400 transition-colors">
                    Password
                  </label>
                  <input
                    id="password"
                    name="password"
                    type="password"
                    required
                    className="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-900/50 dark:text-white transition-all duration-200"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
              </div>

              <div>
                <Button
                  type="submit"
                  loading={isLoading}
                  className="w-full shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30"
                  size="lg"
                >
                  Sign in
                </Button>
              </div>

              <div className="mt-6 text-center">
                <p className="text-xs text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-900/50 py-2 px-3 rounded-md inline-block border border-gray-100 dark:border-gray-700">
                  Demo Credentials: <span className="font-mono font-medium text-gray-700 dark:text-gray-300">admin / admin123</span>
                </p>
              </div>
            </form>
          </Card>
        </div>
      </div>
    </PageTransition>
  );
}
