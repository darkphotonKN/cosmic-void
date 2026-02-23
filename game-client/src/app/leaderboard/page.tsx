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

  return (
    <main className="leaderboard-container">
      {/* Background */}
      <div className="leaderboard-bg" />

      {/* Header */}
      <div className="leaderboard-header">
        <button
          onClick={() => router.push("/game")}
          className="leaderboard-back-btn"
        >
          ← Back to Game
        </button>
        <h1 className="leaderboard-title">Leaderboard</h1>
      </div>

      {/* Content */}
      <div className="leaderboard-content">
        {isLoading ? (
          <div className="leaderboard-loading">
            <div className="loading-spinner"></div>
            <p>Loading leaderboard...</p>
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
            <p>No players on the leaderboard yet</p>
          </div>
        ) : (
          <>
            <div className="leaderboard-table">
              <div className="leaderboard-header-row">
                <div className="leaderboard-col rank-col">Rank</div>
                <div className="leaderboard-col player-col">Player</div>
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
                    {index === 0 && <span className="medal gold">🥇</span>}
                    {index === 1 && <span className="medal silver">🥈</span>}
                    {index === 2 && <span className="medal bronze">🥉</span>}
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

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="leaderboard-pagination">
                <button
                  onClick={() => setCurrentPage((p) => Math.max(0, p - 1))}
                  disabled={currentPage === 0}
                  className="pagination-btn"
                >
                  Previous
                </button>
                <span className="page-info">
                  Page {currentPage + 1} of {totalPages}
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
          background: linear-gradient(135deg, #1a1a2e 0%, #0f0f1e 100%);
          z-index: -1;
        }

        .leaderboard-bg::before {
          content: "";
          position: absolute;
          inset: 0;
          background-image: radial-gradient(circle at 20% 50%, rgba(79, 70, 229, 0.1) 0%, transparent 50%),
                            radial-gradient(circle at 80% 80%, rgba(236, 72, 153, 0.1) 0%, transparent 50%);
        }

        .leaderboard-header {
          max-width: 1200px;
          margin: 0 auto 3rem;
          display: flex;
          align-items: center;
          gap: 2rem;
        }

        .leaderboard-back-btn {
          padding: 0.75rem 1.5rem;
          background: rgba(255, 255, 255, 0.1);
          color: white;
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 0.5rem;
          font-size: 0.875rem;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .leaderboard-back-btn:hover {
          background: rgba(255, 255, 255, 0.2);
          transform: translateX(-4px);
        }

        .leaderboard-title {
          font-size: 2.5rem;
          font-weight: 700;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }

        .leaderboard-content {
          max-width: 1000px;
          margin: 0 auto;
          background: rgba(255, 255, 255, 0.05);
          backdrop-filter: blur(10px);
          border-radius: 1rem;
          padding: 2rem;
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .leaderboard-loading,
        .leaderboard-error,
        .leaderboard-empty {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          min-height: 400px;
          color: rgba(255, 255, 255, 0.7);
          gap: 1rem;
        }

        .loading-spinner {
          width: 48px;
          height: 48px;
          border: 4px solid rgba(255, 255, 255, 0.1);
          border-left-color: #667eea;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }

        .retry-btn {
          padding: 0.5rem 1.5rem;
          background: #667eea;
          color: white;
          border: none;
          border-radius: 0.5rem;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .retry-btn:hover {
          background: #5a67d8;
          transform: translateY(-2px);
        }

        .leaderboard-table {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
        }

        .leaderboard-header-row {
          display: grid;
          grid-template-columns: 100px 1fr 100px 100px;
          padding: 1rem;
          background: rgba(255, 255, 255, 0.05);
          border-radius: 0.5rem;
          color: rgba(255, 255, 255, 0.5);
          font-size: 0.875rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        .leaderboard-row {
          display: grid;
          grid-template-columns: 100px 1fr 100px 100px;
          padding: 1rem;
          background: rgba(255, 255, 255, 0.02);
          border-radius: 0.5rem;
          transition: all 0.3s ease;
          border: 1px solid transparent;
        }

        .leaderboard-row:hover {
          background: rgba(255, 255, 255, 0.05);
          border-color: rgba(255, 255, 255, 0.1);
          transform: translateX(4px);
        }

        .leaderboard-row.top-1 {
          background: linear-gradient(90deg, rgba(255, 215, 0, 0.1) 0%, transparent 100%);
          border-color: rgba(255, 215, 0, 0.3);
        }

        .leaderboard-row.top-2 {
          background: linear-gradient(90deg, rgba(192, 192, 192, 0.1) 0%, transparent 100%);
          border-color: rgba(192, 192, 192, 0.3);
        }

        .leaderboard-row.top-3 {
          background: linear-gradient(90deg, rgba(205, 127, 50, 0.1) 0%, transparent 100%);
          border-color: rgba(205, 127, 50, 0.3);
        }

        .leaderboard-col {
          display: flex;
          align-items: center;
          color: white;
        }

        .rank-col {
          font-weight: 600;
          position: relative;
        }

        .rank-number {
          font-size: 1.25rem;
        }

        .rank-1 {
          color: #ffd700;
        }

        .rank-2 {
          color: #c0c0c0;
        }

        .rank-3 {
          color: #cd7f32;
        }

        .medal {
          position: absolute;
          left: 50px;
          font-size: 1.5rem;
          animation: bounce 2s infinite;
        }

        @keyframes bounce {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-4px); }
        }

        .player-info {
          display: flex;
          align-items: center;
          gap: 1rem;
        }

        .player-avatar {
          width: 40px;
          height: 40px;
          border-radius: 50%;
          overflow: hidden;
          border: 2px solid rgba(255, 255, 255, 0.2);
        }

        .avatar-img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }

        .avatar-placeholder {
          width: 100%;
          height: 100%;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-weight: 600;
          font-size: 1.25rem;
        }

        .player-name {
          font-weight: 500;
        }

        .stat-value {
          font-weight: 600;
          font-size: 1.1rem;
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
          background: rgba(255, 255, 255, 0.1);
          color: white;
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 0.5rem;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .pagination-btn:hover:not(:disabled) {
          background: rgba(255, 255, 255, 0.2);
        }

        .pagination-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .page-info {
          color: rgba(255, 255, 255, 0.7);
        }

        @media (max-width: 768px) {
          .leaderboard-header-row,
          .leaderboard-row {
            grid-template-columns: 60px 1fr 60px 60px;
            font-size: 0.875rem;
          }

          .leaderboard-title {
            font-size: 1.75rem;
          }

          .player-avatar {
            width: 32px;
            height: 32px;
          }

          .medal {
            display: none;
          }
        }
      `}</style>
    </main>
  );
}