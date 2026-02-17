'use client';

import { useEffect, useState } from 'react';

interface AuthGuardProps {
  children: React.ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    // Just a brief delay to ensure client-side hydration
    const timer = setTimeout(() => {
      setIsReady(true);
    }, 10);

    return () => clearTimeout(timer);
  }, []);

  // Brief loading state just for hydration
  if (!isReady) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-black">
        <div className="text-cyan-400">Loading...</div>
      </div>
    );
  }

  return <>{children}</>;
}