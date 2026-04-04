'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import UserMenu from './UserMenu';
import NotificationBell from './NotificationBell';

export default function Header() {
  const pathname = usePathname();
  const { isAuthenticated } = useAuthStore();

  // Don't show header on portal, login, or register pages
  if (pathname === '/portal' || pathname === '/login' || pathname === '/register') {
    return null;
  }

  return (
    <header className="header-main">
      <div className="header-container">
        <div className="header-left">
          <Link href="/portal" className="header-logo">
            <span className="logo-text">COSMIC</span>
            <span className="logo-accent">VOID</span>
          </Link>

          <nav className="header-nav">
            <Link
              href="/"
              className={`nav-link ${pathname === '/' ? 'active' : ''}`}
            >
              Home
            </Link>
            <Link
              href="/game"
              className={`nav-link ${pathname === '/game' ? 'active' : ''}`}
            >
              Game
            </Link>
            <Link
              href="/leaderboard"
              className={`nav-link ${pathname === '/leaderboard' ? 'active' : ''}`}
            >
              Leaderboard
            </Link>
            <Link
              href="/subscription"
              className={`nav-link ${pathname === '/subscription' ? 'active' : ''}`}
            >
              Subscribe
            </Link>
          </nav>
        </div>

        <div className="header-right">
          {isAuthenticated ? (
            <div className="flex items-center gap-2">
              <NotificationBell />
              <UserMenu />
            </div>
          ) : (
            <div className="auth-buttons flex gap-3">
              <Link href="/login" className="btn-secondary px-4 py-2">
                Sign In
              </Link>
              <Link href="/register" className="btn-primary px-4 py-2">
                Sign Up
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}