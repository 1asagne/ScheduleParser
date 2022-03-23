package scheduleparser

import (
	"encoding/json"
	"os"

	"github.com/ledongthuc/pdf"
)

func getSchedule(rawItems []ScheduleItemRaw) ([]ScheduleItem, error) {
	lessons := make([]ScheduleItem, 0)
	for _, rawItem := range rawItems {
		lesson, err := rawItem.extractScheduleItem()
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

func getRawItems(texts []pdf.Text) []ScheduleItemRaw {
	rawItems := []ScheduleItemRaw{}
	var rawItem ScheduleItemRaw
	for i, text := range texts {
		if rawItem.data == "" {
			rawItem.position = pdf.Point{X: text.X, Y: text.Y}
		} else if text.Y != texts[i-1].Y {
			rawItem.data += " "
		}
		rawItem.data += text.S
		if text.S == "]" {
			rawItems = append(rawItems, rawItem)
			rawItem.data = ""
		}
	}
	return rawItems
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
	rawItems := getRawItems(mainTexts)
	lessons, err := getSchedule(rawItems)
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
