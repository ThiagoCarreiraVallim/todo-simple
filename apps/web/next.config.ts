import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Standalone produces a build with minimal node_modules to run in Docker.
  output: 'standalone',
};

export default nextConfig;
