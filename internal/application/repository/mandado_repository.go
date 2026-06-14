package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/joaofilippe/inti/internal/api/dto"
	"github.com/joaofilippe/inti/internal/domain/entities"
)

type MandadoRepository struct {
	db *sqlx.DB
}

func NewMandadoRepository(db *sqlx.DB) *MandadoRepository {
	return &MandadoRepository{db: db}
}

// mandadoExtraidoDB é o modelo de banco para um mandado extraído via IA.
type mandadoExtraidoDB struct {
	Numero         string `db:"numero"`
	Lote           string `db:"lote"`
	NumeroProcesso string `db:"numero_processo"`
	Nome           string `db:"nome"`
	Documento      string `db:"documento"`
	TipoDocumento  string `db:"tipo_documento"`
	Sexo           string `db:"sexo"`
	Posicao        string `db:"posicao"`
	Endereco       string `db:"endereco"`
	Cidade         string `db:"cidade"`
}

// mandadoDB é o modelo de banco para um mandado completo.
type mandadoDB struct {
	Numero        string `db:"numero"`
	Lote          string `db:"lote"`
	Nome          string `db:"nome"`
	Documento     string `db:"documento"`
	TipoDocumento string `db:"tipo_documento"`
	Sexo          string `db:"sexo"`
	Posicao       string `db:"posicao"`
	Endereco      string `db:"endereco"`
	Cidade        string `db:"cidade"`
	Situacao      string `db:"situacao"`
	DataCarga     string `db:"data_carga"`
	Diligencias   string `db:"diligencias"`
	IsPJ                  bool   `db:"is_pj"`
	ReprNome              string `db:"repr_nome"`
	ReprDoc               string `db:"repr_doc"`
	Obs                   string `db:"obs"`
	Atos                  []byte `db:"atos"`
	MotivoNaoRealizacaoID *int   `db:"motivo_nao_realizacao_id"`
}

type contadoDB struct {
	MandadoID       int64  `db:"mandado_id"`
	Tipo            string `db:"tipo"`
	Valor           string `db:"valor"`
	IsTerceiro      bool   `db:"is_terceiro"`
	NomeTerceiro    string `db:"nome_terceiro"`
	RelacaoTerceiro string `db:"relacao_terceiro"`
}

// SalvarExtraido insere ou atualiza um mandado extraído pelo número.
func (r *MandadoRepository) SalvarExtraido(ctx context.Context, m *dto.MandadoExtraido) error {
	row := mandadoExtraidoDB{
		Numero:         m.Mandado,
		Lote:           m.Lote,
		NumeroProcesso: m.NumeroProcesso,
		Nome:           m.Nome,
		Documento:      m.Documento,
		TipoDocumento:  m.TipoDocumento,
		Sexo:           m.Sexo,
		Posicao:        m.Posicao,
		Endereco:       m.Endereco,
		Cidade:         m.Cidade,
	}

	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO mandados (numero, lote, numero_processo, nome, documento, tipo_documento, sexo, posicao, endereco, cidade, updated_at)
		VALUES (:numero, :lote, :numero_processo, :nome, :documento, :tipo_documento, :sexo, :posicao, :endereco, :cidade, NOW())
		ON CONFLICT (numero) DO UPDATE SET
			lote            = EXCLUDED.lote,
			numero_processo = EXCLUDED.numero_processo,
			nome            = EXCLUDED.nome,
			documento       = EXCLUDED.documento,
			tipo_documento  = EXCLUDED.tipo_documento,
			sexo            = EXCLUDED.sexo,
			posicao         = EXCLUDED.posicao,
			endereco        = EXCLUDED.endereco,
			cidade          = EXCLUDED.cidade,
			updated_at      = NOW()
	`, row)
	if err != nil {
		return fmt.Errorf("erro ao salvar mandado extraído: %w", err)
	}
	return nil
}

// SalvarLoteExtraido insere ou atualiza uma lista de mandados extraídos.
func (r *MandadoRepository) SalvarLoteExtraido(ctx context.Context, lista []dto.MandadoExtraido) error {
	for i := range lista {
		if err := r.SalvarExtraido(ctx, &lista[i]); err != nil {
			return fmt.Errorf("mandado %q: %w", lista[i].Mandado, err)
		}
	}
	return nil
}

// SalvarMandado insere ou atualiza um mandado completo com atos e contatos.
func (r *MandadoRepository) SalvarMandado(ctx context.Context, m *entities.Mandado) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	row := mandadoDB{
		Numero:                m.Mandado,
		Lote:                  m.Lote,
		Nome:                  m.Nome,
		Documento:             m.Documento,
		TipoDocumento:         m.TipoDocumento,
		Sexo:                  m.Sexo,
		Posicao:               m.Posicao,
		Endereco:              m.Endereco,
		Cidade:                m.Cidade,
		Situacao:              m.Situacao,
		DataCarga:             m.DataCarga,
		Diligencias:           m.Diligencias,
		IsPJ:                  m.IsPJ,
		ReprNome:              m.RepresentanteNome,
		ReprDoc:               m.RepresentanteDoc,
		Obs:                   m.Obs,
		MotivoNaoRealizacaoID: m.MotivoNaoRealizacaoID,
	}

	atosJSON, err := json.Marshal(m.Atos)
	if err != nil {
		return fmt.Errorf("erro ao serializar atos: %w", err)
	}
	row.Atos = atosJSON

	rows, err := tx.NamedQuery(`
		INSERT INTO mandados (numero, lote, nome, documento, tipo_documento, sexo, posicao, endereco, cidade,
		                      situacao, data_carga, diligencias, is_pj, repr_nome, repr_doc, obs, atos, motivo_nao_realizacao_id, updated_at)
		VALUES (:numero, :lote, :nome, :documento, :tipo_documento, :sexo, :posicao, :endereco, :cidade,
		        :situacao, :data_carga, :diligencias, :is_pj, :repr_nome, :repr_doc, :obs, :atos, :motivo_nao_realizacao_id, NOW())
		ON CONFLICT (numero) DO UPDATE SET
			lote           = EXCLUDED.lote,
			nome           = EXCLUDED.nome,
			documento      = EXCLUDED.documento,
			tipo_documento = EXCLUDED.tipo_documento,
			sexo           = EXCLUDED.sexo,
			posicao        = EXCLUDED.posicao,
			endereco       = EXCLUDED.endereco,
			cidade         = EXCLUDED.cidade,
			situacao       = EXCLUDED.situacao,
			data_carga     = EXCLUDED.data_carga,
			diligencias    = EXCLUDED.diligencias,
			is_pj          = EXCLUDED.is_pj,
			repr_nome      = EXCLUDED.repr_nome,
			repr_doc       = EXCLUDED.repr_doc,
			obs            = EXCLUDED.obs,
			atos           = EXCLUDED.atos,
			motivo_nao_realizacao_id = EXCLUDED.motivo_nao_realizacao_id,
			updated_at     = NOW()
		RETURNING id
	`, row)
	if err != nil {
		return fmt.Errorf("erro ao salvar mandado: %w", err)
	}
	defer rows.Close()
	var id int64
	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("erro ao ler id do mandado: %w", err)
		}
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM contatos WHERE mandado_id = $1`, id); err != nil {
		return fmt.Errorf("erro ao limpar contatos: %w", err)
	}
	for _, c := range m.Contatos {
		if _, err = tx.NamedExecContext(ctx, `
			INSERT INTO contatos (mandado_id, tipo, valor, is_terceiro, nome_terceiro, relacao_terceiro)
			VALUES (:mandado_id, :tipo, :valor, :is_terceiro, :nome_terceiro, :relacao_terceiro)
		`, contadoDB{
			MandadoID:       id,
			Tipo:            string(c.Tipo),
			Valor:           c.Valor,
			IsTerceiro:      c.IsTerceiro,
			NomeTerceiro:    c.NomeTerceiro,
			RelacaoTerceiro: c.RelacaoTerceiro,
		}); err != nil {
			return fmt.Errorf("erro ao salvar contato: %w", err)
		}
	}

	return tx.Commit()
}
