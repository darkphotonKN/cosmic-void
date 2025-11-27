"use client";

import { useEffect, useRef } from "react";
import Phaser from "phaser";
import { BootScene } from "@/scenes/BootScene";
import { PreloadScene } from "@/scenes/PreloadScene";
import { MainMenuScene } from "@/scenes/MainMenuScene";
import { GameScene } from "@/scenes/GameScene";
import { GameOverScene } from "@/scenes/GameOverScene";

export default function PhaserGame() {
  const gameRef = useRef<Phaser.Game | null>(null);

  useEffect(() => {
    if (gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 800,
      height: 600,
      parent: "game-container",
      backgroundColor: "#000000",
      physics: {
        default: "arcade",
        arcade: {
          gravity: { x: 0, y: 0 },
          debug: false,
        },
      },
      scene: [BootScene, PreloadScene, MainMenuScene, GameScene, GameOverScene],
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
    <div id="game-container" className="flex items-center justify-center" />
  );
}
