// Package scheduleparser provides structs and functions for parsing pdf schedules in a specific format to json.

package scheduleparser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// RawEvent is a schedule event that contains data string, position in pdf file,
// and initial date to normalize event dates.
// It is retrieved from pdf file.
type RawEvent struct {
	data        string
	position    pdf.Point
	initialDate time.Time
}

// RawEventTime contains schedule event start and end times strings.
// It is retrieved by RawEvent position.
type RawEventTime struct {
	start string
	end   string
}

var eventTimes = [...]RawEventTime{
	{"08:30", "10:10"},
	{"10:20", "12:00"},
	{"12:20", "14:00"},
	{"14:10", "15:50"},
	{"16:00", "17:40"},
	{"18:00", "19:30"},
	{"19:40", "21:10"},
	{"21:20", "22:50"},
}

// getTime gets event time by schedule event position,
// returns EventTime.
func (rawEvent *RawEvent) getTime(isLab bool) RawEventTime {
	var timesIndex int
	switch int(rawEvent.position.X) {
	case 46:
		timesIndex = 0
	case 139:
		timesIndex = 1
	case 233:
		timesIndex = 2
	case 327:
		timesIndex = 3
	case 420:
		timesIndex = 4
	case 514:
		timesIndex = 5
	case 607:
		timesIndex = 6
	default:
		timesIndex = 7
	}
	if isLab {
		return RawEventTime{eventTimes[timesIndex].start, eventTimes[timesIndex+1].end}
	}
	return eventTimes[timesIndex]
}

// EventDate contains start/end date and frequency of schedule event.
type EventDate struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Frequency string    `json:"frequency"`
}

// normalize adds a year to start date and end date by given date.
func (eventDate *EventDate) normalize(date time.Time) {
	day, month, year := date.Day(), date.Month(), date.Year()
	if (eventDate.Start.Month() > month) || (eventDate.Start.Month() == month && eventDate.Start.Day() >= day) {
		eventDate.Start = eventDate.Start.AddDate(year, 0, 0)
	} else {
		eventDate.Start = eventDate.Start.AddDate(year+1, 0, 0)
	}
	if (eventDate.End.Month() > month) || (eventDate.End.Month() == month && eventDate.End.Day() >= day) {
		eventDate.End = eventDate.End.AddDate(year, 0, 0)
	} else {
		eventDate.End = eventDate.End.AddDate(year+1, 0, 0)
	}
}

const dateStringFormat = "02.01 15:04 -0700"

// makeEventDate creates EventDate by start date and end date strings,
// and returns it.
func makeEventDate(startString string, endString string, frequency string) EventDate {
	eventDate := EventDate{}
	eventDate.Start, _ = time.Parse(dateStringFormat, startString)
	eventDate.End, _ = time.Parse(dateStringFormat, endString)
	eventDate.Frequency = frequency
	return eventDate
}

// Event is a schedule event in json format.
type Event struct {
	Title    string      `json:"title"`
	Teacher  string      `json:"teacher"`
	Type     string      `json:"type"`
	Subgroup string      `json:"subgroup"`
	Location string      `json:"location"`
	Dates    []EventDate `json:"dates"`
}

// extractDates searches for dates in schedule event data and extracts them,
// returns slice of EventDate and index of first occurrence.
func (rawEvent *RawEvent) extractDates(isLab bool) ([]EventDate, int) {
	datesRegexp, _ := regexp.Compile(`\[.+\]$`)
	datesIndex := datesRegexp.FindStringIndex(rawEvent.data)[0]
	datesString := rawEvent.data[datesIndex:]
	time := rawEvent.getTime(isLab)
	timeStartString := " " + time.start + " +0300"
	timeEndString := " " + time.end + " +0300"

	dates := make([]EventDate, 0)
	datesString = strings.Trim(datesString, "[]")
	for _, complexDate := range strings.Split(datesString, ", ") {
		splitDate := strings.Split(complexDate, " ")
		date := EventDate{}
		switch len(splitDate) {
		case 1:
			date = makeEventDate(splitDate[0]+timeStartString, splitDate[0]+timeEndString, "once")
		case 2:
			dateFrequency := splitDate[1]
			splitDate := strings.Split(splitDate[0], "-")
			if dateFrequency == "к.н." {
				date = makeEventDate(splitDate[0]+timeStartString, splitDate[1]+timeEndString, "every")
			} else if dateFrequency == "ч.н." {
				date = makeEventDate(splitDate[0]+timeStartString, splitDate[1]+timeEndString, "throughout")
			}
		}
		date.normalize(rawEvent.initialDate)
		dates = append(dates, date)
	}
	return dates, datesIndex
}

// extractEvent extracts Event from event data using extractTime and extractDates.
func (rawEvent *RawEvent) extractEvent() (Event, error) {
	event := Event{}

	const (
		lecture = "лекции"
		seminar = "семинар"
		lab     = "лабораторные занятия"
	)

	// Extract Event type from RawEvent data.
	typeRegexp, _ := regexp.Compile(fmt.Sprintf(`(%s|%s|%s)\.`, lecture, seminar, lab))
	typeIndexes := typeRegexp.FindStringIndex(rawEvent.data)
	if typeIndexes == nil {
		return Event{}, errors.New("Schedule event type is not found")
	}

	switch rawEvent.data[typeIndexes[0] : typeIndexes[1]-1] {
	case lecture:
		event.Type = "lecture"
	case seminar:
		event.Type = "seminar"
	case lab:
		event.Type = "lab"
	}

	// Extract Event title and teacher (if exists) from RawEvent data.
	beforeType := strings.Split(strings.TrimSpace(rawEvent.data[:typeIndexes[0]]), ". ")
	if len(beforeType) == 1 {
		event.Title = strings.TrimSuffix(beforeType[0], ".")
	} else {
		event.Title = beforeType[0]
		event.Teacher = beforeType[1]
	}

	var datesStartIndex int

	// Exract Event dates with times from RawEvent data and position.
	if event.Type == "lab" {
		event.Dates, datesStartIndex = rawEvent.extractDates(true)
	} else {
		event.Dates, datesStartIndex = rawEvent.extractDates(false)
	}

	// Get RawEvent data slice with subgroup (if exists) and location.
	// This data slice is between type and dates.
	afterType := strings.Split(strings.TrimSuffix(strings.TrimLeft(rawEvent.data[typeIndexes[1]:datesStartIndex], " "), ". "), ". ")

	if len(afterType) == 2 {
		event.Subgroup = strings.Trim(afterType[0], "()")
		event.Location = afterType[1]
	} else {
		event.Location = afterType[0]
	}

	return event, nil
}
