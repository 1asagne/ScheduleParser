package parser

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

const PrefixSymbolsCount = 156

type Cell struct {
	data     string
	position pdf.Point
}

type Time struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Date struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Frequency string `json:"frequency"`
}

type ScheduleItem struct {
	Title    string `json:"title"`
	Teacher  string `json:"teacher"`
	Type     string `json:"type"`
	Subgroup string `json:"subgroup"`
	Location string `json:"location"`
	Time     Time   `json:"time"`
	Dates    []Date `json:"dates"`
}

var times = [...]Time{
	{"8:30", "10:10"},
	{"10:20", "12:00"},
	{"12:20", "14:00"},
	{"14:10", "15:50"},
	{"16:00", "17:40"},
	{"18:00", "19:30"},
	{"19:40", "21:10"},
	{"21:20", "22:50"},
}

func extractTime(position pdf.Point, isLab bool) Time {
	var timesIndex int
	switch int(position.X) {
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
		return Time{Start: times[timesIndex].Start, End: times[timesIndex+1].End}
	}
	return times[timesIndex]
}

func extractDates(datesString string) []Date {
	dates := make([]Date, 0)
	datesString = strings.Trim(datesString, "[]")
	for _, complexDate := range strings.Split(datesString, ", ") {
		splitDate := strings.Split(complexDate, " ")
		date := Date{}
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
	return dates
}

func (cell Cell) extractScheduleItem() ScheduleItem {
	lesson := ScheduleItem{}

	re, _ := regexp.Compile(`^(.+\. )(\[.+\])$`)
	submatchStrings := re.FindStringSubmatch(cell.data)[1:]

	lesson.Dates = extractDates(submatchStrings[1])

	data := strings.Split(submatchStrings[0], ". ")
	data = data[:len(data)-1]

	lesson.Title = data[0]

	const (
		lecture = "лек"
		seminar = "сем"
		lab     = "лаб"
	)

	var typeIndex int = -1

	for i, item := range data {
		if strings.Contains(item, lecture) || strings.Contains(item, seminar) || strings.Contains(item, lab) {
			typeIndex = i
			break
		}
	}

	if typeIndex == 2 {
		lesson.Teacher = data[1]
		if strings.Contains(data[1], ".") {
			lesson.Teacher += "."
		}
	}

	if strings.Contains(data[typeIndex], lecture) || strings.Contains(data[typeIndex], seminar) {
		if strings.Contains(data[typeIndex], lecture) {
			lesson.Type = "Лекция"
		} else {
			lesson.Type = "Семинар"
		}
		if typeIndex+1 < len(data) {
			lesson.Location = data[typeIndex+1]
		}
		lesson.Time = extractTime(cell.position, false)
	} else {
		lesson.Type = "Лабораторная работа"
		lesson.Subgroup = strings.Trim(data[typeIndex+1], "()")
		if typeIndex+2 < len(data) {
			lesson.Location = data[typeIndex+2]
		}
		lesson.Time = extractTime(cell.position, true)
	}

	return lesson
}

func getSchedule(cells []Cell) []ScheduleItem {
	lessons := make([]ScheduleItem, 0)
	for _, cell := range cells {
		lesson := cell.extractScheduleItem()
		lessons = append(lessons, lesson)
	}
	return lessons
}

func readPdfFile(filePath string) ([]pdf.Text, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	page := reader.Page(1)
	texts := page.Content().Text
	return texts, nil
}

func getCells(texts []pdf.Text) []Cell {
	cells := []Cell{}
	var cell Cell
	for i, text := range texts {
		if cell.data == "" {
			cell.position = pdf.Point{X: text.X, Y: text.Y}
		} else if text.Y != texts[i-1].Y {
			cell.data += " "
		}
		cell.data += text.S
		if text.S == "]" {
			cells = append(cells, cell)
			cell.data = ""
		}
	}
	return cells
}

func ParseScheduleText(texts []pdf.Text) ([]byte, error) {
	cells := getCells(texts[PrefixSymbolsCount:])
	lessons := getSchedule(cells)
	return json.Marshal(lessons)
}

func ParseScheduleFile(inputFilePath string, outputFilePath string) error {
	texts, err := readPdfFile(inputFilePath)
	if err != nil {
		return err
	}

	jsonBytes, err := ParseScheduleText(texts)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFilePath, jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
