package document

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
)

// ReplaceInDocx abre um template .docx (ZIP), substitui as tags no XML interno
// e retorna o documento modificado em memória como *bytes.Buffer.
func ReplaceInDocx(templatePath string, replaces map[string]string) (*bytes.Buffer, error) {
	zr, err := zip.OpenReader(templatePath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, f := range zr.File {
		w, err := zw.Create(f.Name)
		if err != nil {
			return nil, err
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		if f.Name == "word/document.xml" || f.Name == "word/header1.xml" || f.Name == "word/footer1.xml" {
			s := string(b)
			for tag, val := range replaces {
				s = strings.ReplaceAll(s, tag, val)
			}
			b = []byte(s)
		}

		if _, err = w.Write(b); err != nil {
			return nil, err
		}
	}

	if err = zw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}
