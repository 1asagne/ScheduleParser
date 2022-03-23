// Package scheduleparser provides structs and functions for parsing pdf schedules in a specific format to json.

package scheduleparser

import (
	"bytes"

	"github.com/ledongthuc/pdf"
)

// readPdfFile opens file on specified path, reads content and returns slice of pdf.Text.
func readPdfFile(filePath string) ([]pdf.Text, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	page := reader.Page(1)
	texts := page.Content().Text
	return texts, nil
}

// readPdfBytes reads specified file bytes and returns slice of pdf.Text.
func readPdfBytes(fileBytes []byte) ([]pdf.Text, error) {
	bytesReader := bytes.NewReader(fileBytes)
	pdfReader, err := pdf.NewReader(bytesReader, int64(len(fileBytes)))
	if err != nil {
		return nil, err
	}
	page := pdfReader.Page(1)
	texts := page.Content().Text
	return texts, nil
}
