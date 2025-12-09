"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

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
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!username.trim()) {
      setError("Please enter username");
      return;
    }

    setIsLoading(true);

    // TODO: Actual login logic
    setTimeout(() => {
      setIsLoading(false);
      router.push("/game");
    }, 1000);
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
          <p className="login-subtitle">Multiplayer Treasure Hunt</p>
        </div>

        {/* 表單 */}
        <form onSubmit={handleSubmit} className="login-form">
          <div className="login-input-group">
            <label htmlFor="username" className="login-label">
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="login-input"
              placeholder="Enter username..."
              autoComplete="username"
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
                Connecting...
              </span>
            ) : (
              "Enter Game"
            )}
          </button>
        </form>

        {/* 底部連結 */}
        <div className="login-footer">
          <p>
            Don&apos;t have an account?{" "}
            <a href="/register" className="login-link">
              Register
            </a>
          </p>
        </div>
      </div>

      {/* 版本資訊 */}
      <div className="login-version">v0.1.0 Alpha</div>
    </main>
  );
}
