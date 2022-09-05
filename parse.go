package scheduleparser

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/qsoulior/scheduleparser/internal/reader"
	"github.com/qsoulior/scheduleparser/pkg/parser"
)

// Event is retrieved from RawEvent. It is contained in output json.
type Event = parser.Event

// getBodyText takes slice of pdf.Text, returns slice of pdf.Text without junk.
func getBodyText(texts []pdf.Text) []pdf.Text {
	mainText := make([]pdf.Text, 0)
	for _, text := range texts {
		if text.Y < 521 && text.X > 42 {
			mainText = append(mainText, text)
		}
	}
	return mainText
}

// parseText takes slice of pdf.Text,
// parses content using getBodyText, getRawEvents and parseEvents,
// returns parsed json content in bytes.
func parseText(text []pdf.Text, initialDate time.Time) ([]byte, error) {
	bodyText := getBodyText(text)
	rawEvents := parser.GetRawEvents(bodyText, initialDate)
	events, err := parser.ParseEvents(rawEvents)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}
	return json.Marshal(events)
}

// ParseFile reads slice of pdf.Text from input file using reader.ReadFile,
// parses content using parseText and writes content bytes to output file.
func ParseFile(inputFilePath string, outputFilePath string, initialDate time.Time) error {
	text, err := reader.ReadFile(inputFilePath)
	if err != nil {
		return err
	}

	jsonBytes, err := parseText(text, initialDate)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFilePath, jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// ParseBytes gets slice of pdf.Text from content bytes using reader.ReadBytes,
// parses content using parseText and returns content bytes.
func ParseBytes(contentBytes []byte, initialDate time.Time) ([]byte, error) {
	text, err := reader.ReadBytes(contentBytes)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := parseText(text, initialDate)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
