"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/utils/api";

interface PlayerRankingStats {
  id: string;
  member_id: string;
  username: string;
  wins: number;
  top_threes: number;
  avatar_url: string;
  rating: number;
  rank_position?: number;
}

interface LeaderboardResponse {
  players: PlayerRankingStats[];
  total_count: number;
}

export default function LeaderboardPage() {
  const router = useRouter();
  const [leaderboard, setLeaderboard] = useState<PlayerRankingStats[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [currentPage, setCurrentPage] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const pageSize = 20;

  useEffect(() => {
    fetchLeaderboard();
  }, [currentPage]);

  const fetchLeaderboard = async () => {
    setIsLoading(true);
    setError("");

    try {
      const response: LeaderboardResponse = await apiClient.getLeaderboard(
        pageSize,
        currentPage * pageSize
      );
      setLeaderboard(response.players || []);
      setTotalCount(response.total_count || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load leaderboard");
    } finally {
      setIsLoading(false);
    }
  };

  const totalPages = Math.ceil(totalCount / pageSize);

  const getRankIndicator = (index: number) => {
    if (index === 0) return "I";
    if (index === 1) return "II";
    if (index === 2) return "III";
    return null;
  };

  return (
    <main className="leaderboard-container">
      <div className="leaderboard-bg" />

      <div className="leaderboard-header">
        <button
          onClick={() => router.push("/game")}
          className="leaderboard-back-btn"
        >
          Back to Game
        </button>
        <h1 className="leaderboard-title">OPERATOR RANKINGS</h1>
      </div>

      <div className="leaderboard-content">
        {isLoading ? (
          <div className="leaderboard-loading">
            <div className="loading-spinner"></div>
            <p>Loading rankings...</p>
          </div>
        ) : error ? (
          <div className="leaderboard-error">
            <p>{error}</p>
            <button onClick={fetchLeaderboard} className="retry-btn">
              Retry
            </button>
          </div>
        ) : leaderboard.length === 0 ? (
          <div className="leaderboard-empty">
            <p>No operators ranked yet</p>
          </div>
        ) : (
          <>
            <div className="leaderboard-table">
              <div className="leaderboard-header-row">
                <div className="leaderboard-col rank-col">Rank</div>
                <div className="leaderboard-col player-col">Operator</div>
                <div className="leaderboard-col wins-col">Wins</div>
                <div className="leaderboard-col top3-col">Top 3</div>
              </div>

              {leaderboard.map((player, index) => (
                <div
                  key={player.id}
                  className={`leaderboard-row ${
                    index < 3 ? `top-${index + 1}` : ""
                  }`}
                >
                  <div className="leaderboard-col rank-col">
                    <span className={`rank-number rank-${index + 1}`}>
                      {currentPage * pageSize + index + 1}
                    </span>
                    {getRankIndicator(index) && (
                      <span className={`rank-badge rank-badge-${index + 1}`}>
                        {getRankIndicator(index)}
                      </span>
                    )}
                  </div>

                  <div className="leaderboard-col player-col">
                    <div className="player-info">
                      <div className="player-avatar">
                        {player.avatar_url ? (
                          <img
                            src={player.avatar_url}
                            alt={player.username}
                            className="avatar-img"
                          />
                        ) : (
                          <div className="avatar-placeholder">
                            {player.username?.charAt(0)?.toUpperCase() || "?"}
                          </div>
                        )}
                      </div>
                      <span className="player-name">{player.username || "Unknown"}</span>
                    </div>
                  </div>

                  <div className="leaderboard-col wins-col">
                    <span className="stat-value">{player.wins}</span>
                  </div>

                  <div className="leaderboard-col top3-col">
                    <span className="stat-value">{player.top_threes}</span>
                  </div>
                </div>
              ))}
            </div>

            {totalPages > 1 && (
              <div className="leaderboard-pagination">
                <button
                  onClick={() => setCurrentPage((p) => Math.max(0, p - 1))}
                  disabled={currentPage === 0}
                  className="pagination-btn"
                >
                  Prev
                </button>
                <span className="page-info">
                  {currentPage + 1} / {totalPages}
                </span>
                <button
                  onClick={() => setCurrentPage((p) => Math.min(totalPages - 1, p + 1))}
                  disabled={currentPage >= totalPages - 1}
                  className="pagination-btn"
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </div>

      <style jsx>{`
        .leaderboard-container {
          min-height: 100vh;
          padding: 2rem;
          position: relative;
        }

        .leaderboard-bg {
          position: fixed;
          inset: 0;
          background: #0a0a12;
          z-index: -1;
        }

        .leaderboard-bg::before {
          content: "";
          position: absolute;
          inset: 0;
          background-image: radial-gradient(circle at 20% 50%, rgba(0, 240, 255, 0.04) 0%, transparent 50%),
                            radial-gradient(circle at 80% 80%, rgba(255, 0, 170, 0.03) 0%, transparent 50%);
        }

        .leaderboard-header {
          max-width: 1200px;
          margin: 0 auto 3rem;
          display: flex;
          align-items: center;
          gap: 2rem;
        }

        .leaderboard-back-btn {
          padding: 0.5rem 1rem;
          background: rgba(0, 240, 255, 0.05);
          color: #556677;
          border: 1px solid rgba(0, 240, 255, 0.15);
          border-radius: 6px;
          font-size: 0.85rem;
          cursor: pointer;
          transition: all 0.3s ease;
          letter-spacing: 0.05em;
        }

        .leaderboard-back-btn:hover {
          background: rgba(0, 240, 255, 0.1);
          color: #00f0ff;
          border-color: rgba(0, 240, 255, 0.3);
        }

        .leaderboard-title {
          font-size: 2rem;
          font-weight: 700;
          color: #00f0ff;
          letter-spacing: 0.15em;
          text-shadow: 0 0 20px rgba(0, 240, 255, 0.3);
        }

        .leaderboard-content {
          max-width: 1000px;
          margin: 0 auto;
          background: rgba(10, 10, 20, 0.85);
          backdrop-filter: blur(10px);
          border-radius: 10px;
          padding: 2rem;
          border: 1px solid rgba(0, 240, 255, 0.1);
        }

        .leaderboard-loading,
        .leaderboard-error,
        .leaderboard-empty {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          min-height: 400px;
          color: #556677;
          gap: 1rem;
          letter-spacing: 0.05em;
        }

        .loading-spinner {
          width: 48px;
          height: 48px;
          border: 2px solid rgba(0, 240, 255, 0.1);
          border-left-color: #00f0ff;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }

        .retry-btn {
          padding: 0.5rem 1.5rem;
          background: rgba(0, 240, 255, 0.1);
          color: #00f0ff;
          border: 1px solid rgba(0, 240, 255, 0.2);
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.3s ease;
          letter-spacing: 0.1em;
          text-transform: uppercase;
          font-size: 0.8rem;
        }

        .retry-btn:hover {
          background: rgba(0, 240, 255, 0.15);
          border-color: rgba(0, 240, 255, 0.4);
        }

        .leaderboard-table {
          display: flex;
          flex-direction: column;
          gap: 0.4rem;
        }

        .leaderboard-header-row {
          display: grid;
          grid-template-columns: 100px 1fr 100px 100px;
          padding: 1rem;
          background: rgba(0, 240, 255, 0.03);
          border-radius: 6px;
          color: #445566;
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.15em;
        }

        .leaderboard-row {
          display: grid;
          grid-template-columns: 100px 1fr 100px 100px;
          padding: 1rem;
          background: rgba(255, 255, 255, 0.01);
          border-radius: 6px;
          transition: all 0.3s ease;
          border: 1px solid transparent;
        }

        .leaderboard-row:hover {
          background: rgba(0, 240, 255, 0.03);
          border-color: rgba(0, 240, 255, 0.08);
        }

        .leaderboard-row.top-1 {
          background: linear-gradient(90deg, rgba(0, 240, 255, 0.06) 0%, transparent 100%);
          border-color: rgba(0, 240, 255, 0.15);
        }

        .leaderboard-row.top-2 {
          background: linear-gradient(90deg, rgba(255, 0, 170, 0.04) 0%, transparent 100%);
          border-color: rgba(255, 0, 170, 0.1);
        }

        .leaderboard-row.top-3 {
          background: linear-gradient(90deg, rgba(0, 240, 255, 0.03) 0%, transparent 100%);
          border-color: rgba(0, 240, 255, 0.08);
        }

        .leaderboard-col {
          display: flex;
          align-items: center;
          color: #889aaa;
        }

        .rank-col {
          font-weight: 600;
          position: relative;
          gap: 0.75rem;
        }

        .rank-number {
          font-size: 1.1rem;
          font-family: var(--font-heading);
        }

        .rank-1 {
          color: #00f0ff;
          text-shadow: 0 0 10px rgba(0, 240, 255, 0.4);
        }

        .rank-2 {
          color: #ff00aa;
          text-shadow: 0 0 10px rgba(255, 0, 170, 0.3);
        }

        .rank-3 {
          color: #00f0ff;
        }

        .rank-badge {
          font-size: 0.7rem;
          font-weight: 700;
          letter-spacing: 0.1em;
          padding: 0.15rem 0.4rem;
          border-radius: 3px;
          font-family: var(--font-heading);
        }

        .rank-badge-1 {
          color: #00f0ff;
          background: rgba(0, 240, 255, 0.1);
          border: 1px solid rgba(0, 240, 255, 0.2);
        }

        .rank-badge-2 {
          color: #ff00aa;
          background: rgba(255, 0, 170, 0.1);
          border: 1px solid rgba(255, 0, 170, 0.2);
        }

        .rank-badge-3 {
          color: #556677;
          background: rgba(85, 102, 119, 0.1);
          border: 1px solid rgba(85, 102, 119, 0.2);
        }

        .player-info {
          display: flex;
          align-items: center;
          gap: 1rem;
        }

        .player-avatar {
          width: 36px;
          height: 36px;
          border-radius: 50%;
          overflow: hidden;
          border: 1px solid rgba(0, 240, 255, 0.15);
        }

        .avatar-img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }

        .avatar-placeholder {
          width: 100%;
          height: 100%;
          background: rgba(0, 240, 255, 0.06);
          display: flex;
          align-items: center;
          justify-content: center;
          color: #556677;
          font-weight: 600;
          font-size: 1rem;
          font-family: var(--font-heading);
        }

        .player-name {
          font-weight: 500;
          color: #ccdde8;
          letter-spacing: 0.03em;
        }

        .stat-value {
          font-weight: 600;
          font-size: 1rem;
          font-family: var(--font-heading);
          letter-spacing: 0.05em;
        }

        .leaderboard-pagination {
          margin-top: 2rem;
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 2rem;
        }

        .pagination-btn {
          padding: 0.5rem 1rem;
          background: rgba(0, 240, 255, 0.05);
          color: #556677;
          border: 1px solid rgba(0, 240, 255, 0.1);
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.3s ease;
          letter-spacing: 0.1em;
          text-transform: uppercase;
          font-size: 0.75rem;
        }

        .pagination-btn:hover:not(:disabled) {
          background: rgba(0, 240, 255, 0.1);
          color: #00f0ff;
          border-color: rgba(0, 240, 255, 0.2);
        }

        .pagination-btn:disabled {
          opacity: 0.3;
          cursor: not-allowed;
        }

        .page-info {
          color: #445566;
          font-size: 0.85rem;
          letter-spacing: 0.1em;
          font-family: var(--font-heading);
        }

        @media (max-width: 768px) {
          .leaderboard-header-row,
          .leaderboard-row {
            grid-template-columns: 60px 1fr 60px 60px;
            font-size: 0.8rem;
          }

          .leaderboard-title {
            font-size: 1.5rem;
          }

          .player-avatar {
            width: 28px;
            height: 28px;
          }

          .rank-badge {
            display: none;
          }
        }
      `}</style>
    </main>
  );
}
