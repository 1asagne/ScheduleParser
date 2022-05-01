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

// RawEvent is a schedule event that contains data string and position.
// It is retrieved from pdf file.
type RawEvent struct {
	data        string
	position    pdf.Point
	initialDate time.Time
}

// EventTime contains start/end time of schedule event.
type EventTime struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

const timeFormat = "15:04 -0700"

// makeEventTime creates EventTime by start time and end time strings,
// and returns it.
func makeEventTime(start string, end string) EventTime {
	eventTime := EventTime{}
	eventTime.Start, _ = time.Parse(timeFormat, start+" +0300")
	eventTime.End, _ = time.Parse(timeFormat, end+" +0300")
	return eventTime
}

// EventDate contains start/end date and frequency of schedule event.
type EventDate struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Frequency string    `json:"frequency"`
}

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

const dateFormat = "02.01"

// makeEventDate creates EventDate by start date and end date strings,
// and returns it.
func makeEventDate(start string, end string, frequency string) EventDate {
	eventDate := EventDate{}
	eventDate.Start, _ = time.Parse(dateFormat, start)
	eventDate.End, _ = time.Parse(dateFormat, end)
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
	Time     EventTime   `json:"time"`
	Dates    []EventDate `json:"dates"`
}

var eventTimes = [...]EventTime{
	makeEventTime("08:30", "10:10"),
	makeEventTime("10:20", "12:00"),
	makeEventTime("12:20", "14:00"),
	makeEventTime("14:10", "15:50"),
	makeEventTime("16:00", "17:40"),
	makeEventTime("18:00", "19:30"),
	makeEventTime("19:40", "21:10"),
	makeEventTime("21:20", "22:50"),
}

// extractTime searches for time in schedule event data and extracts it,
// returns EventTime.
func (rawEvent *RawEvent) extractTime(isLab bool) EventTime {
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
		return EventTime{Start: eventTimes[timesIndex].Start, End: eventTimes[timesIndex+1].End}
	}
	return eventTimes[timesIndex]
}

// extractDates searches for dates in schedule event data and extracts them,
// returns slice of EventDate and index of first occurrence.
func (rawEvent *RawEvent) extractDates() ([]EventDate, int) {
	datesRegexp, _ := regexp.Compile(`\[.+\]$`)
	datesIndex := datesRegexp.FindStringIndex(rawEvent.data)[0]
	datesString := rawEvent.data[datesIndex:]

	dates := make([]EventDate, 0)
	datesString = strings.Trim(datesString, "[]")
	for _, complexDate := range strings.Split(datesString, ", ") {
		splitDate := strings.Split(complexDate, " ")
		date := EventDate{}
		switch len(splitDate) {
		case 1:
			date = makeEventDate(splitDate[0], splitDate[0], "once")
		case 2:
			dateFrequency := splitDate[1]
			splitDate := strings.Split(splitDate[0], "-")
			if dateFrequency == "к.н." {
				date = makeEventDate(splitDate[0], splitDate[1], "once")
			} else if dateFrequency == "ч.н." {
				date = makeEventDate(splitDate[0], splitDate[1], "throughout")
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

	var datesIndex int
	event.Dates, datesIndex = rawEvent.extractDates()

	const (
		lecture = "лекции"
		seminar = "семинар"
		lab     = "лабораторные занятия"
	)
	typeRegexp, _ := regexp.Compile(fmt.Sprintf(`(%s|%s|%s)\.`, lecture, seminar, lab))
	typeIndexes := typeRegexp.FindStringIndex(rawEvent.data)
	if typeIndexes == nil {
		return Event{}, errors.New("Schedule event type is not found")
	}

	switch rawEvent.data[typeIndexes[0] : typeIndexes[1]-1] {
	case lecture:
		event.Type = "Лекция"
	case seminar:
		event.Type = "Семинар"
	case lab:
		event.Type = "Лабораторная работа"
	}

	beforeType := strings.Split(strings.TrimSpace(rawEvent.data[:typeIndexes[0]]), ". ")
	if len(beforeType) == 1 {
		event.Title = strings.TrimSuffix(beforeType[0], ".")
	} else {
		event.Title = beforeType[0]
		event.Teacher = beforeType[1]
	}

	afterType := strings.Split(strings.TrimSuffix(strings.TrimLeft(rawEvent.data[typeIndexes[1]:datesIndex], " "), ". "), ". ")

	if len(afterType) == 2 {
		event.Subgroup = strings.Trim(afterType[0], "()")
		event.Location = afterType[1]
		event.Time = rawEvent.extractTime(true)
	} else {
		event.Location = afterType[0]
		event.Time = rawEvent.extractTime(false)
	}

	return event, nil
}
