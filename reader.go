package scheduleparser

import (
	"bytes"

	"github.com/ledongthuc/pdf"
)

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
