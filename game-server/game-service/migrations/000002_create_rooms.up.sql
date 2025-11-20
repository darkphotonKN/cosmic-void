CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    creator_id UUID NOT NULL,
    max_players INTEGER NOT NULL DEFAULT 4,
    current_players INTEGER NOT NULL DEFAULT 0,
    game_mode VARCHAR(50) NOT NULL DEFAULT 'classic',
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    x FLOAT NOT NULL DEFAULT 0,
    y FLOAT NOT NULL DEFAULT 0,
    velocity_x FLOAT NOT NULL DEFAULT 0,
    velocity_y FLOAT NOT NULL DEFAULT 0,
    health INTEGER NOT NULL DEFAULT 100,
    score INTEGER NOT NULL DEFAULT 0,
    is_alive BOOLEAN NOT NULL DEFAULT true,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, room_id)
);

CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_game_mode ON rooms(game_mode);
CREATE INDEX idx_players_room_id ON players(room_id);
CREATE INDEX idx_players_user_id ON players(user_id);