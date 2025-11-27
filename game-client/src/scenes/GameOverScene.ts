import Phaser from "phaser";

export class GameOverScene extends Phaser.Scene {
  constructor() {
    super({ key: "GameOverScene" });
  }

  create(data: { score: number }): void {
    const width = this.cameras.main.width;
    const height = this.cameras.main.height;

    // Game Over text
    const gameOverText = this.add.text(width / 2, height / 3, "GAME OVER", {
      fontSize: "48px",
      color: "#ff0000",
      fontStyle: "bold",
    });
    gameOverText.setOrigin(0.5);

    // Score
    const scoreText = this.add.text(
      width / 2,
      height / 2,
      `Final Score: ${data.score || 0}`,
      {
        fontSize: "24px",
        color: "#ffffff",
      },
    );
    scoreText.setOrigin(0.5);

    // Restart button
    const restartButton = this.add.text(
      width / 2,
      height * 0.65,
      "PLAY AGAIN",
      {
        fontSize: "20px",
        color: "#ffffff",
      },
    );
    restartButton.setOrigin(0.5);
    restartButton.setInteractive({ useHandCursor: true });

    restartButton.on("pointerover", () => restartButton.setColor("#ffff00"));
    restartButton.on("pointerout", () => restartButton.setColor("#ffffff"));
    restartButton.on("pointerdown", () => this.scene.start("GameScene"));

    // Menu button
    const menuButton = this.add.text(width / 2, height * 0.75, "MAIN MENU", {
      fontSize: "20px",
      color: "#ffffff",
    });
    menuButton.setOrigin(0.5);
    menuButton.setInteractive({ useHandCursor: true });

    menuButton.on("pointerover", () => menuButton.setColor("#ffff00"));
    menuButton.on("pointerout", () => menuButton.setColor("#ffffff"));
    menuButton.on("pointerdown", () => this.scene.start("MainMenuScene"));
  }
}
