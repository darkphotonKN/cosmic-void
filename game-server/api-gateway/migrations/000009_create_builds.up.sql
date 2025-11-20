CREATE TABLE IF NOT EXISTS builds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    main_skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
    ascendancy_id UUID REFERENCES ascendancies(id) ON DELETE RESTRICT,
    avg_end_game_rating DECIMAL(3, 1) DEFAULT 0 CHECK (
        avg_end_game_rating >= 0
        AND avg_end_game_rating <= 10
    ),
    avg_fun_rating DECIMAL(3, 1) DEFAULT 0 CHECK (
        avg_fun_rating >= 0
        AND avg_fun_rating <= 10
    ),
    avg_creative_rating DECIMAL(3, 1) DEFAULT 0 CHECK (
        avg_creative_rating >= 0
        AND avg_creative_rating <= 10
    ),
    avg_speed_farm_rating DECIMAL(3, 1) DEFAULT 0 CHECK (
        avg_speed_farm_rating >= 0
        AND avg_speed_farm_rating <= 10
    ),
    avg_bossing_rating DECIMAL(3, 1) DEFAULT 0 CHECK (
        avg_bossing_rating >= 0
        AND avg_bossing_rating <= 10
    ),
    views INT DEFAULT 0 CHECK (views >= 0),
    status SMALLINT NOT NULL DEFAULT 0 CHECK (status IN (0, 1, 2)), -- 0: Draft, 1: Published, 2: Archived
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 