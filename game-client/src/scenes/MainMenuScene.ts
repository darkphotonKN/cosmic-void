import Phaser from 'phaser';
import { createStarfield } from '../utils/Background';

export class MainMenuScene extends Phaser.Scene {
  constructor() {
    super({ key: 'MainMenuScene' });
  }

  create(): void {
    const width = this.cameras.main.width;
    const height = this.cameras.main.height;

    // 星空背景
    createStarfield(this);

    // Title
    const title = this.add.text(width / 2, height / 3, 'VOID RAIDERS', {
      fontSize: '48px',
      color: '#ffffff',
      fontStyle: 'bold',
    });
    title.setOrigin(0.5);

    // Start button
    const startButton = this.add.text(width / 2, height / 2, 'START GAME', {
      fontSize: '24px',
      color: '#ffffff',
    });
    startButton.setOrigin(0.5);
    startButton.setInteractive({ useHandCursor: true });

    startButton.on('pointerover', () => {
      startButton.setColor('#ffff00');
    });

    startButton.on('pointerout', () => {
      startButton.setColor('#ffffff');
    });

    startButton.on('pointerdown', () => {
      this.scene.start('GameScene');
    });
  }
}
