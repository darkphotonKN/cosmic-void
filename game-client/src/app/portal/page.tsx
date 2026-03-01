"use client";

import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import ParticleText from "@/components/ParticleText";

export default function PortalPage() {
  const router = useRouter();
  const [showHint, setShowHint] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => setShowHint(false), 4000);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="splash-container">
      <ParticleText />
      <div className="splash-overlay">
        <p className="splash-subtitle">
          NAVIGATE THE COSMOS. SURVIVE THE VOID.
        </p>
        <button
          className="splash-enter-btn"
          onClick={() => router.push("/login")}
        >
          ENTER THE VOID
        </button>
      </div>
    </div>
  );
}
