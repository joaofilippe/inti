package service

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/joaofilippe/inti/internal/document"
	"github.com/joaofilippe/inti/internal/domain/entities"
)

func TestDocumentoFormatado(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CPF 11 dígitos sem formatação",
			input:    "12345678901",
			expected: "12345678901",
		},
		{
			name:     "CPF com pontos e traço",
			input:    "123.456.789-01",
			expected: "12345678901",
		},
		{
			name:     "CNPJ com pontuação",
			input:    "12.345.678/0001-90",
			expected: "12345678000190",
		},
		{
			name:     "RG com pontos e traço",
			input:    "12.345.678-9",
			expected: "123456789",
		},
		{
			name:     "RG com dígito X",
			input:    "12.345.678-x",
			expected: "12345678X",
		},
		{
			name:     "Espaços e caracteres especiais removidos",
			input:    " (12) 345-678/90 ",
			expected: "1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := documentoFormatado(tt.input)
			if got != tt.expected {
				t.Errorf("documentoFormatado(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildReplaces_Documento(t *testing.T) {
	svc := NewMandadoService(nil, nil)

	t.Run("Pessoa Física não deve incluir tipo nem pontuação", func(t *testing.T) {
		m := entities.Mandado{
			Mandado:       "100.2026/000001-0",
			Nome:          "Fulano da Silva",
			Sexo:          "M",
			Posicao:       "réu",
			Documento:     "123.456.789-01",
			TipoDocumento: "CPF",
		}

		replaces := svc.BuildReplaces(m)

		expectedDoc := "12345678901"
		if replaces["{{DOCUMENTO}}"] != expectedDoc {
			t.Errorf("expected {{DOCUMENTO}} = %q, got %q", expectedDoc, replaces["{{DOCUMENTO}}"])
		}
		if replaces["{{CPF}}"] != expectedDoc {
			t.Errorf("expected {{CPF}} = %q, got %q", expectedDoc, replaces["{{CPF}}"])
		}
		if strings.Contains(replaces["{{DOCUMENTO}}"], "CPF") {
			t.Errorf("{{DOCUMENTO}} should not contain 'CPF', got %q", replaces["{{DOCUMENTO}}"])
		}
		if strings.ContainsAny(replaces["{{DOCUMENTO}}"], ".-/") {
			t.Errorf("{{DOCUMENTO}} should not contain punctuation, got %q", replaces["{{DOCUMENTO}}"])
		}
	})

	t.Run("Pessoa Jurídica sem pontuação nos documentos", func(t *testing.T) {
		m := entities.Mandado{
			Mandado:           "100.2026/000002-0",
			Nome:              "Empresa XYZ Ltda",
			IsPJ:              true,
			Documento:         "12.345.678/0001-90",
			RepresentanteNome: "Beltrano",
			RepresentanteDoc:  "987.654.321-00",
		}

		replaces := svc.BuildReplaces(m)

		expectedPJDoc := "12345678000190 CPF 98765432100"
		if replaces["{{DOCUMENTO}}"] != expectedPJDoc {
			t.Errorf("expected {{DOCUMENTO}} = %q, got %q", expectedPJDoc, replaces["{{DOCUMENTO}}"])
		}
		if replaces["{{CPF}}"] != expectedPJDoc {
			t.Errorf("expected {{CPF}} = %q, got %q", expectedPJDoc, replaces["{{CPF}}"])
		}
	})
}

func TestReplaceInDocx_2INT(t *testing.T) {
	svc := NewMandadoService(nil, nil)
	m := entities.Mandado{
		Mandado:       "100.2026/000001-0",
		DataCarga:     "12/09/2026",
		Cidade:        "São Paulo",
		Endereco:      "Rua das Flores, 123",
		Nome:          "Fulano da Silva",
		Sexo:          "M",
		Posicao:       "réu",
		Documento:     "123.456.789-01",
		TipoDocumento: "CPF",
	}

	replaces := svc.BuildReplaces(m)

	buf, err := document.ReplaceInDocx("../../../2INT.docx", replaces)
	if err != nil {
		t.Fatalf("ReplaceInDocx failed: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader failed: %v", err)
	}

	var docXML string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Open failed: %v", err)
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}
			docXML = string(b)
			break
		}
	}

	if strings.Contains(docXML, "{{DOCUMENTO}}") {
		t.Errorf("tag {{DOCUMENTO}} was not replaced in docx output!")
	}
	if !strings.Contains(docXML, "12345678901") {
		t.Errorf("clean document '12345678901' was not found in docx output!")
	}
	if strings.Contains(docXML, "123.456.789-01") {
		t.Errorf("document in docx output should not have punctuation!")
	}
	if strings.Contains(docXML, "CPF") {
		t.Errorf("document in docx output should not have 'CPF' prefix for PF!")
	}
}
