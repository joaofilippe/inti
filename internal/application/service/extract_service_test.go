package service

import (
	"testing"

	"github.com/joaofilippe/inti/internal/api/dto"
)

func TestNormalizarExtraido_NovosCampos(t *testing.T) {
	tests := []struct {
		name            string
		input           dto.MandadoExtraido
		expectedEnc     string
		expectedRes     string
		expectedDataC   string
		expectedTipoAto string
		expectedDataCum string
		expectedHoraCum string
	}{
		{
			name: "Normalização com códigos numéricos",
			input: dto.MandadoExtraido{
				Nome:            "FULANO DA SILVA",
				Mandado:         "205.2026/000511-3",
				DataCarga:       " 15/05 ",
				TipoAto:         " 1 ",
				Encontrado:      "2",
				Resultado:       "4",
				DataCumprimento: " 16/05 ",
				HoraCumprimento: " 14:30 ",
			},
			expectedEnc:     "2",
			expectedRes:     "4",
			expectedDataC:   "15/05",
			expectedTipoAto: "1",
			expectedDataCum: "16/05",
			expectedHoraCum: "14:30",
		},
		{
			name: "Normalização com múltiplos atos '1 e 2'",
			input: dto.MandadoExtraido{
				Nome:            "CICLANO DE SOUZA",
				Mandado:         "205.2026/000512-1",
				DataCarga:       "10/05",
				TipoAto:         "1 e 2",
				Encontrado:      "Sim",
				Resultado:       "Positivo",
				DataCumprimento: "12/05",
				HoraCumprimento: "10:00",
			},
			expectedEnc:     "2",
			expectedRes:     "4",
			expectedDataC:   "10/05",
			expectedTipoAto: "1, 2",
			expectedDataCum: "12/05",
			expectedHoraCum: "10:00",
		},
		{
			name: "Normalização com múltiplos atos '1, 2' e negativo",
			input: dto.MandadoExtraido{
				Nome:            "BELTRANO PEREIRA",
				Mandado:         "205.2026/000513-0",
				DataCarga:       "20/05",
				TipoAto:         "1, 2",
				Encontrado:      "não",
				Resultado:       "negativo",
				DataCumprimento: "21/05",
				HoraCumprimento: "16:45",
			},
			expectedEnc:     "3",
			expectedRes:     "6",
			expectedDataC:   "20/05",
			expectedTipoAto: "1, 2",
			expectedDataCum: "21/05",
			expectedHoraCum: "16:45",
		},
		{
			name: "Normalização com múltiplos atos com barra '1/2' e parcial",
			input: dto.MandadoExtraido{
				Nome:            "EMPRESA EXEMPLO LTDA",
				Mandado:         "205.2026/000514-8",
				DataCarga:       "05/06",
				TipoAto:         "1/2",
				Encontrado:      "2 - sim",
				Resultado:       "5 - parcial",
				DataCumprimento: "06/06",
				HoraCumprimento: "09:15",
			},
			expectedEnc:     "2",
			expectedRes:     "5",
			expectedDataC:   "05/06",
			expectedTipoAto: "1, 2",
			expectedDataCum: "06/06",
			expectedHoraCum: "09:15",
		},
		{
			name: "Normalização com código e texto '1 - Citação'",
			input: dto.MandadoExtraido{
				Nome:            "OUTRO ALVO",
				Mandado:         "205.2026/000515-5",
				DataCarga:       "08/06",
				TipoAto:         "1 - Citação",
				Encontrado:      "2",
				Resultado:       "4",
				DataCumprimento: "09/06",
				HoraCumprimento: "11:00",
			},
			expectedEnc:     "2",
			expectedRes:     "4",
			expectedDataC:   "08/06",
			expectedTipoAto: "1",
			expectedDataCum: "09/06",
			expectedHoraCum: "11:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.input
			normalizarExtraido(&m)

			if m.Encontrado != tt.expectedEnc {
				t.Errorf("Encontrado: esperado %q, obtido %q", tt.expectedEnc, m.Encontrado)
			}
			if m.Resultado != tt.expectedRes {
				t.Errorf("Resultado: esperado %q, obtido %q", tt.expectedRes, m.Resultado)
			}
			if m.DataCarga != tt.expectedDataC {
				t.Errorf("DataCarga: esperado %q, obtido %q", tt.expectedDataC, m.DataCarga)
			}
			if m.TipoAto != tt.expectedTipoAto {
				t.Errorf("TipoAto: esperado %q, obtido %q", tt.expectedTipoAto, m.TipoAto)
			}
			if m.DataCumprimento != tt.expectedDataCum {
				t.Errorf("DataCumprimento: esperado %q, obtido %q", tt.expectedDataCum, m.DataCumprimento)
			}
			if m.HoraCumprimento != tt.expectedHoraCum {
				t.Errorf("HoraCumprimento: esperado %q, obtido %q", tt.expectedHoraCum, m.HoraCumprimento)
			}
		})
	}
}
