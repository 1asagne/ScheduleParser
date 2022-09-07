// Package scheduleparser implements structs and functions to parse events from pdf content.

package scheduleparser

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/qsoulior/scheduleparser/internal/reader"
)

// parseText takes slice of pdf.Text,
// parses content using getBodyText, getRawEvents and parseEvents,
// returns parsed json content in bytes.
func parseText(text []pdf.Text, initialDate time.Time) ([]byte, error) {
	rawEvents := getRawEvents(text, initialDate)
	events, err := parseEvents(rawEvents)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}
	return json.Marshal(events)
}

// ParseFile reads slice of pdf.Text from input file using reader.ReadFile,
// parses content using parseText and writes content bytes to output file.
func ParseFile(inputPath string, outputPath string, initialDate time.Time) error {
	text, err := reader.ReadFile(inputPath)
	if err != nil {
		return err
	}

	jsonBytes, err := parseText(text, initialDate)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, jsonBytes, 0644)
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
