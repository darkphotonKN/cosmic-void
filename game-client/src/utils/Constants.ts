export const GAME_CONFIG = {
  WIDTH: 900,
  HEIGHT: 600,
  PLAYER_SPEED: 200,
  BULLET_SPEED: 400,
  // Treasure Hunt specific
  MAP_WIDTH: 3000,
  MAP_HEIGHT: 3000,
  VIEW_RADIUS: 250,
  UPDATE_THRESHOLD: 10,
  ATTACK_RANGE: 60,
  COLLECT_RANGE: 50,
} as const;

export const SCENES = {
  BOOT: 'BootScene',
  PRELOAD: 'PreloadScene',
  MAIN_MENU: 'MainMenuScene',
  GAME: 'GameScene',
  GAME_OVER: 'GameOverScene',
  TREASURE_HUNT: 'TreasureHuntScene',
} as const;

export const ASSET_KEYS = {
  IMAGES: {
    PLAYER: 'player',
    ENEMY: 'enemy',
    BULLET: 'bullet',
    BACKGROUND: 'background',
    // Treasure Hunt specific
    TREASURE_PLAYER: 'treasurePlayer',
    OTHER_TREASURE_PLAYER: 'otherTreasurePlayer',
    GOLD_CHEST: 'goldChest',
    SILVER_CHEST: 'silverChest',
    SKELETON: 'skeleton',
    GOBLIN: 'goblin',
  },
  AUDIO: {
    BGM: 'bgm',
    SHOOT: 'shoot',
    EXPLOSION: 'explosion',
  },
} as const;

export const BUILDING_TYPES = {
  HOUSE: 'house',
  TOWER: 'tower',
  RUINS: 'ruins',
  SHRINE: 'shrine',
} as const;

export const BUILDING_COLORS = {
  FLOOR: {
    house: 0x8b7355,
    tower: 0x606060,
    ruins: 0x4a5240,
    shrine: 0xdaa520,
  },
  WALL: {
    house: 0x654321,
    tower: 0x505050,
    ruins: 0x3d4235,
    shrine: 0xb8860b,
  },
  ROOF: {
    house: 0x8b4513,
    tower: 0x404040,
    ruins: 0x2d3228,
    shrine: 0xffd700,
  },
} as const;
