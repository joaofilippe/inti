ALTER TABLE mandados ADD COLUMN atos JSONB NOT NULL DEFAULT '[]';

UPDATE mandados
SET atos = COALESCE(
    (SELECT json_agg(
        json_build_object(
            'codigo_ato', a.tipo_ato_codigo,
            'data_cumprimento', a.data_cumprimento,
            'horario', a.horario
        ) ORDER BY a.id
    )
    FROM atos a
    WHERE a.mandado_id = mandados.id),
    '[]'::json
);

DROP TABLE atos;
