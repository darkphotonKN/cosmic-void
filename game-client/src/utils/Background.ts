import Phaser from "phaser";

export function createStarfield(
  scene: Phaser.Scene,
  starCount: number = 200,
): void {
  const width = scene.cameras.main.width;
  const height = scene.cameras.main.height;

  for (let i = 0; i < starCount; i++) {
    const x = Phaser.Math.Between(0, width);
    const y = Phaser.Math.Between(0, height);
    const size = Phaser.Math.Between(1, 2);
    const star = scene.add.circle(x, y, size, 0xffffff, 0.8);
    star.setDepth(10);
  }
}
