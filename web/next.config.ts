import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // output: 'export' is only for production static builds
  // Disabled for E2E testing with dev server
  output: process.env.NODE_ENV === 'production' ? 'export' : undefined,
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },
  // rewrites are not supported in static export
  // async rewrites() {
  //   return [
  //     {
  //       source: '/api/v1/:path*',
  //       destination: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/:path*`,
  //     },
  //   ];
  // },
};

export default nextConfig;
