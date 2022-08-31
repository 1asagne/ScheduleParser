// Package reader provides functions for reading pdf files.

package reader

import (
	"bytes"
	"io"
	"os"

	"github.com/ledongthuc/pdf"
)

// read returns pdf content from reader.
func read(reader io.ReaderAt, size int64) ([]pdf.Text, error) {
	pdfReader, err := pdf.NewReader(reader, size)
	if err != nil {
		return nil, err
	}

	page := pdfReader.Page(1)
	texts := page.Content().Text
	return texts, nil
}

// ReadFile reads file and returns slice of pdf.Text.
func ReadFile(filePath string) ([]pdf.Text, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return read(file, fileInfo.Size())
}

// ReadBytes reads file bytes and returns slice of pdf.Text.
func ReadBytes(fileBytes []byte) ([]pdf.Text, error) {
	reader := bytes.NewReader(fileBytes)
	return read(reader, reader.Size())
}
