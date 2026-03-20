-- Migration 010 down: Restore durability and type_id columns

ALTER TABLE weapons ADD COLUMN IF NOT EXISTS durability INTEGER NOT NULL DEFAULT 0;
ALTER TABLE armors ADD COLUMN IF NOT EXISTS durability INTEGER NOT NULL DEFAULT 0;

ALTER TABLE weapons ADD COLUMN IF NOT EXISTS type_id UUID;
ALTER TABLE armors ADD COLUMN IF NOT EXISTS type_id UUID;
ALTER TABLE consumables ADD COLUMN IF NOT EXISTS type_id UUID;

CREATE INDEX IF NOT EXISTS idx_weapons_type ON weapons(type_id);
CREATE INDEX IF NOT EXISTS idx_armors_type ON armors(type_id);
CREATE INDEX IF NOT EXISTS idx_consumables_type ON consumables(type_id);
