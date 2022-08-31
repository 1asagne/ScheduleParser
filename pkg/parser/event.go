// Package parser implements structs and functions to parse events from pdf content.

package parser

import (
	"time"

	"github.com/ledongthuc/pdf"
)

// RawEvent contains data, position in pdf file, and initial date to normalize event dates.
// It is retrieved from input pdf.
type RawEvent struct {
	data        string
	position    pdf.Point
	initialDate time.Time
}

// Event is retrieved from RawEvent. It is contained in output json.
type Event struct {
	Title    string      `json:"title"`
	Teacher  string      `json:"teacher"`
	Type     string      `json:"type"`
	Subgroup string      `json:"subgroup"`
	Location string      `json:"location"`
	Dates    []EventDate `json:"dates"`
}

// getRawEvents takes slice of pdf.Text, forms slice of RawEvent and returns it.
func getRawEvents(texts []pdf.Text, initialDate time.Time) []RawEvent {
	rawEvents := make([]RawEvent, 0)
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
