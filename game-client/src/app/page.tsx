'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/authStore';

export default function Home() {
  const { isAuthenticated } = useAuthStore();

  return (
    <main className="min-h-screen bg-[#0a0a12] text-white flex flex-col">
      <div className="flex-grow flex flex-col justify-center container mx-auto px-12 py-32 max-w-7xl">
        {/* Hero Section */}
        <div className="text-center space-y-10 mb-32">
          <h1 className="text-7xl md:text-8xl font-bold font-orbitron text-[#00f0ff] pb-4" style={{ textShadow: '0 0 30px rgba(0,240,255,0.3), 0 0 60px rgba(0,240,255,0.1)', letterSpacing: '0.15em' }}>
            COSMIC VOID
          </h1>
          <p className="text-lg md:text-xl text-[#556677] max-w-2xl mx-auto leading-relaxed tracking-wide">
            Drop into the void. Loot derelict stations. Fight to extract.
          </p>

          <div className="pt-14">
            {!isAuthenticated ? (
              <div className="flex flex-col sm:flex-row gap-6 justify-center">
                <Link
                  href="/register"
                  className="inline-flex items-center justify-center px-12 py-4 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md shadow-lg shadow-[#00f0ff]/15 hover:shadow-[#00f0ff]/30 hover:scale-105 transition-all duration-300 uppercase tracking-[0.2em]"
                >
                  Enlist Now
                </Link>
                <Link
                  href="/login"
                  className="inline-flex items-center justify-center px-12 py-4 text-sm font-bold text-[#556677] bg-transparent border border-[#00f0ff]/20 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 hover:text-[#00f0ff] hover:shadow-lg hover:shadow-[#00f0ff]/10 transition-all duration-300 uppercase tracking-[0.2em]"
                >
                  Sign In
                </Link>
              </div>
            ) : (
              <Link
                href="/game"
                className="inline-flex items-center justify-center px-14 py-4 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md shadow-lg shadow-[#00f0ff]/15 hover:shadow-[#00f0ff]/30 hover:scale-105 transition-all duration-300 uppercase tracking-[0.2em]"
              >
                Deploy
              </Link>
            )}
          </div>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-3 gap-16 mb-32 px-8">
          <div className="text-center space-y-4">
            <h3 className="text-xl font-bold text-[#00f0ff] tracking-widest">PVP COMBAT</h3>
            <p className="text-[#445566] text-sm leading-relaxed tracking-wide">
              Engage hostiles across derelict stations in the void
            </p>
          </div>
          <div className="text-center space-y-4">
            <h3 className="text-xl font-bold text-[#ff00aa] tracking-widest">EXTRACTION</h3>
            <p className="text-[#445566] text-sm leading-relaxed tracking-wide">
              Loot high-value gear and extract before the void claims you
            </p>
          </div>
          <div className="text-center space-y-4">
            <h3 className="text-xl font-bold text-[#00f0ff] tracking-widest">RANKINGS</h3>
            <p className="text-[#445566] text-sm leading-relaxed tracking-wide">
              Climb the operator leaderboard and prove your worth
            </p>
          </div>
        </div>

        {/* Quick Links */}
        <div className="text-center pt-16">
          <div className="flex gap-8 justify-center flex-wrap">
            <Link
              href="/leaderboard"
              className="px-8 py-3 text-[#334455] hover:text-[#00f0ff] hover:bg-[#00f0ff]/3 rounded-md transition-all duration-200 text-sm tracking-widest uppercase"
            >
              Leaderboard
            </Link>
            <Link
              href="/portal"
              className="px-8 py-3 text-[#334455] hover:text-[#ff00aa] hover:bg-[#ff00aa]/3 rounded-md transition-all duration-200 text-sm tracking-widest uppercase"
            >
              Portal
            </Link>
            <Link
              href="/subscription"
              className="px-8 py-3 text-[#334455] hover:text-[#00f0ff] hover:bg-[#00f0ff]/3 rounded-md transition-all duration-200 text-sm tracking-widest uppercase"
            >
              Premium Access
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
