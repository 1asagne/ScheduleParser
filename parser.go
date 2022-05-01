// Package scheduleparser provides structs and functions for parsing pdf schedules in a specific format to json.

package scheduleparser

import (
	"encoding/json"
	"os"
	"time"

	"github.com/ledongthuc/pdf"
)

// getEvents takes slice of RawEvent,
// forms slice of Event and returns it.
func getEvents(rawEvents []RawEvent) ([]Event, error) {
	events := make([]Event, 0)
	for _, rawEvent := range rawEvents {
		event, err := rawEvent.extractEvent()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// getRawEvents takes slice of pdf.Text,
// forms slice of RawEvent and returns it.
func getRawEvents(texts []pdf.Text, initialDate time.Time) []RawEvent {
	rawEvents := []RawEvent{}
	var rawEvent RawEvent
	for i, text := range texts {
		if rawEvent.data == "" {
			rawEvent.position = pdf.Point{X: text.X, Y: text.Y}
		} else if text.Y != texts[i-1].Y {
			rawEvent.data += " "
		}
		rawEvent.data += text.S
		if text.S == "]" {
			rawEvent.initialDate = initialDate
			rawEvents = append(rawEvents, rawEvent)
			rawEvent.data = ""
		}
	}
	return rawEvents
}

// getMainTexts takes slice of pdf.Text,
// returns slice of pdf.Text without junk.
func getMainTexts(texts []pdf.Text) []pdf.Text {
	mainTexts := []pdf.Text{}
	for _, text := range texts {
		if text.Y < 521 && text.X > 42 {
			mainTexts = append(mainTexts, text)
		}
	}
	return mainTexts
}

// parseTexts takes slice of pdf.Text,
// parses content using getMainTexts, getRawEvents and getEvents,
// returns parsed content in bytes
func parseTexts(texts []pdf.Text, initialDate time.Time) ([]byte, error) {
	mainTexts := getMainTexts(texts)
	rawEvents := getRawEvents(mainTexts, initialDate)
	events, err := getEvents(rawEvents)
	if err != nil {
		return nil, err
	}
	return json.Marshal(events)
}

// ParseFile reads slice of pdf.Text from input file using readPdfFile,
// parses content using parseTexts and writes content bytes to output file.
func ParseFile(inputFilePath string, outputFilePath string, initialDate time.Time) error {
	texts, err := readPdfFile(inputFilePath)
	if err != nil {
		return err
	}

	jsonBytes, err := parseTexts(texts, initialDate)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFilePath, jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// ParseBytes gets slice of pdf.Text from content bytes using readPdfBytes,
// parses content using parseTexts and returns content bytes in json format.
func ParseBytes(contentBytes []byte, initialDate time.Time) ([]byte, error) {
	texts, err := readPdfBytes(contentBytes)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := parseTexts(texts, initialDate)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
