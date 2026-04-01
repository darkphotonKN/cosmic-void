"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:7001";

const ParticleAnimation = () => {
  const [randomNumber, setRandomNumber] = useState(0);
  useEffect(() => {
    setRandomNumber(Math.random());
  }, []);
  return (
    <div
      className="login-particle"
      style={{
        left: `${randomNumber * 100}%`,
        animationDelay: `${randomNumber * 5}s`,
        animationDuration: `${3 + randomNumber * 4}s`,
      }}
    />
  );
};

export default function RegisterPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isPolling, setIsPolling] = useState(false);
  const [error, setError] = useState("");
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
    setIsPolling(false);
  }, []);

  useEffect(() => {
    return () => stopPolling();
  }, [stopPolling]);

  const startPolling = useCallback(
    (targetEmail: string) => {
      setIsPolling(true);
      let attempts = 0;
      const maxAttempts = 15;

      pollingRef.current = setInterval(async () => {
        attempts++;
        try {
          const res = await fetch(
            `${API_BASE_URL}/api/member/check-email?email=${encodeURIComponent(targetEmail)}`,
          );
          const data = await res.json();

          if (data.exists) {
            stopPolling();
            router.push("/login");
            return;
          }
        } catch {
          // ignore, keep trying
        }

        if (attempts >= maxAttempts) {
          stopPolling();
          setError(
            "Registration is taking longer than expected. Please try logging in or try again.",
          );
        }
      }, 1000);
    },
    [router, stopPolling],
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("Please enter your callsign");
      return;
    }

    if (!email.trim()) {
      setError("Please enter email");
      return;
    }

    if (password.length < 6) {
      setError("Password must be at least 6 characters");
      return;
    }

    setIsLoading(true);

    try {
      const response = await fetch(`${API_BASE_URL}/api/member/signup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name, email, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Registration failed");
        return;
      }

      // Signup accepted, start polling
      setIsLoading(false);
      startPolling(email);
    } catch (err) {
      setError("Connection error. Please try again.");
      setIsLoading(false);
    }
  };

  return (
    <main className="login-container">
      {/* 背景格線效果 */}
      <div className="login-grid-bg" />

      {/* 浮動粒子效果 */}
      <div className="login-particles">
        {[...Array(20)].map((_, i) => (
          <ParticleAnimation key={i} />
        ))}
      </div>

      {/* 註冊框 */}
      <div className="login-box">
        {/* 標題 */}
        <div className="login-header">
          <h1 className="login-title">COSMIC VOID</h1>
          <p className="login-subtitle">Operator Registration</p>
        </div>

        {/* 表單 */}
        <form onSubmit={handleSubmit} className="login-form">
          <div className="login-input-group">
            <label htmlFor="name" className="login-label">
              Callsign
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="login-input"
              placeholder="Enter callsign..."
              autoComplete="name"
            />
          </div>

          <div className="login-input-group">
            <label htmlFor="email" className="login-label">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="login-input"
              placeholder="Enter email..."
              autoComplete="email"
            />
          </div>

          <div className="login-input-group">
            <label htmlFor="password" className="login-label">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="login-input"
              placeholder="At least 6 characters..."
              autoComplete="new-password"
            />
          </div>

          {error && <p className="login-error">{error}</p>}

          <button
            type="submit"
            className={`login-button ${isLoading || isPolling ? "loading" : ""}`}
            disabled={isLoading || isPolling}
          >
            {isLoading ? (
              <span className="login-loading">
                <span className="login-spinner" />
                Transmitting...
              </span>
            ) : isPolling ? (
              <span className="login-loading">
                <span className="login-spinner" />
                Provisioning Access...
              </span>
            ) : (
              "Register Operator"
            )}
          </button>
        </form>

        {/* 底部連結 */}
        <div className="login-footer">
          <p>
            Already registered?{" "}
            <a href="/login" className="login-link">
              Sign In
            </a>
          </p>
        </div>
      </div>

      {/* 版本資訊 */}
      <div className="login-version">v0.1 // SECTOR 7-G</div>
    </main>
  );
}
