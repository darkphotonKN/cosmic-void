interface Position {
  x: number;
  y: number;
}

interface Treasure {
  id: string;
  x: number;
  y: number;
  type: "gold" | "silver";
  value: number;
}

interface Enemy {
  id: string;
  x: number;
  y: number;
  type: "goblin" | "skeleton";
  hp: number;
}

interface Player {
  id: string;
  x: number;
  y: number;
  hp: number;
  name: string;
}

interface Wall {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

interface Building {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  type: "house" | "tower" | "ruins" | "shrine";
}
