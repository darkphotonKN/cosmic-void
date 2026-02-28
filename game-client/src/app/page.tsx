'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/authStore';

export default function Home() {
  const { isAuthenticated } = useAuthStore();

  return (
    <main className="min-h-screen bg-black text-white">
      <div className="container mx-auto px-4 py-16 max-w-6xl">
        <section className="text-center mb-16">
          <h1 className="text-6xl font-bold mb-6 font-orbitron bg-gradient-to-r from-cyan-400 to-purple-500 bg-clip-text text-transparent">
            COSMIC VOID
          </h1>
          <p className="text-xl text-gray-300 mb-8 max-w-3xl mx-auto">
            Navigate through the infinite darkness of space. Collect treasures, battle enemies,
            and survive the cosmic void in this thrilling multiplayer adventure.
          </p>
          {!isAuthenticated ? (
            <div className="flex gap-4 justify-center">
              <Link
                href="/register"
                className="px-8 py-4 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors text-lg font-semibold"
              >
                Start Playing
              </Link>
              <Link
                href="/portal"
                className="px-8 py-4 bg-gray-800 text-cyan-400 border border-cyan-500/50 rounded-lg hover:bg-gray-700 transition-colors text-lg font-semibold"
              >
                View Portal
              </Link>
            </div>
          ) : (
            <Link
              href="/game"
              className="inline-block px-8 py-4 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors text-lg font-semibold"
            >
              Play Now
            </Link>
          )}
        </section>

        <section className="grid md:grid-cols-3 gap-8 mb-16">
          <div className="bg-gray-900/50 p-6 rounded-lg border border-gray-800">
            <h3 className="text-2xl font-bold mb-3 text-cyan-400">Explore</h3>
            <p className="text-gray-400">
              Discover mysterious artifacts and hidden treasures scattered across the cosmic void.
            </p>
          </div>
          <div className="bg-gray-900/50 p-6 rounded-lg border border-gray-800">
            <h3 className="text-2xl font-bold mb-3 text-purple-400">Battle</h3>
            <p className="text-gray-400">
              Engage in intense multiplayer combat with players from around the galaxy.
            </p>
          </div>
          <div className="bg-gray-900/50 p-6 rounded-lg border border-gray-800">
            <h3 className="text-2xl font-bold mb-3 text-pink-400">Survive</h3>
            <p className="text-gray-400">
              Master your skills and strategy to outlast your opponents in the void.
            </p>
          </div>
        </section>

        <section className="text-center bg-gradient-to-r from-purple-900/20 to-cyan-900/20 p-8 rounded-lg border border-purple-500/20">
          <h2 className="text-3xl font-bold mb-4">Ready to Enter the Void?</h2>
          <p className="text-gray-300 mb-6">
            Join thousands of players in the ultimate space survival experience.
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              href="/leaderboard"
              className="px-6 py-3 bg-gray-800 text-cyan-400 border border-cyan-500/50 rounded hover:bg-gray-700 transition-colors"
            >
              View Leaderboard
            </Link>
            <Link
              href="/subscription"
              className="px-6 py-3 bg-purple-600 text-white rounded hover:bg-purple-700 transition-colors"
            >
              Premium Access
            </Link>
          </div>
        </section>
      </div>
    </main>
  );
}
