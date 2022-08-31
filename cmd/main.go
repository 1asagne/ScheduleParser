// Package main provides entrypoint to use schedule parser as executable.

package main

import (
	"flag"
	"fmt"
	"time"

	sp "github.com/qsoulior/scheduleparser/pkg/parser"
)

// main reads command-line arguments, parses input
// and output files paths and uses ParseFile from parser package.
func main() {
	var inputFilePath, outputFilePath, date string
	flag.StringVar(&inputFilePath, "i", "", "Input pdf file path")
	flag.StringVar(&outputFilePath, "o", "", "Output json file path")
	flag.StringVar(&date, "d", "", "Initial date in 'dd.MM.YYYY' format to determine year of events dates")
	flag.Parse()

	if inputFilePath == "" || outputFilePath == "" {
		flag.PrintDefaults()
		return
	}

	var (
		initialDate time.Time
		err         error
	)
	if date == "" {
		initialDate = time.Now()
	} else {
		initialDate, err = time.Parse("02.01.2006", date)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	err = sp.ParseFile(inputFilePath, outputFilePath, initialDate)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Parsing completed successfully.\nOutput JSON file: %v\n", outputFilePath)
}
