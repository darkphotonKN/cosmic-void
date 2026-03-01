'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/authStore';

export default function Home() {
  const { isAuthenticated } = useAuthStore();

  return (
    <main className="min-h-screen bg-black text-white flex flex-col">
      <div className="flex-grow flex flex-col justify-center container mx-auto px-12 py-32 max-w-7xl">
        {/* Hero Section */}
        <div className="text-center space-y-12 mb-32">
          <h1 className="text-7xl md:text-8xl font-bold font-orbitron bg-gradient-to-r from-cyan-400 via-purple-500 to-pink-500 bg-clip-text text-transparent pb-4">
            COSMIC VOID
          </h1>
          <p className="text-xl md:text-2xl text-gray-300 max-w-3xl mx-auto leading-loose px-8">
            Enter the ultimate multiplayer space survival experience. Battle, explore, and dominate the void.
          </p>

          <div className="pt-16">
            {!isAuthenticated ? (
              <div className="flex flex-col sm:flex-row gap-8 justify-center">
                <Link
                  href="/register"
                  className="inline-flex items-center justify-center px-12 py-5 text-lg font-bold text-white bg-gradient-to-r from-purple-600 to-pink-600 rounded-xl shadow-2xl shadow-purple-500/25 hover:shadow-purple-500/40 hover:scale-105 transition-all duration-300 uppercase tracking-wider"
                >
                  Start Playing
                </Link>
                <Link
                  href="/login"
                  className="inline-flex items-center justify-center px-12 py-5 text-lg font-bold text-cyan-400 bg-gray-900/80 border-2 border-cyan-400/50 rounded-xl hover:bg-cyan-400/10 hover:border-cyan-400 hover:text-cyan-300 hover:shadow-lg hover:shadow-cyan-400/25 transition-all duration-300 uppercase tracking-wider backdrop-blur-sm"
                >
                  Sign In
                </Link>
              </div>
            ) : (
              <Link
                href="/game"
                className="inline-flex items-center justify-center px-14 py-5 text-lg font-bold text-white bg-gradient-to-r from-purple-600 to-pink-600 rounded-xl shadow-2xl shadow-purple-500/25 hover:shadow-purple-500/40 hover:scale-105 transition-all duration-300 uppercase tracking-wider"
              >
                Launch Game
              </Link>
            )}
          </div>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-3 gap-16 mb-32 px-8">
          <div className="text-center space-y-6">
            <h3 className="text-2xl font-bold text-cyan-400">Real-time Combat</h3>
            <p className="text-gray-400 text-lg leading-loose">
              Engage in intense PvP battles across the galaxy
            </p>
          </div>
          <div className="text-center space-y-6">
            <h3 className="text-2xl font-bold text-purple-400">Collect Treasures</h3>
            <p className="text-gray-400 text-lg leading-loose">
              Discover rare artifacts and powerful upgrades
            </p>
          </div>
          <div className="text-center space-y-6">
            <h3 className="text-2xl font-bold text-pink-400">Climb Rankings</h3>
            <p className="text-gray-400 text-lg leading-loose">
              Compete for the top spots on the leaderboard
            </p>
          </div>
        </div>

        {/* Quick Links */}
        <div className="text-center pt-20">
          <div className="flex gap-10 justify-center flex-wrap">
            <Link
              href="/leaderboard"
              className="px-10 py-4 text-gray-400 hover:text-cyan-400 hover:bg-cyan-400/5 rounded-lg transition-all duration-200 font-medium text-lg"
            >
              View Leaderboard
            </Link>
            <Link
              href="/portal"
              className="px-10 py-4 text-gray-400 hover:text-purple-400 hover:bg-purple-400/5 rounded-lg transition-all duration-200 font-medium text-lg"
            >
              Experience Portal
            </Link>
            <Link
              href="/subscription"
              className="px-10 py-4 text-gray-400 hover:text-pink-400 hover:bg-pink-400/5 rounded-lg transition-all duration-200 font-medium text-lg"
            >
              Premium Access
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
