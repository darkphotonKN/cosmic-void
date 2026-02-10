'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';

interface AuthGuardProps {
  children: React.ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    const protectedPaths = ['/game', '/profile'];
    const isProtectedPath = protectedPaths.some(path => pathname.startsWith(path));

    if (isProtectedPath && !isAuthenticated) {
      router.push(`/login?redirect=${pathname}`);
    }

    // If authenticated and on login/register, redirect to game
    if (isAuthenticated && (pathname === '/login' || pathname === '/register')) {
      router.push('/game');
    }
  }, [isAuthenticated, pathname, router]);

  return <>{children}</>;
}