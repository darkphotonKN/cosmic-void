import Phaser from "phaser";

export class Player extends Phaser.Physics.Arcade.Sprite {
  private speed: number = 200;
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;

  constructor(scene: Phaser.Scene, x: number, y: number, texture: string) {
    super(scene, x, y, texture);

    scene.add.existing(this);
    scene.physics.add.existing(this);

    this.setCollideWorldBounds(true);
    this.cursors = scene.input.keyboard!.createCursorKeys();
  }

  update(): void {
    this.setVelocity(0);

    if (this.cursors.left.isDown) {
      this.setVelocityX(-this.speed);
    } else if (this.cursors.right.isDown) {
      this.setVelocityX(this.speed);
    }

    if (this.cursors.up.isDown) {
      this.setVelocityY(-this.speed);
    } else if (this.cursors.down.isDown) {
      this.setVelocityY(this.speed);
    }
  }

  setSpeed(speed: number): void {
    this.speed = speed;
  }
}
