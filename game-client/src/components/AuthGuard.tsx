'use client';

import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/authStore';

interface AuthGuardProps {
  children: React.ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
  const [hasHydrated, setHasHydrated] = useState(useAuthStore.persist.hasHydrated());

  useEffect(() => {
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
