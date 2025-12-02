import Phaser from 'phaser';

export class MainMenuScene extends Phaser.Scene {
  constructor() {
    super({ key: 'MainMenuScene' });
  }

  create(): void {
    const width = this.cameras.main.width;
    const height = this.cameras.main.height;

    // 背景
    this.cameras.main.setBackgroundColor('#1a1a2e');

    // 創建格子背景
    const graphics = this.add.graphics();
    graphics.lineStyle(1, 0x333355, 0.3);
    for (let x = 0; x <= width; x += 50) {
      graphics.lineBetween(x, 0, x, height);
    }
    for (let y = 0; y <= height; y += 50) {
      graphics.lineBetween(0, y, width, y);
    }

    // Title
    const title = this.add.text(width / 2, height / 4, '🗺️ 多人尋寶遊戲', {
      fontSize: '42px',
      color: '#4ecca3',
      fontStyle: 'bold',
    });
    title.setOrigin(0.5);

    // Subtitle
    const subtitle = this.add.text(width / 2, height / 4 + 50, 'Treasure Hunt Demo', {
      fontSize: '20px',
      color: '#888888',
    });
    subtitle.setOrigin(0.5);

    // Description
    const desc = this.add.text(width / 2, height / 2 - 30, '展示視野系統 (Fog of War)\n建築碰撞 + 室內視線遮擋', {
      fontSize: '16px',
      color: '#aaaaaa',
      align: 'center',
    });
    desc.setOrigin(0.5);

    // Start button
    const buttonBg = this.add.rectangle(width / 2, height / 2 + 60, 200, 50, 0x4ecca3);
    buttonBg.setInteractive({ useHandCursor: true });

    const startButton = this.add.text(width / 2, height / 2 + 60, '開始遊戲', {
      fontSize: '24px',
      color: '#1a1a2e',
      fontStyle: 'bold',
    });
    startButton.setOrigin(0.5);

    buttonBg.on('pointerover', () => {
      buttonBg.setFillStyle(0x3dbb92);
    });

    buttonBg.on('pointerout', () => {
      buttonBg.setFillStyle(0x4ecca3);
    });

    buttonBg.on('pointerdown', () => {
      this.scene.start('TreasureHuntScene');
    });

    // Controls info
    const controlsText = this.add.text(width / 2, height - 80,
      '🎮 WASD/方向鍵 移動  |  ⚔️ 空白鍵 攻擊  |  📦 E 撿取', {
      fontSize: '14px',
      color: '#666666',
    });
    controlsText.setOrigin(0.5);
  }
}
