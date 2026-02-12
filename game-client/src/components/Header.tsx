'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import UserMenu from './UserMenu';

export default function Header() {
  const pathname = usePathname();
  const { isAuthenticated } = useAuthStore();

  // Don't show header on splash, login, or register pages
  if (pathname === '/' || pathname === '/login' || pathname === '/register') {
    return null;
  }

  return (
    <header className="header-main">
      <div className="header-container">
        <div className="header-left">
          <Link href="/" className="header-logo">
            <span className="logo-text">COSMIC</span>
            <span className="logo-accent">VOID</span>
          </Link>

          {isAuthenticated && (
            <nav className="header-nav">
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
          )}
        </div>

        <div className="header-right">
          {isAuthenticated ? (
            <UserMenu />
          ) : (
            <div className="auth-buttons">
              <Link href="/login" className="btn-secondary">
                Login
              </Link>
              <Link href="/register" className="btn-primary">
                Sign Up
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}