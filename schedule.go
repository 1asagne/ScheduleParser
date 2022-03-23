package scheduleparser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

type ScheduleItemRaw struct {
	data     string
	position pdf.Point
}

type ScheduleItemTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ScheduleItemDate struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Frequency string `json:"frequency"`
}

type ScheduleItem struct {
	Title    string             `json:"title"`
	Teacher  string             `json:"teacher"`
	Type     string             `json:"type"`
	Subgroup string             `json:"subgroup"`
	Location string             `json:"location"`
	Time     ScheduleItemTime   `json:"time"`
	Dates    []ScheduleItemDate `json:"dates"`
}

var times = [...]ScheduleItemTime{
	{"8:30", "10:10"},
	{"10:20", "12:00"},
	{"12:20", "14:00"},
	{"14:10", "15:50"},
	{"16:00", "17:40"},
	{"18:00", "19:30"},
	{"19:40", "21:10"},
	{"21:20", "22:50"},
}

func (rawItem *ScheduleItemRaw) extractTime(isLab bool) ScheduleItemTime {
	var timesIndex int
	switch int(rawItem.position.X) {
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
		return ScheduleItemTime{Start: times[timesIndex].Start, End: times[timesIndex+1].End}
	}
	return times[timesIndex]
}

func (rawItem *ScheduleItemRaw) extractDates() ([]ScheduleItemDate, int) {
	datesRegexp, _ := regexp.Compile(`\[.+\]$`)
	datesIndex := datesRegexp.FindStringIndex(rawItem.data)[0]
	datesString := rawItem.data[datesIndex:]

	dates := make([]ScheduleItemDate, 0)
	datesString = strings.Trim(datesString, "[]")
	for _, complexDate := range strings.Split(datesString, ", ") {
		splitDate := strings.Split(complexDate, " ")
		date := ScheduleItemDate{}
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

func (rawItem *ScheduleItemRaw) extractScheduleItem() (ScheduleItem, error) {
	scheduleItem := ScheduleItem{}

	var datesIndex int
	scheduleItem.Dates, datesIndex = rawItem.extractDates()

	const (
		lecture = "лекции"
		seminar = "семинар"
		lab     = "лабораторные занятия"
	)
	typeRegexp, _ := regexp.Compile(fmt.Sprintf(`(%s|%s|%s)\.`, lecture, seminar, lab))
	typeIndexes := typeRegexp.FindStringIndex(rawItem.data)
	if typeIndexes == nil {
		return ScheduleItem{}, errors.New("Schedule item type is not found")
	}

	switch rawItem.data[typeIndexes[0] : typeIndexes[1]-1] {
	case lecture:
		scheduleItem.Type = "Лекция"
	case seminar:
		scheduleItem.Type = "Семинар"
	case lab:
		scheduleItem.Type = "Лабораторная работа"
	}

	beforeType := strings.Split(strings.TrimSpace(rawItem.data[:typeIndexes[0]]), ". ")
	if len(beforeType) == 1 {
		scheduleItem.Title = strings.TrimSuffix(beforeType[0], ".")
	} else {
		scheduleItem.Title = beforeType[0]
		scheduleItem.Teacher = beforeType[1]
	}

	afterType := strings.Split(strings.TrimSuffix(strings.TrimLeft(rawItem.data[typeIndexes[1]:datesIndex], " "), ". "), ". ")

	if len(afterType) == 2 {
		scheduleItem.Subgroup = strings.Trim(afterType[0], "()")
		scheduleItem.Location = afterType[1]
		scheduleItem.Time = rawItem.extractTime(true)
	} else {
		scheduleItem.Location = afterType[0]
		scheduleItem.Time = rawItem.extractTime(false)
	}

	return scheduleItem, nil
}
