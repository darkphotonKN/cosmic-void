import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: false, // 關閉以避免 Phaser 場景重複初始化
  // 保留 Webpack 設定以支援 Phaser
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      '@': require('path').resolve(__dirname, 'src'),
    };
    return config;
  },
};

export default nextConfig;
