CREATE TABLE IF NOT EXISTS motivos_nao_realizacao (
    id BIGSERIAL PRIMARY KEY,
    codigo VARCHAR(50) UNIQUE NOT NULL,
    explicacao TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO motivos_nao_realizacao (codigo, explicacao) VALUES
('MUD', 'Mudou-se'),
('EI', 'Endereço Insuficiente'),
('DESC', 'Desconhecido no local'),
('AUS', 'Ausente'),
('REC', 'Recusa no recebimento'),
('FAL', 'Falecido'),
('IF', 'Imóvel fechado'),
('NNE', 'Número não existe'),
('FL', 'Familiar no local'),
('FIF', 'Familiar no local, informa falecimento'),
('FIM', 'Familiar no local, informa mudança'),
('FDNE', 'Familiar no local, desconhece novo endereço'),
('DMHT', 'Desconhecido mora há algum tempo no local'),
('FJJ', 'Familiar no local, informa que o executado mora fora da jurisdição'),
('TSC', 'Terreno sem construções'),
('TCC', 'Terreno com construções'),
('ICC', 'Imóvel em construção'),
('IE', 'Imóvel em ruínas'),
('IEA', 'Imóvel em construção abandonado'),
('IER', 'Imóvel em reformas'),
('OU', 'Outros')
ON CONFLICT (codigo) DO NOTHING;
