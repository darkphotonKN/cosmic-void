'use client';

import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/authStore';

interface AuthGuardProps {
  children: React.ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
  // Start false on both server and client to avoid SSR/CSR markup mismatch
  // and to avoid touching useAuthStore.persist during SSR (it can be undefined
  // there under Zustand v5 + Next.js App Router, throwing "Cannot read
  // properties of undefined (reading 'hasHydrated')").
  const [hasHydrated, setHasHydrated] = useState(false);

  useEffect(() => {
    // Already hydrated by the time the effect runs? Flip immediately.
    if (useAuthStore.persist.hasHydrated()) {
      setHasHydrated(true);
      return;
    }
    const unsub = useAuthStore.persist.onFinishHydration(() => setHasHydrated(true));
    return () => unsub();
  }, []);

  if (!hasHydrated) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-black">
        <div className="text-cyan-400">Loading...</div>
      </div>
    );
  }

  return <>{children}</>;
}
