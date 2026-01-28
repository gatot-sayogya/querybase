export default function Home() {
  return (
    <main className="min-h-screen p-24">
      <div className="max-w-5xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">
          Welcome to QueryBase
        </h1>
        <p className="text-lg text-gray-600 dark:text-gray-400 mb-4">
          Database explorer with approval workflow
        </p>
        <div className="grid gap-4 mt-8">
          <a
            href="/login"
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-center"
          >
            Login to Get Started
          </a>
        </div>
      </div>
    </main>
  );
}
