import Phaser from 'phaser';

export class BootScene extends Phaser.Scene {
  constructor() {
    super({ key: 'BootScene' });
  }

  preload(): void {
    // Load minimal assets needed for preloader
  }

  create(): void {
    this.scene.start('PreloadScene');
  }
}
