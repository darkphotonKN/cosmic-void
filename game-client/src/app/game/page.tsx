'use client';

import dynamic from 'next/dynamic';
import { useAuthStore } from '@/stores/authStore';
import Link from 'next/link';

const PhaserGame = dynamic(() => import('@/components/PhaserGame'), {
  ssr: false,
  loading: () => (
    <div className="flex flex-col items-center justify-center min-h-screen bg-[#0a0a12] gap-4">
      <p className="text-[#00f0ff] text-3xl font-orbitron font-bold tracking-[0.15em] uppercase" style={{ textShadow: '0 0 30px rgba(0,240,255,0.2)' }}>Initializing Uplink</p>
      <p className="text-[#556677] text-sm tracking-[0.2em] uppercase">Preparing deployment zone...</p>
    </div>
  ),
});

export default function GamePage() {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) {
    return (
      <main className="min-h-screen flex items-center justify-center bg-black">
        <div className="text-center space-y-6 p-8 bg-gray-900/50 rounded-lg border border-cyan-500/30 backdrop-blur-sm">
          <h1 className="text-4xl font-bold text-cyan-400 font-orbitron">
            Access Restricted
          </h1>
          <p className="text-gray-300 text-lg max-w-md">
            You need to be signed in to play Cosmic Void. Join the adventure by creating an account or signing in.
          </p>
          <div className="flex gap-4 justify-center pt-4">
            <Link
              href="/login"
              className="px-6 py-3 bg-cyan-500/10 text-cyan-400 border border-cyan-500/50 rounded hover:bg-cyan-500/20 transition-colors"
            >
              Sign In
            </Link>
            <Link
              href="/register"
              className="px-6 py-3 bg-purple-500 text-white rounded hover:bg-purple-600 transition-colors"
            >
              Create Account
            </Link>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-black">
      <div className="flex flex-col items-center justify-center min-h-screen pb-4">
        <PhaserGame />
      </div>
    </main>
  );
}
