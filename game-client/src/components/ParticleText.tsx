'use client';

import { useEffect, useRef, useCallback } from 'react';

interface Particle {
  x: number;
  y: number;
  originX: number;
  originY: number;
  color: string;
  size: number;
  vx: number;
  vy: number;
}

interface Star {
  x: number;
  y: number;
  size: number;
  baseOpacity: number;
  twinkleSpeed: number;
  twinkleOffset: number;
}

export default function ParticleText() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const particlesRef = useRef<Particle[]>([]);
  const starsRef = useRef<Star[]>([]);
  const mouseRef = useRef({ x: -9999, y: -9999 });
  const animationRef = useRef<number>(0);

  const initParticles = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const w = window.innerWidth;
    const h = window.innerHeight;
    canvas.width = w;
    canvas.height = h;

    // Resolve the CSS variable to get the actual loaded font family
    const fontFamily =
      getComputedStyle(document.documentElement)
        .getPropertyValue('--font-orbitron')
        .trim() || 'Orbitron';

    // Offscreen canvas to render text and sample pixels
    const offscreen = document.createElement('canvas');
    offscreen.width = w;
    offscreen.height = h;
    const offCtx = offscreen.getContext('2d');
    if (!offCtx) return;

    const fontSize = Math.min(w / 6, 140);
    const lineGap = fontSize * 0.15;

    offCtx.fillStyle = '#fff';
    offCtx.font = `900 ${fontSize}px ${fontFamily}, Orbitron, sans-serif`;
    offCtx.textAlign = 'center';
    offCtx.textBaseline = 'middle';
    offCtx.fillText('COSMIC', w / 2, h / 2 - fontSize / 2 - lineGap);
    offCtx.fillText('VOID', w / 2, h / 2 + fontSize / 2 + lineGap);

    // Sample pixel data to find text positions
    const imageData = offCtx.getImageData(0, 0, w, h);
    const data = imageData.data;
    const particles: Particle[] = [];
    const gap = Math.max(3, Math.floor(Math.min(w, h) / 250));
    const colors = ['#ffffff'];

    for (let y = 0; y < h; y += gap) {
      for (let x = 0; x < w; x += gap) {
        const alpha = data[(y * w + x) * 4 + 3];
        if (alpha > 128) {
          particles.push({
            x: Math.random() * w,
            y: Math.random() * h,
            originX: x,
            originY: y,
            color: colors[Math.floor(Math.random() * colors.length)],
            size: Math.random() * 1.5 + 0.5,
            vx: 0,
            vy: 0,
          });
        }
      }
    }

    particlesRef.current = particles;

    // Generate background stars
    const starCount = Math.floor((w * h) / 8000);
    const stars: Star[] = [];
    for (let i = 0; i < starCount; i++) {
      stars.push({
        x: Math.random() * w,
        y: Math.random() * h,
        size: Math.random() * 1.8 + 0.2,
        baseOpacity: Math.random() * 0.5 + 0.15,
        twinkleSpeed: Math.random() * 0.02 + 0.005,
        twinkleOffset: Math.random() * Math.PI * 2,
      });
    }
    starsRef.current = stars;
  }, []);

  const animate = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Draw stars with twinkle
    const time = performance.now() * 0.001;
    for (const s of starsRef.current) {
      const opacity =
        s.baseOpacity + Math.sin(time * s.twinkleSpeed * 60 + s.twinkleOffset) * s.baseOpacity * 0.6;
      ctx.fillStyle = `rgba(255, 255, 255, ${opacity})`;
      ctx.beginPath();
      ctx.arc(s.x, s.y, s.size, 0, Math.PI * 2);
      ctx.fill();
    }

    const { x: mx, y: my } = mouseRef.current;
    const repulseRadius = 220;
    const friction = 0.96;
    const lerpSpeed = 0.025;

    for (const p of particlesRef.current) {
      // Mouse repulsion — gentle, wide push
      const dx = p.x - mx;
      const dy = p.y - my;
      const distSq = dx * dx + dy * dy;

      if (distSq < repulseRadius * repulseRadius && distSq > 0) {
        const dist = Math.sqrt(distSq);
        const force = ((repulseRadius - dist) / repulseRadius) * 0.4;
        p.vx += (dx / dist) * force;
        p.vy += (dy / dist) * force;
      }

      // Ambient drift — subtle space-floating feel
      p.vx += (Math.random() - 0.5) * 0.03;
      p.vy += (Math.random() - 0.5) * 0.03;

      // Apply and decay velocity (repulsion + drift only)
      p.x += p.vx;
      p.y += p.vy;
      p.vx *= friction;
      p.vy *= friction;

      // Lerp toward origin — smooth convergence, no bounce
      p.x += (p.originX - p.x) * lerpSpeed;
      p.y += (p.originY - p.y) * lerpSpeed;

      // Draw particle
      ctx.fillStyle = p.color;
      ctx.fillRect(p.x, p.y, p.size, p.size);
    }

    animationRef.current = requestAnimationFrame(animate);
  }, []);

  useEffect(() => {
    let resizeTimer: ReturnType<typeof setTimeout>;

    const start = () => {
      initParticles();
      animationRef.current = requestAnimationFrame(animate);
    };

    // Wait for fonts before sampling text pixels
    document.fonts.ready.then(start);

    const handleResize = () => {
      cancelAnimationFrame(animationRef.current);
      clearTimeout(resizeTimer);
      resizeTimer = setTimeout(start, 200);
    };

    const handleMouseMove = (e: MouseEvent) => {
      mouseRef.current = { x: e.clientX, y: e.clientY };
    };

    const handleMouseLeave = () => {
      mouseRef.current = { x: -9999, y: -9999 };
    };

    const handleTouchMove = (e: TouchEvent) => {
      if (e.touches.length > 0) {
        mouseRef.current = {
          x: e.touches[0].clientX,
          y: e.touches[0].clientY,
        };
      }
    };

    const handleTouchEnd = () => {
      mouseRef.current = { x: -9999, y: -9999 };
    };

    window.addEventListener('resize', handleResize);
    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseleave', handleMouseLeave);
    window.addEventListener('touchmove', handleTouchMove);
    window.addEventListener('touchend', handleTouchEnd);

    return () => {
      cancelAnimationFrame(animationRef.current);
      clearTimeout(resizeTimer);
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseleave', handleMouseLeave);
      window.removeEventListener('touchmove', handleTouchMove);
      window.removeEventListener('touchend', handleTouchEnd);
    };
  }, [initParticles, animate]);

  return <canvas ref={canvasRef} className="splash-canvas" />;
}
