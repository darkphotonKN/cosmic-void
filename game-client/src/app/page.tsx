'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/authStore';

export default function Home() {
  const { isAuthenticated } = useAuthStore();

  return (
    <main className="min-h-screen bg-[#0a0a12] text-white overflow-x-hidden">

      {/* ── Hero ── */}
      <section className="relative min-h-screen flex flex-col items-center justify-center text-center px-6 sm:px-12 pt-16">
        {/* Background glow layers */}
        <div className="absolute inset-0 pointer-events-none" style={{
          background: 'radial-gradient(ellipse 70% 50% at 50% 35%, rgba(0,240,255,0.05) 0%, transparent 60%)',
        }} />
        <div className="absolute inset-0 pointer-events-none" style={{
          background: 'radial-gradient(ellipse 40% 30% at 50% 70%, rgba(255,0,170,0.02) 0%, transparent 60%)',
        }} />

        <p className="relative text-[11px] text-[#ff00aa] tracking-[0.4em] uppercase mb-6 font-medium">
          Extraction // Survive the Void
        </p>
        <h1
          className="relative text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-bold font-orbitron text-[#00f0ff] mb-6"
          style={{ textShadow: '0 0 60px rgba(0,240,255,0.2), 0 0 120px rgba(0,240,255,0.05)', letterSpacing: '0.12em' }}
        >
          COSMIC VOID
        </h1>
        <p className="relative text-lg sm:text-xl text-[#667788] max-w-xl leading-relaxed tracking-wide mb-12">
          Drop into derelict stations adrift in the void. Scavenge weapons, armor, and consumables. Eliminate hostiles or find the escape route.
        </p>
        <p className="relative text-sm text-[#445566] mb-12 tracking-wider">
          Only the extracted survive.
        </p>

        <div className="relative">
          {!isAuthenticated ? (
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link
                href="/register"
                className="px-10 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.35)] hover:scale-[1.02] transition-all duration-300 uppercase tracking-[0.2em]"
              >
                Enlist Now
              </Link>
              <Link
                href="/login"
                className="px-10 py-3.5 text-sm font-bold text-[#667788] border border-[#00f0ff]/20 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 hover:text-[#00f0ff] transition-all duration-300 uppercase tracking-[0.2em]"
              >
                Sign In
              </Link>
            </div>
          ) : (
            <Link
              href="/game"
              className="px-12 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.35)] hover:scale-[1.02] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              Deploy
            </Link>
          )}
        </div>

        {/* Scroll indicator */}
        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2 opacity-30 animate-pulse">
          <span className="text-[10px] tracking-[0.3em] uppercase text-[#556677]">Scroll</span>
          <div className="w-px h-8 bg-gradient-to-b from-[#556677] to-transparent" />
        </div>
      </section>

      {/* ── What is Cosmic Void ── */}
      <section className="max-w-4xl mx-auto px-6 sm:px-12 py-28">
        <SectionLabel>About</SectionLabel>
        <SectionTitle>What is Cosmic Void?</SectionTitle>
        <p className="text-center text-[#667788] max-w-2xl mx-auto leading-[1.9] text-base tracking-wide">
          Cosmic Void is a real-time multiplayer extraction game. You deploy as an operator into
          hostile stations, scavenge randomized loot, fight other players, and race to extract
          before the void consumes everything. Every match is different. Every decision matters.
          Server-authoritative, skill-based, and unforgiving.
        </p>
      </section>

      <Divider />

      {/* ── Core Mechanics ── */}
      <section className="max-w-5xl mx-auto px-6 sm:px-12 lg:px-16 py-28">
        <SectionLabel>Systems</SectionLabel>
        <SectionTitle>Core Mechanics</SectionTitle>

        <div className="grid md:grid-cols-3 gap-6">
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

      <Divider variant="magenta" />

      {/* ── Gear & Items ── */}
      <section className="max-w-5xl mx-auto px-6 sm:px-12 lg:px-16 py-28">
        <SectionLabel>Arsenal</SectionLabel>
        <SectionTitle>Gear System</SectionTitle>
        <p className="text-sm text-[#556677] text-center max-w-lg mx-auto mb-16 tracking-wide leading-relaxed">
          Every container spawns randomized loot from the item pool — 40% weapons, 35% armor, 25% consumables. Four rarity tiers from common to legendary.
        </p>

        <div className="grid sm:grid-cols-3 gap-6">
          <GearCard
            title="WEAPONS"
            items={['Attack Power scaling', 'Critical Rate (0-100%)', 'Types: Sword, Axe, Bow']}
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

        <div className="flex justify-center gap-4 mt-14 flex-wrap">
          {(['Common', 'Rare', 'Epic', 'Legendary'] as const).map((rarity) => (
            <RarityBadge key={rarity} rarity={rarity} />
          ))}
        </div>
      </section>

      <Divider />

      {/* ── How a Match Works ── */}
      <section className="max-w-5xl mx-auto px-6 sm:px-12 lg:px-16 py-28">
        <SectionLabel>Match Flow</SectionLabel>
        <SectionTitle>How It Works</SectionTitle>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-5">
          <StepCard step="01" title="QUEUE" description="Find a match. Operators are paired and deployed into a shared station." />
          <StepCard step="02" title="LOOT" description="Open containers for randomized gear. Equip weapons and armor to gain an edge." />
          <StepCard step="03" title="FIGHT" description="Engage other operators. Every elimination is tracked — kills, deaths, positioning." />
          <StepCard step="04" title="EXTRACT" description="Activate switches to unlock the escape door. Extract to keep your stats and rank up." />
        </div>
      </section>

      <Divider variant="magenta" />

      {/* ── Controls ── */}
      <section className="max-w-3xl mx-auto px-6 sm:px-12 lg:px-16 py-28">
        <SectionLabel>Controls</SectionLabel>
        <SectionTitle>Operator Controls</SectionTitle>

        <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 max-w-md mx-auto">
          <ControlKey keys="WASD" label="Move" />
          <ControlKey keys="SPACE" label="Attack" />
          <ControlKey keys="E" label="Interact" />
          <ControlKey keys="F" label="Loot Item" />
          <ControlKey keys="I" label="Inventory" />
          <ControlKey keys="ESC" label="Close Menu" />
        </div>
      </section>

      <Divider />

      {/* ── CTA ── */}
      <section className="relative max-w-3xl mx-auto px-6 sm:px-12 lg:px-16 py-32 text-center">
        <div className="absolute inset-0 pointer-events-none" style={{
          background: 'radial-gradient(ellipse 60% 50% at 50% 50%, rgba(0,240,255,0.03) 0%, transparent 70%)',
        }} />

        <h2
          className="relative text-3xl sm:text-4xl font-bold font-orbitron text-[#00f0ff] mb-6 tracking-[0.1em]"
          style={{ textShadow: '0 0 30px rgba(0,240,255,0.15)' }}
        >
          ENTER THE VOID
        </h2>
        <p className="relative text-base text-[#556677] max-w-sm mx-auto mb-12 tracking-wide leading-relaxed">
          Register as an operator. Deploy into your first match. Nothing survives the void forever.
        </p>

        {!isAuthenticated ? (
          <div className="relative flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/register"
              className="px-10 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.35)] hover:scale-[1.02] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              Enlist Now
            </Link>
            <Link
              href="/leaderboard"
              className="px-10 py-3.5 text-sm font-bold text-[#667788] border border-[#00f0ff]/20 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 hover:text-[#00f0ff] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              View Rankings
            </Link>
          </div>
        ) : (
          <div className="relative flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/game"
              className="px-12 py-3.5 text-sm font-bold text-[#0a0a12] bg-[#00f0ff] rounded-md hover:shadow-[0_0_30px_rgba(0,240,255,0.35)] hover:scale-[1.02] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              Deploy
            </Link>
            <Link
              href="/leaderboard"
              className="px-10 py-3.5 text-sm font-bold text-[#667788] border border-[#00f0ff]/20 rounded-md hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 hover:text-[#00f0ff] transition-all duration-300 uppercase tracking-[0.2em]"
            >
              View Rankings
            </Link>
          </div>
        )}
      </section>

      {/* ── Footer ── */}
      <footer className="border-t border-[#00f0ff]/10 py-10 px-6 sm:px-12">
        <div className="max-w-5xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-6">
          <span className="text-[11px] text-[#334455] tracking-[0.2em] uppercase">
            v0.1 // Sector 7-G // Cosmic Void
          </span>
          <div className="flex gap-8">
            <Link href="/leaderboard" className="text-[11px] text-[#445566] hover:text-[#00f0ff] tracking-[0.15em] uppercase transition-colors">
              Rankings
            </Link>
            <Link href="/subscription" className="text-[11px] text-[#445566] hover:text-[#00f0ff] tracking-[0.15em] uppercase transition-colors">
              Premium
            </Link>
            <Link href="/profile" className="text-[11px] text-[#445566] hover:text-[#00f0ff] tracking-[0.15em] uppercase transition-colors">
              Profile
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}

/* ── Shared Section Components ── */

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-[11px] text-[#ff00aa] tracking-[0.35em] uppercase mb-4 text-center font-medium">
      {children}
    </p>
  );
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2
      className="text-2xl sm:text-3xl font-bold font-orbitron text-center text-[#00f0ff] mb-16 tracking-[0.1em] uppercase"
      style={{ textShadow: '0 0 25px rgba(0,240,255,0.12)' }}
    >
      {children}
    </h2>
  );
}

function Divider({ variant = 'cyan' }: { variant?: 'cyan' | 'magenta' }) {
  const color = variant === 'cyan' ? 'rgba(0,240,255,0.12)' : 'rgba(255,0,170,0.08)';
  return (
    <div className="max-w-xl mx-auto px-6 sm:px-12">
      <div className="h-px" style={{ background: `linear-gradient(to right, transparent, ${color}, transparent)` }} />
    </div>
  );
}

/* ── Card Components ── */

function MechanicCard({ title, description, color }: { title: string; description: string; color: 'cyan' | 'magenta' }) {
  const isCyan = color === 'cyan';
  const borderColor = isCyan ? 'rgba(0,240,255,0.1)' : 'rgba(255,0,170,0.1)';
  const titleColor = isCyan ? '#00f0ff' : '#ff00aa';
  const hoverBorder = isCyan ? 'hover:border-[#00f0ff]/25' : 'hover:border-[#ff00aa]/25';
  const bgGlow = isCyan
    ? 'radial-gradient(ellipse at 50% 0%, rgba(0,240,255,0.03) 0%, transparent 70%)'
    : 'radial-gradient(ellipse at 50% 0%, rgba(255,0,170,0.02) 0%, transparent 70%)';

  return (
    <div
      className={`p-7 rounded-lg border transition-all duration-300 ${hoverBorder}`}
      style={{ borderColor, background: bgGlow }}
    >
      <h3 className="text-sm font-bold tracking-[0.15em] mb-4" style={{ color: titleColor }}>
        {title}
      </h3>
      <p className="text-[13px] text-[#667788] leading-[1.9]">
        {description}
      </p>
    </div>
  );
}

function GearCard({ title, items, color }: { title: string; items: string[]; color: 'cyan' | 'magenta' }) {
  const isCyan = color === 'cyan';
  const borderColor = isCyan ? 'rgba(0,240,255,0.1)' : 'rgba(255,0,170,0.1)';
  const titleColor = isCyan ? '#00f0ff' : '#ff00aa';
  const dotColor = isCyan ? 'bg-[#00f0ff]/50' : 'bg-[#ff00aa]/50';
  const bgGlow = isCyan
    ? 'radial-gradient(ellipse at 50% 0%, rgba(0,240,255,0.03) 0%, transparent 70%)'
    : 'radial-gradient(ellipse at 50% 0%, rgba(255,0,170,0.02) 0%, transparent 70%)';

  return (
    <div
      className="p-7 rounded-lg border transition-all duration-300"
      style={{ borderColor, background: bgGlow }}
    >
      <h3 className="text-sm font-bold tracking-[0.15em] mb-5" style={{ color: titleColor }}>
        {title}
      </h3>
      <ul className="space-y-3">
        {items.map((item) => (
          <li key={item} className="flex items-start gap-3 text-[13px] text-[#667788] leading-relaxed">
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
    <div className="p-6 rounded-lg border border-[#00f0ff]/10" style={{
      background: 'radial-gradient(ellipse at 50% 0%, rgba(0,240,255,0.02) 0%, transparent 70%)',
    }}>
      <span className="text-[10px] text-[#445566] tracking-[0.25em] font-bold">{step}</span>
      <h3 className="text-sm font-bold text-[#00f0ff] tracking-[0.12em] mt-3 mb-3">{title}</h3>
      <p className="text-[13px] text-[#556677] leading-[1.8]">{description}</p>
    </div>
  );
}

function ControlKey({ keys, label }: { keys: string; label: string }) {
  return (
    <div className="flex items-center gap-3 p-3.5 rounded-md border border-[#00f0ff]/10 bg-[#00f0ff]/[0.02]">
      <kbd className="text-xs font-bold text-[#00f0ff] tracking-[0.1em] font-mono bg-[#00f0ff]/5 px-2.5 py-1.5 rounded border border-[#00f0ff]/15 min-w-[48px] text-center">
        {keys}
      </kbd>
      <span className="text-xs text-[#667788] tracking-wide">{label}</span>
    </div>
  );
}

const rarityColors = {
  Common: { text: '#667788', border: 'rgba(102,119,136,0.25)', bg: 'rgba(102,119,136,0.06)' },
  Rare: { text: '#00f0ff', border: 'rgba(0,240,255,0.25)', bg: 'rgba(0,240,255,0.06)' },
  Epic: { text: '#bf5fff', border: 'rgba(191,95,255,0.25)', bg: 'rgba(191,95,255,0.06)' },
  Legendary: { text: '#ff00aa', border: 'rgba(255,0,170,0.25)', bg: 'rgba(255,0,170,0.06)' },
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
