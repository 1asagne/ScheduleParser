// Package scheduleparser provides structs and functions for parsing pdf schedules in a specific format to json.

package scheduleparser

import (
	"encoding/json"
	"os"

	"github.com/ledongthuc/pdf"
)

// getSchedule takes slice of ScheduleEventRaw,
// forms slice of ScheduleEvent and returns it.
func getSchedule(rawItems []ScheduleEventRaw) ([]ScheduleEvent, error) {
	lessons := make([]ScheduleEvent, 0)
	for _, rawItem := range rawItems {
		lesson, err := rawItem.extractScheduleEvent()
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

// getRawEvents takes slice of pdf.Text,
// forms slice of ScheduleEventRaw and returns it.
func getRawEvents(texts []pdf.Text) []ScheduleEventRaw {
	rawItems := []ScheduleEventRaw{}
	var rawItem ScheduleEventRaw
	for i, text := range texts {
		if rawItem.data == "" {
			rawItem.position = pdf.Point{X: text.X, Y: text.Y}
		} else if text.Y != texts[i-1].Y {
			rawItem.data += " "
		}
		rawItem.data += text.S
		if text.S == "]" {
			rawItems = append(rawItems, rawItem)
			rawItem.data = ""
		}
	}
	return rawItems
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

// parseScheduleText takes slice of pdf.Text,
// parses content using getMainTexts, getRawEvents and getSchedule,
// returns parsed content in bytes
func parseScheduleText(texts []pdf.Text) ([]byte, error) {
	mainTexts := getMainTexts(texts)
	rawItems := getRawEvents(mainTexts)
	lessons, err := getSchedule(rawItems)
	if err != nil {
		return nil, err
	}
	return json.Marshal(lessons)
}

// ParseScheduleFile gets slice of pdf.Text from input file using readPdfFile,
// parses content using parseScheduleText and writes content bytes to output file.
func ParseScheduleFile(inputFilePath string, outputFilePath string) error {
	texts, err := readPdfFile(inputFilePath)
	if err != nil {
		return err
	}

	jsonBytes, err := parseScheduleText(texts)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFilePath, jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// ParseScheduleBytes gets slice of pdf.Text from content bytes using readPdfBytes,
// parses content using parseScheduleText and returns content bytes in json format.
func ParseScheduleBytes(contentBytes []byte) ([]byte, error) {
	texts, err := readPdfBytes(contentBytes)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := parseScheduleText(texts)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
