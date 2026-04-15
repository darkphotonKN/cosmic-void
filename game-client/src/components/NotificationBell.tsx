'use client';

import { useState, useEffect, useRef } from 'react';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/stores/authStore';

interface Notification {
  id: string;
  user_id: string;
  title: string;
  message: string;
  notification_type: 'game' | 'achievement' | 'system' | 'friend';
  event_type: string;
  read: boolean;
  data: Record<string, any>;
  created_at: string | { seconds: number; nanos: number };
  updated_at: string | { seconds: number; nanos: number };
}

// SVG Bell Icon
const BellIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
  </svg>
);

export default function NotificationBell() {
  const [isOpen, setIsOpen] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { isAuthenticated } = useAuthStore();

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Fetch notifications from API
  const fetchNotifications = async () => {
    if (!isAuthenticated) return;

    setIsLoading(true);
    try {
      const response = await apiClient.getNotifications();
      if (response.notifications) {
        setNotifications(response.notifications);
        // Calculate unread count from notifications
        const unread = response.notifications.filter((n: Notification) => !n.read).length;
        setUnreadCount(unread);
      }
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Fetch notifications on mount and when authenticated
  useEffect(() => {
    fetchNotifications();

    // Optional: Set up polling to refresh notifications every 30 seconds
    const interval = setInterval(fetchNotifications, 30000);
    return () => clearInterval(interval);
  }, [isAuthenticated]);

  const handleMarkAsRead = async (id: string) => {
    try {
      await apiClient.markNotificationAsRead(id);
      setNotifications(prev =>
        prev.map(n => (n.id === id ? { ...n, read: true } : n))
      );
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (error) {
      console.error('Failed to mark notification as read:', error);
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await apiClient.markAllNotificationsAsRead();
      setNotifications(prev => prev.map(n => ({ ...n, read: true })));
      setUnreadCount(0);
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error);
    }
  };

  const getNotificationIcon = (type: Notification['notification_type']) => {
    switch (type) {
      case 'game':
        return '🎮';
      case 'achievement':
        return '🏆';
      case 'system':
        return '⚙️';
      case 'friend':
        return '👥';
      default:
        return '📬';
    }
  };

  const getTimeAgo = (timestamp: string | { seconds: number; nanos: number }) => {
    const now = new Date();
    // Handle protobuf Timestamp format {seconds, nanos}
    let past: Date;
    if (typeof timestamp === 'object' && timestamp?.seconds != null) {
      past = new Date(Number(timestamp.seconds) * 1000);
    } else {
      past = new Date(timestamp as string);
    }

    // Guard against invalid dates
    if (isNaN(past.getTime())) return '';

    const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000);

    if (diffInSeconds < 0) return 'Just now';
    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;

    const days = Math.floor(diffInSeconds / 86400);
    const hours = Math.floor((diffInSeconds % 86400) / 3600);
    const minutes = Math.floor((diffInSeconds % 3600) / 60);
    const seconds = diffInSeconds % 60;

    const parts: string[] = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (seconds > 0 && days === 0) parts.push(`${seconds}s`); // skip seconds if days shown

    return parts.join(' ') + ' ago';
  };

  return (
    <div className="notification-bell" ref={dropdownRef}>
      <button
        className="notification-bell-trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-label="Notifications"
      >
        <BellIcon className="bell-icon" />
        {unreadCount > 0 && (
          <span className="notification-count">{unreadCount > 99 ? '99+' : unreadCount}</span>
        )}
      </button>

      {isOpen && (
        <div className="notification-dropdown">
          <div className="notification-dropdown-header">
            <h3>Notifications</h3>
            {unreadCount > 0 && (
              <button className="mark-read-btn" onClick={handleMarkAllAsRead}>
                Mark all read
              </button>
            )}
          </div>

          <div className="notification-list">
            {notifications.length === 0 ? (
              <div className="notification-empty">
                <div style={{ fontSize: '2rem', opacity: 0.3, marginBottom: '0.5rem' }}>🔔</div>
                <p style={{ margin: 0, color: '#556677' }}>No notifications</p>
              </div>
            ) : (
              notifications.map((notification) => (
                <div
                  key={notification.id}
                  className={`notification-item ${notification.read ? '' : 'unread'}`}
                  onClick={() => !notification.read && handleMarkAsRead(notification.id)}
                >
                  <div className="notification-icon">
                    {getNotificationIcon(notification.notification_type)}
                  </div>
                  <div className="notification-content">
                    <div className="notification-header">
                      <span className="notification-title">{notification.title}</span>
                      <span className="notification-time">{getTimeAgo(notification.created_at)}</span>
                    </div>
                    <p className="notification-message">{notification.message}</p>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      <style jsx>{`
        .notification-bell {
          position: relative;
        }

        .notification-bell-trigger {
          position: relative;
          display: flex;
          align-items: center;
          justify-content: center;
          width: 36px;
          height: 36px;
          background: transparent;
          border: 1px solid rgba(0, 240, 255, 0.2);
          border-radius: 6px;
          color: rgba(0, 240, 255, 0.7);
          cursor: pointer;
          transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .notification-bell-trigger:hover {
          background: rgba(0, 240, 255, 0.05);
          border-color: var(--color-primary);
          color: var(--color-primary);
          box-shadow:
            0 0 20px rgba(0, 240, 255, 0.2),
            inset 0 0 20px rgba(0, 240, 255, 0.05);
          transform: translateY(-1px);
        }

        .bell-icon {
          width: 18px;
          height: 18px;
          filter: drop-shadow(0 0 4px rgba(0, 240, 255, 0.3));
        }

        .notification-count {
          position: absolute;
          top: -6px;
          right: -6px;
          min-width: 16px;
          height: 16px;
          padding: 0 4px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: linear-gradient(135deg, var(--color-primary) 0%, #00d4e6 100%);
          color: #000;
          border-radius: 8px;
          font-size: 9px;
          font-weight: 800;
          letter-spacing: 0.03em;
          border: 1.5px solid var(--color-bg-dark);
          box-shadow:
            0 0 12px rgba(0, 240, 255, 0.6),
            0 2px 8px rgba(0, 0, 0, 0.4);
          animation: neonPulse 2s ease-in-out infinite;
        }

        @keyframes neonPulse {
          0%, 100% {
            box-shadow:
              0 0 12px rgba(0, 240, 255, 0.6),
              0 2px 8px rgba(0, 0, 0, 0.4);
          }
          50% {
            box-shadow:
              0 0 20px rgba(0, 240, 255, 0.9),
              0 2px 12px rgba(0, 0, 0, 0.6);
          }
        }

        .notification-dropdown {
          position: absolute;
          top: calc(100% + 0.75rem);
          right: 0;
          width: 340px;
          max-height: 480px;
          background: linear-gradient(145deg, rgba(10, 10, 20, 0.97) 0%, rgba(6, 6, 14, 0.97) 100%);
          backdrop-filter: blur(24px) saturate(180%);
          border: 1px solid rgba(0, 240, 255, 0.2);
          border-radius: 8px;
          box-shadow:
            0 0 0 1px rgba(0, 240, 255, 0.1),
            0 8px 32px rgba(0, 0, 0, 0.8),
            0 0 60px rgba(0, 240, 255, 0.08);
          animation: dropdownSlide 0.25s cubic-bezier(0.4, 0, 0.2, 1);
          overflow: hidden;
        }

        @keyframes dropdownSlide {
          from {
            opacity: 0;
            transform: translateY(-12px) scale(0.96);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        .notification-dropdown::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          height: 1px;
          background: linear-gradient(90deg,
            transparent 0%,
            rgba(0, 240, 255, 0.4) 50%,
            transparent 100%);
        }

        .notification-dropdown-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 0.875rem 1.125rem;
          border-bottom: 1px solid rgba(0, 240, 255, 0.08);
          background: rgba(0, 240, 255, 0.02);
        }

        .notification-dropdown-header h3 {
          font-family: var(--font-heading);
          font-size: 0.75rem;
          font-weight: 700;
          color: var(--color-primary);
          letter-spacing: 0.15em;
          text-transform: uppercase;
          margin: 0;
          text-shadow: 0 0 8px rgba(0, 240, 255, 0.3);
        }

        .mark-read-btn {
          background: transparent;
          border: 1px solid rgba(0, 240, 255, 0.15);
          color: rgba(0, 240, 255, 0.6);
          font-size: 0.65rem;
          font-weight: 600;
          cursor: pointer;
          padding: 0.25rem 0.625rem;
          border-radius: 4px;
          transition: all 0.2s ease;
          letter-spacing: 0.05em;
          text-transform: uppercase;
        }

        .mark-read-btn:hover {
          background: rgba(0, 240, 255, 0.08);
          border-color: var(--color-primary);
          color: var(--color-primary);
          box-shadow: 0 0 12px rgba(0, 240, 255, 0.15);
        }

        .notification-list {
          max-height: 400px;
          overflow-y: auto;
        }

        .notification-list::-webkit-scrollbar {
          width: 3px;
        }

        .notification-list::-webkit-scrollbar-track {
          background: rgba(0, 0, 0, 0.2);
        }

        .notification-list::-webkit-scrollbar-thumb {
          background: linear-gradient(180deg,
            rgba(0, 240, 255, 0.3) 0%,
            rgba(0, 240, 255, 0.15) 100%);
          border-radius: 2px;
        }

        .notification-list::-webkit-scrollbar-thumb:hover {
          background: linear-gradient(180deg,
            rgba(0, 240, 255, 0.5) 0%,
            rgba(0, 240, 255, 0.25) 100%);
        }

        .notification-empty {
          padding: 3rem 1.5rem;
          text-align: center;
        }

        .notification-item {
          display: flex;
          gap: 0.75rem;
          padding: 0.875rem 1.125rem;
          border-bottom: 1px solid rgba(0, 240, 255, 0.04);
          cursor: pointer;
          transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
          position: relative;
          background: transparent;
        }

        .notification-item:last-child {
          border-bottom: none;
        }

        .notification-item.unread {
          background: linear-gradient(90deg,
            rgba(0, 240, 255, 0.04) 0%,
            rgba(0, 240, 255, 0.02) 100%);
        }

        .notification-item.unread::before {
          content: '';
          position: absolute;
          left: 0;
          top: 0;
          bottom: 0;
          width: 2px;
          background: linear-gradient(180deg,
            var(--color-primary) 0%,
            rgba(0, 240, 255, 0.4) 100%);
          box-shadow: 0 0 8px rgba(0, 240, 255, 0.6);
        }

        .notification-item:hover {
          background: linear-gradient(90deg,
            rgba(0, 240, 255, 0.08) 0%,
            rgba(0, 240, 255, 0.04) 100%);
          border-color: rgba(0, 240, 255, 0.1);
        }

        .notification-icon {
          font-size: 1.125rem;
          flex-shrink: 0;
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(0, 240, 255, 0.06);
          border-radius: 6px;
          border: 1px solid rgba(0, 240, 255, 0.15);
          filter: drop-shadow(0 0 6px rgba(0, 240, 255, 0.2));
        }

        .notification-content {
          flex: 1;
          min-width: 0;
        }

        .notification-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          gap: 0.5rem;
          margin-bottom: 0.25rem;
        }

        .notification-title {
          font-size: 0.8rem;
          font-weight: 600;
          color: rgba(255, 255, 255, 0.9);
          letter-spacing: 0.01em;
          line-height: 1.3;
        }

        .notification-time {
          font-size: 0.65rem;
          color: rgba(0, 240, 255, 0.5);
          white-space: nowrap;
          flex-shrink: 0;
          letter-spacing: 0.05em;
          text-transform: uppercase;
          font-weight: 500;
        }

        .notification-message {
          font-size: 0.75rem;
          color: rgba(255, 255, 255, 0.5);
          margin: 0;
          line-height: 1.5;
          overflow: hidden;
          text-overflow: ellipsis;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
        }
      `}</style>
    </div>
  );
}
