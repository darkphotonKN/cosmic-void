"use client";

import { useEffect, useRef, useState } from "react";
import Phaser from "phaser";
import { MainMenuScene } from "@/scenes/MainMenuScene";
import { TreasureHuntScene } from "@/scenes/TreasureHuntScene";

export default function PhaserGame() {
  const gameRef = useRef<Phaser.Game | null>(null);
  const [status, setStatus] = useState({ text: "", color: "#4ecca3" });
  const [isInGame, setIsInGame] = useState(false);

  useEffect(() => {
    if (gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 900,
      height: 600,
      parent: "game-container",
      backgroundColor: "#1a1a2e",
      physics: {
        default: "arcade",
        arcade: {
          gravity: { x: 0, y: 0 },
          debug: false,
        },
      },
      scene: [MainMenuScene, TreasureHuntScene],
    };

    gameRef.current = new Phaser.Game(config);

    // 監聽場景切換
    gameRef.current.events.on("ready", () => {
      const game = gameRef.current;
      if (!game) return;

      // 監聽 TreasureHuntScene 啟動
      game.scene.getScene("TreasureHuntScene")?.events.on("create", () => {
        setIsInGame(true);
        const scene = game.scene.getScene("TreasureHuntScene") as TreasureHuntScene;
        scene.setStatusCallback((text, color) => {
          setStatus({ text, color });
        });
      });

      // 監聽回到主選單
      game.scene.getScene("MainMenuScene")?.events.on("create", () => {
        setIsInGame(false);
        setStatus({ text: "", color: "#4ecca3" });
      });
    });

    return () => {
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
      }
    };
  }, []);

  return (
    <div className="treasure-hunt-wrapper">
      <div id="game-container" className="treasure-hunt-game-container" />

      {isInGame && (
        <>
          <div className="treasure-hunt-controls">
            <div className="treasure-hunt-control-group">
              <h3>🎮 移動</h3>
              <p>WASD 或 方向鍵</p>
            </div>
            <div className="treasure-hunt-control-group">
              <h3>⚔️ 攻擊</h3>
              <p>空白鍵</p>
            </div>
            <div className="treasure-hunt-control-group">
              <h3>📦 撿取</h3>
              <p>E 鍵（靠近物品時）</p>
            </div>
            <div className="treasure-hunt-control-group">
              <h3>🚪 返回</h3>
              <p>ESC 鍵</p>
            </div>
          </div>

          {status.text && (
            <div
              className="treasure-hunt-status"
              style={{ color: status.color }}
            >
              {status.text}
            </div>
          )}
        </>
      )}
    </div>
  );
}
