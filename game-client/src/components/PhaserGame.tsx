"use client";

import { useEffect, useRef, useState } from "react";
import Phaser from "phaser";
import { MainMenuScene } from "@/scenes/MainMenuScene";
import { TreasureHuntScene } from "@/scenes/TreasureHuntScene";
import { PreloadScene } from "@/scenes/PreloadScene";
import { BootScene } from "@/scenes/BootScene";

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
      scene: [BootScene, PreloadScene, MainMenuScene, TreasureHuntScene],
    };

    gameRef.current = new Phaser.Game(config);

    // 監聽場景切換
    gameRef.current.events.on("ready", () => {
      const game = gameRef.current;
      if (!game) return;

      // 監聽 TreasureHuntScene 啟動
      game.scene.getScene("TreasureHuntScene")?.events.on("create", () => {
        setIsInGame(true);
        const scene = game.scene.getScene(
          "TreasureHuntScene",
        ) as TreasureHuntScene;
        scene.setStatusCallback((text, color) => {
          setStatus({ text, color });
        });
      });

      // 監聯回到主選單
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
              <h3>🎮 Move</h3>
              <p>WASD or Arrow Keys</p>
            </div>
            <div className="treasure-hunt-control-group">
              <h3>🚪 Back</h3>
              <p>ESC Key</p>
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
