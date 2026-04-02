'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/authStore';

export default function Home() {
  const { isAuthenticated } = useAuthStore();

  return (
    <main className="min-h-screen bg-[#0a0a12] text-white overflow-x-hidden">

      {/* ── Hero ── */}
      <section className="relative min-h-[90vh] flex flex-col items-center justify-center text-center px-8 sm:px-12 lg:px-20">
        {/* Subtle radial glow */}
        <div className="absolute inset-0 pointer-events-none" style={{
          background: 'radial-gradient(ellipse 60% 40% at 50% 40%, rgba(0,240,255,0.06) 0%, transparent 70%)',
        }} />

        <p className="relative text-xs text-[#ff00aa] tracking-[0.4em] uppercase mb-8">
          Extraction // Survive the Void
        </p>
        <h1
          className="relative text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-bold font-orbitron text-[#00f0ff] mb-10"
          style={{ textShadow: '0 0 40px rgba(0,240,255,0.25), 0 0 80px rgba(0,240,255,0.08)', letterSpacing: '0.12em' }}
        >
          COSMIC VOID
        </h1>
        <p className="relative text-base sm:text-lg text-[#556677] max-w-lg leading-relaxed tracking-wide mb-16">
          Drop into derelict stations adrift in the void. Scavenge weapons, armor, and consumables. Eliminate hostiles or find the escape route. Only the extracted survive.
        </p>

        <div className="relative">
          {!isAuthenticated ? (
            <div className="flex flex-col sm:flex-row gap-5 justify-center">
              <Link
                href="/register"
                className="px-10 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] hover:scale-[1.03] transition-all duration-300 uppercase tracking-[0.2em]"
              >
                Enlist Now
              </Link>
              <Link
                href="/login"
                className="px-10 py-3.5 text-sm font-bold text-[#556677] border border-[#00f0ff]/20 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 hover:text-[#00f0ff] transition-all duration-300 uppercase tracking-[0.2em]"
              >
                Sign In
              </Link>
            </div>
          ) : (
            <Link
              href="/game"
              className="px-12 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] hover:scale-[1.03] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              Deploy
            </Link>
          )}
        </div>

        {/* Scroll hint */}
        <div className="absolute bottom-12 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2 opacity-25">
          <span className="text-[10px] tracking-[0.3em] uppercase text-[#445566]">Scroll</span>
          <div className="w-px h-8 bg-gradient-to-b from-[#445566] to-transparent" />
        </div>
      </section>

      {/* ── Divider line ── */}
      <div className="max-w-3xl mx-auto px-8 sm:px-12">
        <div className="h-px bg-gradient-to-r from-transparent via-[#00f0ff]/15 to-transparent" />
      </div>

      {/* ── Core Mechanics ── */}
      <section className="max-w-5xl mx-auto px-8 sm:px-12 lg:px-16 py-32">
        <p className="text-xs text-[#ff00aa] tracking-[0.3em] uppercase mb-5 text-center">Systems</p>
        <h2
          className="text-2xl sm:text-3xl font-bold font-orbitron text-center text-[#00f0ff] mb-20 tracking-[0.1em]"
          style={{ textShadow: '0 0 20px rgba(0,240,255,0.15)' }}
        >
          CORE MECHANICS
        </h2>

        <div className="grid md:grid-cols-3 gap-8">
          <MechanicCard
            title="REAL-TIME COMBAT"
            description="Server-authoritative hit detection. Attack, take damage, and eliminate opponents — all synced at 30 ticks per second."
            color="cyan"
          />
          <MechanicCard
            title="LOOT & EXTRACTION"
            description="Breach containers for randomized weapons, armor, and consumables. Activate switches, unlock the escape door, and extract with your haul."
            color="magenta"
          />
          <MechanicCard
            title="RANKED MATCHES"
            description="Compete in elimination matches. Every kill, death, and extraction is tracked. Climb the operator leaderboard."
            color="cyan"
          />
        </div>
      </section>

      {/* ── Divider ── */}
      <div className="max-w-3xl mx-auto px-8 sm:px-12">
        <div className="h-px bg-gradient-to-r from-transparent via-[#ff00aa]/10 to-transparent" />
      </div>

      {/* ── Gear & Items ── */}
      <section className="max-w-5xl mx-auto px-8 sm:px-12 lg:px-16 py-32">
        <p className="text-xs text-[#ff00aa] tracking-[0.3em] uppercase mb-5 text-center">Arsenal</p>
        <h2
          className="text-2xl sm:text-3xl font-bold font-orbitron text-center text-[#00f0ff] mb-8 tracking-[0.1em]"
          style={{ textShadow: '0 0 20px rgba(0,240,255,0.15)' }}
        >
          GEAR SYSTEM
        </h2>
        <p className="text-sm text-[#445566] text-center max-w-md mx-auto mb-20 tracking-wide leading-relaxed">
          Every container spawns randomized loot from the item pool — 40% weapons, 35% armor, 25% consumables. Four rarity tiers from common to legendary.
        </p>

        <div className="grid sm:grid-cols-3 gap-8">
          <GearCard
            title="WEAPONS"
            items={['Attack Power scaling', 'Critical Rate (0–100%)', 'Types: Sword, Axe, Bow']}
            color="cyan"
          />
          <GearCard
            title="ARMOR"
            items={['Defense Rating', 'Magic Resistance', 'Slots: Head, Chest, Legs, Gloves, Shield']}
            color="magenta"
          />
          <GearCard
            title="CONSUMABLES"
            items={['Health restoration', 'Mana restoration', 'Timed buffs with duration']}
            color="cyan"
          />
        </div>

        <div className="flex justify-center gap-5 mt-16 flex-wrap">
          {(['Common', 'Rare', 'Epic', 'Legendary'] as const).map((rarity) => (
            <RarityBadge key={rarity} rarity={rarity} />
          ))}
        </div>
      </section>

      {/* ── Divider ── */}
      <div className="max-w-3xl mx-auto px-8 sm:px-12">
        <div className="h-px bg-gradient-to-r from-transparent via-[#00f0ff]/15 to-transparent" />
      </div>

      {/* ── How a Match Works ── */}
      <section className="max-w-4xl mx-auto px-8 sm:px-12 lg:px-16 py-32">
        <p className="text-xs text-[#ff00aa] tracking-[0.3em] uppercase mb-5 text-center">Match Flow</p>
        <h2
          className="text-2xl sm:text-3xl font-bold font-orbitron text-center text-[#00f0ff] mb-20 tracking-[0.1em]"
          style={{ textShadow: '0 0 20px rgba(0,240,255,0.15)' }}
        >
          HOW IT WORKS
        </h2>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <StepCard step="01" title="QUEUE" description="Find a match. Operators are paired and deployed into a shared station." />
          <StepCard step="02" title="LOOT" description="Open containers for randomized gear. Equip weapons and armor to gain an edge." />
          <StepCard step="03" title="FIGHT" description="Engage other operators. Every elimination is tracked — kills, deaths, positioning." />
          <StepCard step="04" title="EXTRACT" description="Activate switches to unlock the escape door. Extract to keep your stats and rank up." />
        </div>
      </section>

      {/* ── Divider ── */}
      <div className="max-w-3xl mx-auto px-8 sm:px-12">
        <div className="h-px bg-gradient-to-r from-transparent via-[#ff00aa]/10 to-transparent" />
      </div>

      {/* ── Controls ── */}
      <section className="max-w-3xl mx-auto px-8 sm:px-12 lg:px-16 py-32">
        <p className="text-xs text-[#ff00aa] tracking-[0.3em] uppercase mb-5 text-center">Controls</p>
        <h2
          className="text-2xl sm:text-3xl font-bold font-orbitron text-center text-[#00f0ff] mb-20 tracking-[0.1em]"
          style={{ textShadow: '0 0 20px rgba(0,240,255,0.15)' }}
        >
          OPERATOR CONTROLS
        </h2>

        <div className="grid grid-cols-2 sm:grid-cols-3 gap-5 max-w-lg mx-auto">
          <ControlKey keys="WASD" label="Move" />
          <ControlKey keys="SPACE" label="Attack" />
          <ControlKey keys="E" label="Interact" />
          <ControlKey keys="F" label="Loot Item" />
          <ControlKey keys="I" label="Inventory" />
          <ControlKey keys="ESC" label="Close Menu" />
        </div>
      </section>

      {/* ── Divider ── */}
      <div className="max-w-3xl mx-auto px-8 sm:px-12">
        <div className="h-px bg-gradient-to-r from-transparent via-[#00f0ff]/15 to-transparent" />
      </div>

      {/* ── CTA ── */}
      <section className="max-w-3xl mx-auto px-8 sm:px-12 lg:px-16 py-32 text-center">
        <h2
          className="text-2xl sm:text-3xl font-bold font-orbitron text-[#00f0ff] mb-8 tracking-[0.1em]"
          style={{ textShadow: '0 0 20px rgba(0,240,255,0.15)' }}
        >
          ENTER THE VOID
        </h2>
        <p className="text-sm text-[#445566] max-w-sm mx-auto mb-14 tracking-wide leading-relaxed">
          Register as an operator. Deploy into your first match. Nothing survives the void forever.
        </p>

        {!isAuthenticated ? (
          <div className="flex flex-col sm:flex-row gap-5 justify-center">
            <Link
              href="/register"
              className="px-10 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] hover:scale-[1.03] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              Enlist Now
            </Link>
            <Link
              href="/leaderboard"
              className="px-10 py-3.5 text-sm font-bold text-[#556677] border border-[#00f0ff]/15 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/30 hover:text-[#00f0ff] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              View Rankings
            </Link>
          </div>
        ) : (
          <div className="flex flex-col sm:flex-row gap-5 justify-center">
            <Link
              href="/game"
              className="px-12 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] hover:scale-[1.03] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              Deploy
            </Link>
            <Link
              href="/leaderboard"
              className="px-10 py-3.5 text-sm font-bold text-[#556677] border border-[#00f0ff]/15 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/30 hover:text-[#00f0ff] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              View Rankings
            </Link>
          </div>
        )}
      </section>

      {/* ── Footer ── */}
      <footer className="border-t border-[#00f0ff]/8 py-12 px-8 sm:px-12 mt-8">
        <div className="max-w-4xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-6">
          <span className="text-[11px] text-[#1a1a28] tracking-[0.2em] uppercase">
            v0.1 // Sector 7-G // Cosmic Void
          </span>
          <div className="flex gap-8">
            <Link href="/leaderboard" className="text-[11px] text-[#334455] hover:text-[#00f0ff] tracking-[0.15em] uppercase transition-colors">
              Rankings
            </Link>
            <Link href="/subscription" className="text-[11px] text-[#334455] hover:text-[#00f0ff] tracking-[0.15em] uppercase transition-colors">
              Premium
            </Link>
            <Link href="/profile" className="text-[11px] text-[#334455] hover:text-[#00f0ff] tracking-[0.15em] uppercase transition-colors">
              Profile
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}

/* ── Sub-components ── */

function MechanicCard({ title, description, color }: { title: string; description: string; color: 'cyan' | 'magenta' }) {
  const borderColor = color === 'cyan' ? 'rgba(0,240,255,0.1)' : 'rgba(255,0,170,0.1)';
  const titleColor = color === 'cyan' ? '#00f0ff' : '#ff00aa';
  const hoverBorder = color === 'cyan' ? 'hover:border-[#00f0ff]/20' : 'hover:border-[#ff00aa]/20';

  return (
    <div
      className={`p-8 rounded-lg border bg-[#0a0a12] transition-all duration-300 ${hoverBorder}`}
      style={{ borderColor }}
    >
      <h3 className="text-sm font-bold tracking-[0.15em] mb-4" style={{ color: titleColor }}>
        {title}
      </h3>
      <p className="text-[13px] text-[#556677] leading-[1.8]">
        {description}
      </p>
    </div>
  );
}

function GearCard({ title, items, color }: { title: string; items: string[]; color: 'cyan' | 'magenta' }) {
  const borderColor = color === 'cyan' ? 'rgba(0,240,255,0.1)' : 'rgba(255,0,170,0.1)';
  const titleColor = color === 'cyan' ? '#00f0ff' : '#ff00aa';
  const dotColor = color === 'cyan' ? 'bg-[#00f0ff]/40' : 'bg-[#ff00aa]/40';

  return (
    <div
      className="p-8 rounded-lg border bg-[#0a0a12] transition-all duration-300"
      style={{ borderColor }}
    >
      <h3 className="text-sm font-bold tracking-[0.15em] mb-6" style={{ color: titleColor }}>
        {title}
      </h3>
      <ul className="space-y-3.5">
        {items.map((item) => (
          <li key={item} className="flex items-start gap-3 text-[13px] text-[#556677] leading-relaxed">
            <span className={`w-1 h-1 rounded-full mt-2 flex-shrink-0 ${dotColor}`} />
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}

function StepCard({ step, title, description }: { step: string; title: string; description: string }) {
  return (
    <div className="p-7 rounded-lg border border-[#00f0ff]/8 bg-[#0a0a12]">
      <span className="text-[10px] text-[#334455] tracking-[0.25em] font-bold">{step}</span>
      <h3 className="text-sm font-bold text-[#00f0ff] tracking-[0.12em] mt-3 mb-3">{title}</h3>
      <p className="text-[13px] text-[#445566] leading-[1.8]">{description}</p>
    </div>
  );
}

function ControlKey({ keys, label }: { keys: string; label: string }) {
  return (
    <div className="flex items-center gap-4 p-4 rounded-md border border-[#00f0ff]/8 bg-[#0a0a12]">
      <kbd className="text-xs font-bold text-[#00f0ff] tracking-[0.1em] font-mono bg-[#00f0ff]/5 px-2.5 py-1.5 rounded border border-[#00f0ff]/15 min-w-[48px] text-center">
        {keys}
      </kbd>
      <span className="text-xs text-[#556677] tracking-wide">{label}</span>
    </div>
  );
}

const rarityColors = {
  Common: { text: '#556677', border: 'rgba(85,102,119,0.2)', bg: 'rgba(85,102,119,0.05)' },
  Rare: { text: '#00f0ff', border: 'rgba(0,240,255,0.2)', bg: 'rgba(0,240,255,0.05)' },
  Epic: { text: '#bf5fff', border: 'rgba(191,95,255,0.2)', bg: 'rgba(191,95,255,0.05)' },
  Legendary: { text: '#ff00aa', border: 'rgba(255,0,170,0.2)', bg: 'rgba(255,0,170,0.05)' },
} as const;

function RarityBadge({ rarity }: { rarity: keyof typeof rarityColors }) {
  const c = rarityColors[rarity];
  return (
    <span
      className="text-[11px] font-bold tracking-[0.15em] uppercase px-4 py-2 rounded"
      style={{ color: c.text, border: `1px solid ${c.border}`, background: c.bg }}
    >
      {rarity}
    </span>
  );
}
