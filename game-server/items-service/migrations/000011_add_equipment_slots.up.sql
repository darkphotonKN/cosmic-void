-- Migration 007: Add equipment slots to player_loadouts

-- Rename armor_instance_id to body_instance_id
ALTER TABLE player_loadouts RENAME COLUMN armor_instance_id TO body_instance_id;

-- Add new equipment slot columns
ALTER TABLE player_loadouts
    ADD COLUMN head_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    ADD COLUMN hands_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    ADD COLUMN feet_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    ADD COLUMN ring_1_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    ADD COLUMN ring_2_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    ADD COLUMN consumable_3_id UUID REFERENCES item_instances(id) ON DELETE SET NULL;
