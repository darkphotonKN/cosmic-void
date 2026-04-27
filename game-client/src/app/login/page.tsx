"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";

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
export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  // Surface auth errors from a forced disconnect (e.g. duplicate queue / session conflict).
  useEffect(() => {
    const reason = sessionStorage.getItem("auth-error-message");
    if (reason) {
      setError(reason);
      sessionStorage.removeItem("auth-error-message");
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!email.trim()) {
      setError("Please enter email");
      return;
    }

    if (!password.trim()) {
      setError("Please enter password");
      return;
    }

    setIsLoading(true);

    try {
      const response = await fetch(`${API_BASE_URL}/api/member/signin`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Login failed");
        return;
      }

      // Store auth data in zustand
      const result = data.result;
      setAuth({
        accessToken: result.access_token,
        refreshToken: result.refresh_token,
        memberInfo: result.member_info,
      });

      // Redirect to the intended destination or default to game
      const redirectTo = searchParams.get("redirect") || "/game";
      router.push(redirectTo);
    } catch (err) {
      setError("Connection error. Please try again.");
    } finally {
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

      {/* 登入框 */}
      <div className="login-box">
        {/* 標題 */}
        <div className="login-header">
          <h1 className="login-title">COSMIC VOID</h1>
          <p className="login-subtitle">Operator Authentication</p>
        </div>

        {/* 表單 */}
        <form onSubmit={handleSubmit} className="login-form">
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
              placeholder="Enter password..."
              autoComplete="current-password"
            />
          </div>

          {error && <p className="login-error">{error}</p>}

          <button
            type="submit"
            className={`login-button ${isLoading ? "loading" : ""}`}
            disabled={isLoading}
          >
            {isLoading ? (
              <span className="login-loading">
                <span className="login-spinner" />
                Verifying...
              </span>
            ) : (
              "Authenticate"
            )}
          </button>
        </form>

        {/* 底部連結 */}
        <div className="login-footer">
          <p>
            No credentials?{" "}
            <a href="/register" className="login-link">
              Register
            </a>
          </p>
        </div>
      </div>

      {/* 版本資訊 */}
      <div className="login-version">v0.1 // SECTOR 7-G</div>
    </main>
  );
}
