CREATE TABLE IF NOT EXISTS motivos_nao_realizacao (
    id BIGSERIAL PRIMARY KEY,
    codigo VARCHAR(50) UNIQUE NOT NULL,
    explicacao TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO motivos_nao_realizacao (codigo, explicacao) VALUES
('MUD', 'Mudou-se'),
('END_INSUF', 'Endereço Insuficiente'),
('DESCONHECIDO', 'Desconhecido no local'),
('AUSENTE', 'Ausente'),
('RECUSA', 'Recusa no recebimento'),
('FALECIDO', 'Falecido'),
('FORA_JURISD', 'Fora da jurisdição')
ON CONFLICT (codigo) DO NOTHING;
