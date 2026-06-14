ALTER TABLE mandados ADD COLUMN motivo_nao_realizacao_id BIGINT REFERENCES motivos_nao_realizacao(id) ON DELETE SET NULL;
