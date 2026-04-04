-- Rollback Migration 007: Remove equipment slots from player_loadouts

-- Drop added columns
ALTER TABLE player_loadouts
    DROP COLUMN IF EXISTS head_instance_id,
    DROP COLUMN IF EXISTS hands_instance_id,
    DROP COLUMN IF EXISTS feet_instance_id,
    DROP COLUMN IF EXISTS ring_1_instance_id,
    DROP COLUMN IF EXISTS ring_2_instance_id,
    DROP COLUMN IF EXISTS consumable_3_id;

-- Rename body_instance_id back to armor_instance_id
ALTER TABLE player_loadouts RENAME COLUMN body_instance_id TO armor_instance_id;
