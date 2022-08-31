// Package parser implements structs and functions to parse events from pdf content.

package parser

import "errors"

// Clock contains hours and minutes values
type Clock struct {
	hour int
	min  int
}

// EventTime contains start/end Clock.
// It is retrieved by RawEvent position.
type EventTime struct {
	start Clock
	end   Clock
}

// eventTimes is slice of determined EventTime instances
var eventTimes = [...]EventTime{
	{Clock{8, 30}, Clock{10, 10}},
	{Clock{10, 20}, Clock{12, 0}},
	{Clock{12, 20}, Clock{14, 0}},
	{Clock{14, 10}, Clock{15, 50}},
	{Clock{16, 0}, Clock{17, 40}},
	{Clock{18, 0}, Clock{19, 30}},
	{Clock{19, 40}, Clock{21, 10}},
	{Clock{21, 20}, Clock{22, 50}},
}

// parseTime gets *EventTime by raw event position,
// and returns it.
func parseTime(raw *RawEvent, shift int) (*EventTime, error) {
	var timesIndex int

	pos := map[int]int{46: 0, 139: 1, 233: 2, 327: 3, 420: 4, 514: 5, 607: 6}
	timesIndex, ok := pos[int(raw.position.X)]
	if !ok {
		timesIndex = 7
	}

	if shift != 0 {
		if timesIndex+shift >= len(eventTimes) {
			return nil, errors.New("shift is out of range")
		}
		return &EventTime{eventTimes[timesIndex].start, eventTimes[timesIndex+shift].end}, nil
	}
	return &eventTimes[timesIndex], nil
}
