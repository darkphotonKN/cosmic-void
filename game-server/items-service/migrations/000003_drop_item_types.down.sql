-- Rollback Migration 003: Recreate item_types table and restore references

-- 1. Recreate item_types table
CREATE TABLE item_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,

    CONSTRAINT non_negative_sort_order CHECK (sort_order >= 0)
);

-- 2. Re-add type_id to weapons table
ALTER TABLE weapons ADD COLUMN type_id UUID;
ALTER TABLE weapons ADD CONSTRAINT weapons_type_id_fkey FOREIGN KEY (type_id) REFERENCES item_types(id);
CREATE INDEX idx_weapons_type ON weapons(type_id);

-- 3. Re-add type_id to armors table
ALTER TABLE armors ADD COLUMN type_id UUID;
ALTER TABLE armors ADD CONSTRAINT armors_type_id_fkey FOREIGN KEY (type_id) REFERENCES item_types(id);
CREATE INDEX idx_armors_type ON armors(type_id);

-- 4. Re-add type_id to consumables table
ALTER TABLE consumables ADD COLUMN type_id UUID;
ALTER TABLE consumables ADD CONSTRAINT consumables_type_id_fkey FOREIGN KEY (type_id) REFERENCES item_types(id);
CREATE INDEX idx_consumables_type ON consumables(type_id);

-- 5. Re-add type_id to item_templates table
ALTER TABLE item_templates ADD COLUMN type_id UUID;
ALTER TABLE item_templates ADD CONSTRAINT item_templates_type_id_fkey FOREIGN KEY (type_id) REFERENCES item_types(id);
CREATE INDEX idx_item_templates_type ON item_templates(type_id);

-- 6. Recreate indexes
CREATE INDEX idx_item_types_active ON item_types(is_active);
CREATE INDEX idx_item_types_sort ON item_types(sort_order);
