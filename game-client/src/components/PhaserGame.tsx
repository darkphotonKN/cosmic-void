"use client";

import { useEffect, useRef } from "react";
import Phaser from "phaser";
import { MainMenuScene } from "@/scenes/MainMenuScene";
import { CosmicVoidScene } from "@/scenes/CosmicVoidScene";
import { PreloadScene } from "@/scenes/PreloadScene";
import { BootScene } from "@/scenes/BootScene";
import { LoadoutScene } from "@/scenes/LoadoutScene";

export default function PhaserGame() {
  const gameRef = useRef<Phaser.Game | null>(null);

  useEffect(() => {
    if (gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 1080,
      height: 720,
      parent: "game-container",
      backgroundColor: "#1a1a2e",
      physics: {
        default: "arcade",
        arcade: {
          gravity: { x: 0, y: 0 },
          debug: false,
        },
      },
      scene: [BootScene, PreloadScene, MainMenuScene, LoadoutScene, CosmicVoidScene],
    };

    gameRef.current = new Phaser.Game(config);

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
    </div>
  );
}
