package scheduleparser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

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

func (cell Cell) extractScheduleItem() (ScheduleItem, error) {
	lesson := ScheduleItem{}

	datesRegexp, _ := regexp.Compile(`\[.+\]$`)
	datesIndex := datesRegexp.FindStringIndex(cell.data)[0]
	lesson.Dates = extractDates(cell.data[datesIndex:])

	const (
		lecture = "лекции"
		seminar = "семинар"
		lab     = "лабораторные занятия"
	)
	typeRegexp, _ := regexp.Compile(fmt.Sprintf(`(%s|%s|%s)\.`, lecture, seminar, lab))
	typeIndexes := typeRegexp.FindStringIndex(cell.data)
	if typeIndexes == nil {
		return ScheduleItem{}, errors.New("Schedule item type is not found")
	}

	switch cell.data[typeIndexes[0] : typeIndexes[1]-1] {
	case lecture:
		lesson.Type = "Лекция"
	case seminar:
		lesson.Type = "Семинар"
	case lab:
		lesson.Type = "Лабораторная работа"
	}

	beforeType := strings.Split(strings.TrimSpace(cell.data[:typeIndexes[0]]), ". ")
	if len(beforeType) == 1 {
		lesson.Title = strings.TrimSuffix(beforeType[0], ".")
	} else {
		lesson.Title = beforeType[0]
		lesson.Teacher = beforeType[1]
	}

	afterType := strings.Split(strings.TrimSuffix(strings.TrimLeft(cell.data[typeIndexes[1]:datesIndex], " "), ". "), ". ")

	if len(afterType) == 2 {
		lesson.Subgroup = strings.Trim(afterType[0], "()")
		lesson.Location = afterType[1]
		lesson.Time = extractTime(cell.position, true)
	} else {
		lesson.Location = afterType[0]
		lesson.Time = extractTime(cell.position, false)
	}

	return lesson, nil
}

func getSchedule(cells []Cell) ([]ScheduleItem, error) {
	lessons := make([]ScheduleItem, 0)
	for _, cell := range cells {
		lesson, err := cell.extractScheduleItem()
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
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

func readPdfBytes(fileBytes []byte) ([]pdf.Text, error) {
	bytesReader := bytes.NewReader(fileBytes)
	pdfReader, err := pdf.NewReader(bytesReader, int64(len(fileBytes)))
	if err != nil {
		return nil, err
	}
	page := pdfReader.Page(1)
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

func getMainTexts(texts []pdf.Text) []pdf.Text {
	mainTexts := []pdf.Text{}
	for _, text := range texts {
		if text.Y < 521 && text.X > 42 {
			mainTexts = append(mainTexts, text)
		}
	}
	return mainTexts
}

func parseScheduleText(texts []pdf.Text) ([]byte, error) {
	mainTexts := getMainTexts(texts)
	cells := getCells(mainTexts)
	lessons, err := getSchedule(cells)
	if err != nil {
		return nil, err
	}
	return json.Marshal(lessons)
}

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

func ParseScheduleBytes(fileBytes []byte) ([]byte, error) {
	texts, err := readPdfBytes(fileBytes)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := parseScheduleText(texts)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
