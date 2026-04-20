CREATE TABLE IF NOT EXISTS mandados (
    id               BIGSERIAL PRIMARY KEY,
    numero           VARCHAR(100) UNIQUE NOT NULL,
    nome             VARCHAR(255),
    documento        VARCHAR(30),
    sexo             VARCHAR(20),
    posicao          VARCHAR(100),
    endereco         TEXT,
    cidade           VARCHAR(150),
    situacao         VARCHAR(50),
    data_carga       VARCHAR(50),
    diligencias      TEXT,
    is_pj            BOOLEAN NOT NULL DEFAULT FALSE,
    repr_nome        VARCHAR(255),
    repr_doc         VARCHAR(30),
    obs              TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS atos (
    id               BIGSERIAL PRIMARY KEY,
    mandado_id       BIGINT NOT NULL REFERENCES mandados(id) ON DELETE CASCADE,
    nome_ato         VARCHAR(255) NOT NULL,
    data_cumprimento VARCHAR(50),
    horario          VARCHAR(20),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contatos (
    id               BIGSERIAL PRIMARY KEY,
    mandado_id       BIGINT NOT NULL REFERENCES mandados(id) ON DELETE CASCADE,
    tipo             VARCHAR(20) NOT NULL,
    valor            VARCHAR(255),
    is_terceiro      BOOLEAN NOT NULL DEFAULT FALSE,
    nome_terceiro    VARCHAR(255),
    relacao_terceiro VARCHAR(100),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
