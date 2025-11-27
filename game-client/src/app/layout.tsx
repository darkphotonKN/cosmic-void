import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Void Raiders',
  description: 'A Phaser 3 game',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
