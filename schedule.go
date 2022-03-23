// Package scheduleparser provides structs and functions for parsing pdf schedules in a specific format to json.

package scheduleparser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ScheduleEventRaw is a schedule event that contains data string and position.
// It is retrieved from pdf file.
type ScheduleEventRaw struct {
	data     string
	position pdf.Point
}

// ScheduleEventTime contains start/end time of schedule event.
type ScheduleEventTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ScheduleEventDate contains start/end date and frequency of schedule event.
type ScheduleEventDate struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Frequency string `json:"frequency"`
}

// ScheduleEvent is a schedule event in json format.
type ScheduleEvent struct {
	Title    string              `json:"title"`
	Teacher  string              `json:"teacher"`
	Type     string              `json:"type"`
	Subgroup string              `json:"subgroup"`
	Location string              `json:"location"`
	Time     ScheduleEventTime   `json:"time"`
	Dates    []ScheduleEventDate `json:"dates"`
}

var eventTimes = [...]ScheduleEventTime{
	{"8:30", "10:10"},
	{"10:20", "12:00"},
	{"12:20", "14:00"},
	{"14:10", "15:50"},
	{"16:00", "17:40"},
	{"18:00", "19:30"},
	{"19:40", "21:10"},
	{"21:20", "22:50"},
}

// extractTime searches for time in schedule event data and extracts it,
// returns ScheduleEventTime.
func (rawEvent *ScheduleEventRaw) extractTime(isLab bool) ScheduleEventTime {
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
		return ScheduleEventTime{Start: eventTimes[timesIndex].Start, End: eventTimes[timesIndex+1].End}
	}
	return eventTimes[timesIndex]
}

// extractDates searches for dates in schedule event data and extracts them,
// returns slice of ScheduleEventDate and index of first occurrence.
func (rawEvent *ScheduleEventRaw) extractDates() ([]ScheduleEventDate, int) {
	datesRegexp, _ := regexp.Compile(`\[.+\]$`)
	datesIndex := datesRegexp.FindStringIndex(rawEvent.data)[0]
	datesString := rawEvent.data[datesIndex:]

	dates := make([]ScheduleEventDate, 0)
	datesString = strings.Trim(datesString, "[]")
	for _, complexDate := range strings.Split(datesString, ", ") {
		splitDate := strings.Split(complexDate, " ")
		date := ScheduleEventDate{}
		switch len(splitDate) {
		case 1:
			date.Start = splitDate[0]
			date.End = splitDate[0]
			date.Frequency = "once"
		case 2:
			switch splitDate[1] {
			case "к.н.":
				date.Frequency = "every"
			case "ч.н.":
				date.Frequency = "throughout"
			}
			splitDate = strings.Split(splitDate[0], "-")
			date.Start = splitDate[0]
			date.End = splitDate[1]
		}
		dates = append(dates, date)
	}
	return dates, datesIndex
}

// extractScheduleEvent extracts ScheduleEvent from event data using extractTime and extractDates.
func (rawEvent *ScheduleEventRaw) extractScheduleEvent() (ScheduleEvent, error) {
	scheduleEvent := ScheduleEvent{}

	var datesIndex int
	scheduleEvent.Dates, datesIndex = rawEvent.extractDates()

	const (
		lecture = "лекции"
		seminar = "семинар"
		lab     = "лабораторные занятия"
	)
	typeRegexp, _ := regexp.Compile(fmt.Sprintf(`(%s|%s|%s)\.`, lecture, seminar, lab))
	typeIndexes := typeRegexp.FindStringIndex(rawEvent.data)
	if typeIndexes == nil {
		return ScheduleEvent{}, errors.New("Schedule event type is not found")
	}

	switch rawEvent.data[typeIndexes[0] : typeIndexes[1]-1] {
	case lecture:
		scheduleEvent.Type = "Лекция"
	case seminar:
		scheduleEvent.Type = "Семинар"
	case lab:
		scheduleEvent.Type = "Лабораторная работа"
	}

	beforeType := strings.Split(strings.TrimSpace(rawEvent.data[:typeIndexes[0]]), ". ")
	if len(beforeType) == 1 {
		scheduleEvent.Title = strings.TrimSuffix(beforeType[0], ".")
	} else {
		scheduleEvent.Title = beforeType[0]
		scheduleEvent.Teacher = beforeType[1]
	}

	afterType := strings.Split(strings.TrimSuffix(strings.TrimLeft(rawEvent.data[typeIndexes[1]:datesIndex], " "), ". "), ". ")

	if len(afterType) == 2 {
		scheduleEvent.Subgroup = strings.Trim(afterType[0], "()")
		scheduleEvent.Location = afterType[1]
		scheduleEvent.Time = rawEvent.extractTime(true)
	} else {
		scheduleEvent.Location = afterType[0]
		scheduleEvent.Time = rawEvent.extractTime(false)
	}

	return scheduleEvent, nil
}
