CREATE TABLE IF NOT EXISTS atos (
    id               BIGSERIAL PRIMARY KEY,
    mandado_id       BIGINT NOT NULL REFERENCES mandados(id) ON DELETE CASCADE,
    tipo_ato_codigo  INT NOT NULL REFERENCES tipos_ato(codigo),
    data_cumprimento VARCHAR(50),
    horario          VARCHAR(20),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO atos (mandado_id, tipo_ato_codigo, data_cumprimento, horario)
SELECT
    m.id,
    (a->>'codigo_ato')::INT,
    a->>'data_cumprimento',
    a->>'horario'
FROM mandados m,
     jsonb_array_elements(m.atos) AS a
WHERE jsonb_array_length(m.atos) > 0;

ALTER TABLE mandados DROP COLUMN atos;
