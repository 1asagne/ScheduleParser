// Package scheduleparser implements structs and functions to parse events from pdf content.

package scheduleparser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// EventDate contains start/end datetime and frequency of schedule event.
type EventDate struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Frequency string    `json:"frequency"`
}

// normalize adds a year to start datetime and end datetime by given date.
func (eventDate *EventDate) normalize(date time.Time) {
	year, month, day := date.Date()
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

var loc = time.FixedZone("UTC+3", 3*60*60)

// NewEventDate creates EventDate by start date and end date strings,
// adds time to date by eventTime and returns *EventDate.
func NewEventDate(start string, end string, eventTime *EventTime, frequency string) *EventDate {
	dateStart, _ := time.ParseInLocation(dateFormat, start, loc)
	dateEnd, _ := time.ParseInLocation(dateFormat, end, loc)
	dateStart = dateStart.Add(time.Hour*time.Duration(eventTime.start.hour) + time.Minute*time.Duration(eventTime.start.min))
	dateEnd = dateEnd.Add(time.Hour*time.Duration(eventTime.end.hour) + time.Minute*time.Duration(eventTime.end.min))

	return &EventDate{dateStart, dateEnd, frequency}
}

// parseDates searches for dates in raw event data and extracts them,
// returns slice of EventDate and index of first occurrence.
func parseDates(raw *RawEvent, shift int) ([]EventDate, int, error) {
	datesRegexp := regexp.MustCompile(`\[.+\]$`)
	datesIndex := datesRegexp.FindStringIndex(raw.data)[0]

	eventTime, err := parseTime(raw, shift)
	if err != nil {
		return nil, -1, fmt.Errorf("parseTime error: %w", err)
	}

	// [09.09-28.10 к.н., 11.11, 18.11]
	datesString := strings.Trim(raw.data[datesIndex:], "[]")
	dates := make([]EventDate, 0)
	for _, complexDate := range strings.Split(datesString, ", ") {
		splitDate := strings.Split(complexDate, " ")
		var date *EventDate

		if dateLength := len(splitDate); dateLength == 1 {
			date = NewEventDate(splitDate[0], splitDate[0], eventTime, "once")
		} else if dateLength == 2 {
			dateFrequency := splitDate[1]
			splitDate := strings.Split(splitDate[0], "-")
			if dateFrequency == "к.н." {
				date = NewEventDate(splitDate[0], splitDate[1], eventTime, "every")
			} else if dateFrequency == "ч.н." {
				date = NewEventDate(splitDate[0], splitDate[1], eventTime, "throughout")
			}
		}
		date.normalize(raw.initialDate)
		dates = append(dates, *date)
	}
	return dates, datesIndex, nil
}
