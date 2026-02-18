/**
 * sprite generator for astronaut character
 * creates programmatic sprites for the game since we dont have image assets
 */

export class AstronautSpriteGenerator {
  private canvas: HTMLCanvasElement;
  private ctx: CanvasRenderingContext2D;
  private frameWidth = 36;
  private frameHeight = 36;

  constructor() {
    this.canvas = document.createElement('canvas');
    this.ctx = this.canvas.getContext('2d')!;
  }

  generateSpriteSheet(): string {
    // sprite sheet layout: 8 directions x 5 frames (1 idle + 4 walk)
    const cols = 5; // frames per direction
    const rows = 8; // directions: N, NE, E, SE, S, SW, W, NW

    this.canvas.width = this.frameWidth * cols;
    this.canvas.height = this.frameHeight * rows;

    // clear canvas
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

    // generate frames for each direction
    const directions = ['S', 'SW', 'W', 'NW', 'N', 'NE', 'E', 'SE'];

    directions.forEach((dir, dirIndex) => {
      // generate idle frame
      this.drawAstronaut(0, dirIndex, dir, 0);

      // generate 4 walking frames
      for (let frame = 1; frame <= 4; frame++) {
        this.drawAstronaut(frame, dirIndex, dir, frame);
      }
    });

    return this.canvas.toDataURL();
  }

  private drawAstronaut(col: number, row: number, direction: string, animFrame: number) {
    const x = col * this.frameWidth + this.frameWidth / 2;
    const y = row * this.frameHeight + this.frameHeight / 2;

    this.ctx.save();
    this.ctx.translate(x, y);

    // rotation based on direction
    let rotation = 0;
    switch(direction) {
      case 'N': rotation = -Math.PI/2; break;
      case 'NE': rotation = -Math.PI/4; break;
      case 'E': rotation = 0; break;
      case 'SE': rotation = Math.PI/4; break;
      case 'S': rotation = Math.PI/2; break;
      case 'SW': rotation = 3*Math.PI/4; break;
      case 'W': rotation = Math.PI; break;
      case 'NW': rotation = -3*Math.PI/4; break;
    }

    this.ctx.rotate(rotation);

    // walking animation offset
    const walkOffset = animFrame > 0 ? Math.sin(animFrame * Math.PI / 2) * 2 : 0;

    // visor glass gradient
    const visorGradient = this.ctx.createRadialGradient(0, -2, 0, 0, -2, 8);
    visorGradient.addColorStop(0, 'rgba(100, 200, 255, 0.8)');
    visorGradient.addColorStop(0.5, 'rgba(50, 150, 220, 0.6)');
    visorGradient.addColorStop(1, 'rgba(20, 100, 180, 0.4)');

    // body (spacesuit)
    this.ctx.fillStyle = '#e8e8e8';
    this.ctx.beginPath();
    this.ctx.ellipse(0, 2 + walkOffset * 0.5, 10, 12, 0, 0, Math.PI * 2);
    this.ctx.fill();

    // body details/lines
    this.ctx.strokeStyle = '#b0b0b0';
    this.ctx.lineWidth = 1;
    this.ctx.beginPath();
    this.ctx.moveTo(-6, 0);
    this.ctx.lineTo(-6, 10);
    this.ctx.moveTo(6, 0);
    this.ctx.lineTo(6, 10);
    this.ctx.stroke();

    // backpack
    this.ctx.fillStyle = '#606060';
    this.ctx.fillRect(-12, -2, 4, 8);

    // helmet
    this.ctx.fillStyle = '#f0f0f0';
    this.ctx.beginPath();
    this.ctx.arc(0, -3 + walkOffset * 0.3, 9, 0, Math.PI * 2);
    this.ctx.fill();

    // helmet outline
    this.ctx.strokeStyle = '#c0c0c0';
    this.ctx.lineWidth = 1.5;
    this.ctx.stroke();

    // visor
    this.ctx.fillStyle = visorGradient;
    this.ctx.beginPath();
    this.ctx.arc(0, -2, 7, 0, Math.PI * 2);
    this.ctx.fill();

    // visor reflection
    this.ctx.fillStyle = 'rgba(255, 255, 255, 0.4)';
    this.ctx.beginPath();
    this.ctx.ellipse(2, -4, 3, 2, Math.PI / 4, 0, Math.PI * 2);
    this.ctx.fill();

    // arms (with walking animation)
    const armSwing = animFrame > 0 ? Math.sin(animFrame * Math.PI / 2) * 0.3 : 0;

    this.ctx.strokeStyle = '#e8e8e8';
    this.ctx.lineWidth = 3;
    this.ctx.lineCap = 'round';

    // left arm
    this.ctx.beginPath();
    this.ctx.moveTo(-8, 0);
    this.ctx.lineTo(-10 - armSwing * 4, 6 + Math.abs(armSwing) * 2);
    this.ctx.stroke();

    // right arm
    this.ctx.beginPath();
    this.ctx.moveTo(8, 0);
    this.ctx.lineTo(10 + armSwing * 4, 6 - Math.abs(armSwing) * 2);
    this.ctx.stroke();

    // legs (with walking animation)
    const legSwing = animFrame > 0 ? Math.sin(animFrame * Math.PI) * 4 : 0;

    // left leg
    this.ctx.beginPath();
    this.ctx.moveTo(-4, 10);
    this.ctx.lineTo(-4 + legSwing, 14);
    this.ctx.stroke();

    // right leg
    this.ctx.beginPath();
    this.ctx.moveTo(4, 10);
    this.ctx.lineTo(4 - legSwing, 14);
    this.ctx.stroke();

    this.ctx.restore();
  }

  generateOtherPlayerSpriteSheet(): string {
    // similar to main player but with different colors
    const cols = 5;
    const rows = 8;

    this.canvas.width = this.frameWidth * cols;
    this.canvas.height = this.frameHeight * rows;

    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

    const directions = ['S', 'SW', 'W', 'NW', 'N', 'NE', 'E', 'SE'];

    directions.forEach((dir, dirIndex) => {
      this.drawOtherAstronaut(0, dirIndex, dir, 0);
      for (let frame = 1; frame <= 4; frame++) {
        this.drawOtherAstronaut(frame, dirIndex, dir, frame);
      }
    });

    return this.canvas.toDataURL();
  }

  private drawOtherAstronaut(col: number, row: number, direction: string, animFrame: number) {
    const x = col * this.frameWidth + this.frameWidth / 2;
    const y = row * this.frameHeight + this.frameHeight / 2;

    this.ctx.save();
    this.ctx.translate(x, y);

    let rotation = 0;
    switch(direction) {
      case 'N': rotation = -Math.PI/2; break;
      case 'NE': rotation = -Math.PI/4; break;
      case 'E': rotation = 0; break;
      case 'SE': rotation = Math.PI/4; break;
      case 'S': rotation = Math.PI/2; break;
      case 'SW': rotation = 3*Math.PI/4; break;
      case 'W': rotation = Math.PI; break;
      case 'NW': rotation = -3*Math.PI/4; break;
    }

    this.ctx.rotate(rotation);

    const walkOffset = animFrame > 0 ? Math.sin(animFrame * Math.PI / 2) * 2 : 0;

    // red tinted visor for other players
    const visorGradient = this.ctx.createRadialGradient(0, -2, 0, 0, -2, 8);
    visorGradient.addColorStop(0, 'rgba(255, 100, 100, 0.8)');
    visorGradient.addColorStop(0.5, 'rgba(220, 50, 50, 0.6)');
    visorGradient.addColorStop(1, 'rgba(180, 20, 20, 0.4)');

    // body in darker shade
    this.ctx.fillStyle = '#d0d0d0';
    this.ctx.beginPath();
    this.ctx.ellipse(0, 2 + walkOffset * 0.5, 10, 12, 0, 0, Math.PI * 2);
    this.ctx.fill();

    this.ctx.strokeStyle = '#909090';
    this.ctx.lineWidth = 1;
    this.ctx.beginPath();
    this.ctx.moveTo(-6, 0);
    this.ctx.lineTo(-6, 10);
    this.ctx.moveTo(6, 0);
    this.ctx.lineTo(6, 10);
    this.ctx.stroke();

    this.ctx.fillStyle = '#505050';
    this.ctx.fillRect(-12, -2, 4, 8);

    this.ctx.fillStyle = '#e0e0e0';
    this.ctx.beginPath();
    this.ctx.arc(0, -3 + walkOffset * 0.3, 9, 0, Math.PI * 2);
    this.ctx.fill();

    this.ctx.strokeStyle = '#a0a0a0';
    this.ctx.lineWidth = 1.5;
    this.ctx.stroke();

    this.ctx.fillStyle = visorGradient;
    this.ctx.beginPath();
    this.ctx.arc(0, -2, 7, 0, Math.PI * 2);
    this.ctx.fill();

    this.ctx.fillStyle = 'rgba(255, 255, 255, 0.3)';
    this.ctx.beginPath();
    this.ctx.ellipse(2, -4, 3, 2, Math.PI / 4, 0, Math.PI * 2);
    this.ctx.fill();

    const armSwing = animFrame > 0 ? Math.sin(animFrame * Math.PI / 2) * 0.3 : 0;

    this.ctx.strokeStyle = '#d0d0d0';
    this.ctx.lineWidth = 3;
    this.ctx.lineCap = 'round';

    this.ctx.beginPath();
    this.ctx.moveTo(-8, 0);
    this.ctx.lineTo(-10 - armSwing * 4, 6 + Math.abs(armSwing) * 2);
    this.ctx.stroke();

    this.ctx.beginPath();
    this.ctx.moveTo(8, 0);
    this.ctx.lineTo(10 + armSwing * 4, 6 - Math.abs(armSwing) * 2);
    this.ctx.stroke();

    const legSwing = animFrame > 0 ? Math.sin(animFrame * Math.PI) * 4 : 0;

    this.ctx.beginPath();
    this.ctx.moveTo(-4, 10);
    this.ctx.lineTo(-4 + legSwing, 14);
    this.ctx.stroke();

    this.ctx.beginPath();
    this.ctx.moveTo(4, 10);
    this.ctx.lineTo(4 - legSwing, 14);
    this.ctx.stroke();

    this.ctx.restore();
  }
}