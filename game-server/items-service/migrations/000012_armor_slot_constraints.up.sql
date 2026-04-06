-- Migration 012: Add armor_slot CHECK constraints + rename loadout columns to match
--
-- Why:
--   - armor_slot in `armors` and `item_instances` is currently free-text VARCHAR(50)
--     with no CHECK constraint, allowing typos and invalid values.
--   - player_loadouts column names diverged from armor_slot values (body vs chest,
--     feet vs legs, hands vs gloves), forcing translation everywhere.
--   - item_type and source already use CHECK constraints in the same table —
--     armor_slot was just missed.

-- Step 1: Rename player_loadouts columns to match canonical armor_slot values
ALTER TABLE player_loadouts RENAME COLUMN body_instance_id  TO chest_instance_id;
ALTER TABLE player_loadouts RENAME COLUMN feet_instance_id  TO legs_instance_id;
ALTER TABLE player_loadouts RENAME COLUMN hands_instance_id TO gloves_instance_id;
-- head_instance_id, ring_1_instance_id, ring_2_instance_id are unchanged

-- Step 2: Normalize any existing non-canonical armor_slot data on `armors`
UPDATE armors SET armor_slot = 'chest'  WHERE armor_slot = 'body';
UPDATE armors SET armor_slot = 'legs'   WHERE armor_slot = 'feet';
UPDATE armors SET armor_slot = 'gloves' WHERE armor_slot = 'hands';

-- Step 3: Normalize any existing non-canonical armor_slot data on `item_instances`
UPDATE item_instances SET armor_slot = 'chest'  WHERE armor_slot = 'body';
UPDATE item_instances SET armor_slot = 'legs'   WHERE armor_slot = 'feet';
UPDATE item_instances SET armor_slot = 'gloves' WHERE armor_slot = 'hands';

-- Step 4: CHECK constraint on `armors` (always an armor row, so simple set check)
ALTER TABLE armors
    ADD CONSTRAINT armors_armor_slot_check
    CHECK (armor_slot IN ('head', 'chest', 'legs', 'gloves'));

-- Step 5: Conditional CHECK constraint on `item_instances`
--   - When item_type = 'armor': armor_slot must be in the valid set
--   - When item_type != 'armor': armor_slot must be NULL (it's irrelevant)
ALTER TABLE item_instances
    ADD CONSTRAINT item_instances_armor_slot_check
    CHECK (
        (item_type = 'armor' AND armor_slot IN ('head', 'chest', 'legs', 'gloves'))
        OR
        (item_type <> 'armor' AND armor_slot IS NULL)
    );
