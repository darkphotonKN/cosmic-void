'use client';

import dynamic from 'next/dynamic';

const PhaserGame = dynamic(() => import('@/components/PhaserGame'), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center min-h-screen">
      <p className="text-[#ff00ff] text-xl">載入遊戲中...</p>
    </div>
  ),
});

export default function GamePage() {
  return (
    <main className="min-h-screen flex items-center justify-center">
      <PhaserGame />
    </main>
  );
}
