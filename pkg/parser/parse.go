// Package parser implements structs and functions to parse events from pdf content.

package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/qsoulior/scheduleparser/internal/reader"
)

// parseEvent parses *RawEvent and returns *Event.
func parseEvent(raw *RawEvent) (*Event, error) {
	// Parse type from data.
	const (
		lecture = "лекции"
		seminar = "семинар"
		lab     = "лабораторные занятия"
	)
	typeRegexp := regexp.MustCompile(fmt.Sprintf(`(%s|%s|%s)\.`, lecture, seminar, lab))
	typeIndexes := typeRegexp.FindStringIndex(raw.data)
	if typeIndexes == nil {
		return nil, errors.New("schedule event type is not found")
	}
	eventTypes := map[string]string{
		lecture: "lecture",
		seminar: "seminar",
		lab:     "lab",
	}
	eventType := eventTypes[raw.data[typeIndexes[0]:typeIndexes[1]-1]]

	// Parse title and teacher from data.
	var eventTitle, eventTeacher string

	stringsBeforeType := strings.Split(raw.data[:typeIndexes[0]-1], ". ")
	if len(stringsBeforeType) == 1 {
		eventTitle = stringsBeforeType[0][:len(stringsBeforeType)-1]
	} else {
		eventTitle = stringsBeforeType[0]
		eventTeacher = stringsBeforeType[1]
	}

	// Parse dates from data and position.
	var (
		eventDates      []EventDate
		datesStartIndex int
		err             error
	)
	if eventType == "lab" {
		eventDates, datesStartIndex, err = parseDates(raw, 1)
	} else {
		eventDates, datesStartIndex, err = parseDates(raw, 0)
	}
	if err != nil {
		return nil, fmt.Errorf("parseDates error: %w", err)
	}

	// Parse subgroup and location from data.
	var eventSubgroup, eventLocation string

	stringsAfterType := strings.Split(raw.data[typeIndexes[1]+1:datesStartIndex-2], ". ")
	if len(stringsAfterType) == 2 {
		eventSubgroup = strings.Trim(stringsAfterType[0], "()")
		eventLocation = stringsAfterType[1]
	} else {
		eventLocation = stringsAfterType[0]
	}

	return &Event{
		eventTitle,
		eventTeacher,
		eventType,
		eventSubgroup,
		eventLocation,
		eventDates,
	}, nil
}

// parseEvents takes slice of RawEvent, forms slice of Event and returns it.
func parseEvents(rawEvents []RawEvent) ([]Event, error) {
	events := make([]Event, 0)
	for i, rawEvent := range rawEvents {
		event, err := parseEvent(&rawEvent)
		if err != nil {
			return nil, fmt.Errorf("parse events[%d]: %w", i, err)
		}
		events = append(events, *event)
	}
	return events, nil
}

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
	rawEvents := getRawEvents(bodyText, initialDate)
	events, err := parseEvents(rawEvents)
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
