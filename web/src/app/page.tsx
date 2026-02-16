import Link from 'next/link';
import PageTransition from '@/components/layout/PageTransition';
import Button from '@/components/ui/Button';
import Card from '@/components/ui/Card';
import { CircleStackIcon, ShieldCheckIcon, BoltIcon, ChartBarIcon } from '@heroicons/react/24/outline';

export default function Home() {
  const features = [
    {
      title: 'SQL Editor',
      description: 'Powerful Monaco-based editor with intelligent autocomplete for tables and columns.',
      icon: <CircleStackIcon className="w-6 h-6 text-blue-500" />,
      delay: 'delay-100',
    },
    {
      title: 'Approval Workflow',
      description: 'Secure write operations with a robust single-stage approval process.',
      icon: <ShieldCheckIcon className="w-6 h-6 text-green-500" />,
      delay: 'delay-200',
    },
    {
      title: 'Performant',
      description: 'Built with Go and Redis for blazing fast query execution and background processing.',
      icon: <BoltIcon className="w-6 h-6 text-amber-500" />,
      delay: 'delay-300',
    },
    {
      title: 'Visual Results',
      description: 'Interactive data grid with sorting, pagination, and export capabilities.',
      icon: <ChartBarIcon className="w-6 h-6 text-purple-500" />,
      delay: 'delay-400',
    },
  ];

  return (
    <PageTransition>
      <main className="min-h-screen relative overflow-hidden bg-gray-50 dark:bg-gray-900">
        {/* Background Gradients */}
        <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0 pointer-events-none">
          <div className="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] rounded-full bg-blue-500/10 blur-[100px]" />
          <div className="absolute bottom-[-10%] left-[-10%] w-[600px] h-[600px] rounded-full bg-indigo-500/10 blur-[100px]" />
        </div>

        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16 relative z-10">
          <div className="text-center max-w-3xl mx-auto mb-16 animate-slide-up">
            <div className="inline-flex items-center px-3 py-1 rounded-full bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 text-sm font-medium mb-6 border border-blue-100 dark:border-blue-800">
              <span className="flex h-2 w-2 rounded-full bg-blue-600 mr-2 animate-pulse"></span>
              v1.0 is now available
            </div>

            <h1 className="text-5xl md:text-6xl font-bold mb-6 tracking-tight text-gray-900 dark:text-white">
              Database exploration <br />
              <span className="text-gradient">reimagined</span>
            </h1>

            <p className="text-xl text-gray-600 dark:text-gray-300 mb-10 leading-relaxed">
              QueryBase provides a secure, modern interface for your databases with built-in approval workflows for write operations.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link href="/login">
                <Button size="lg" className="w-full sm:w-auto shadow-lg shadow-blue-500/20">
                  Get Started
                </Button>
              </Link>

              <Link href="https://github.com/yourorg/querybase" target="_blank">
                <Button variant="outline" size="lg" className="w-full sm:w-auto">
                  View on GitHub
                </Button>
              </Link>
            </div>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mt-20">
            {features.map((feature, index) => (
              <Card
                key={index}
                variant="interactive"
                className={`h-full animate-slide-up bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm border-gray-200/50 dark:border-gray-700/50 ${feature.delay}`}
              >
                <div className="flex flex-col h-full">
                  <div className="mb-4 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg w-fit">
                    {feature.icon}
                  </div>
                  <h3 className="text-lg font-semibold mb-2 text-gray-900 dark:text-white">
                    {feature.title}
                  </h3>
                  <p className="text-gray-600 dark:text-gray-400 text-sm leading-relaxed">
                    {feature.description}
                  </p>
                </div>
              </Card>
            ))}
          </div>

          <div className="mt-32 border-t border-gray-200 dark:border-gray-800 pt-8 text-center text-sm text-gray-500 dark:text-gray-400">
            <p>&copy; {new Date().getFullYear()} QueryBase. Open source software.</p>
          </div>
        </div>
      </main>
    </PageTransition >
  );
}
