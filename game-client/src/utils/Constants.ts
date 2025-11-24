export const GAME_CONFIG = {
  WIDTH: 800,
  HEIGHT: 600,
  PLAYER_SPEED: 200,
  BULLET_SPEED: 400,
} as const;

export const SCENES = {
  BOOT: 'BootScene',
  PRELOAD: 'PreloadScene',
  MAIN_MENU: 'MainMenuScene',
  GAME: 'GameScene',
  GAME_OVER: 'GameOverScene',
} as const;

export const ASSET_KEYS = {
  IMAGES: {
    PLAYER: 'player',
    ENEMY: 'enemy',
    BULLET: 'bullet',
    BACKGROUND: 'background',
  },
  AUDIO: {
    BGM: 'bgm',
    SHOOT: 'shoot',
    EXPLOSION: 'explosion',
  },
} as const;
