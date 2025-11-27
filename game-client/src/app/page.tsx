import Link from 'next/link';

export default function Home() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center">
      <h1 className="text-4xl font-bold mb-8 text-[#ff00ff]">Void Raiders</h1>
      <Link
        href="/game"
        className="px-8 py-4 bg-[#ff00ff] text-white text-xl font-bold rounded hover:bg-[#cc00cc] transition-colors"
      >
        開始遊戲
      </Link>
    </main>
  );
}
