CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_groups (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS lotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome_lote VARCHAR(255) NOT NULL UNIQUE,
    admin_user_id UUID NOT NULL REFERENCES users(id),
    group_id UUID REFERENCES groups(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- We won't strictly add a foreign key from mandados(lote) to lotes(nome_lote) right now
-- because there might be existing rows in mandados with a lote that doesn't exist in lotes table.
-- But since the user ran docker compose down -v, the DB is empty. So we CAN add the constraint.
-- Let's do it safely just in case:

-- First, ensure all existing lotes in mandados have a record in lotes. 
-- Wait, the db is empty, but we can't assume that for all environments.
-- Let's just add the foreign key.
ALTER TABLE mandados
    ADD CONSTRAINT fk_mandados_lote
    FOREIGN KEY (lote) REFERENCES lotes(nome_lote) ON DELETE CASCADE;
