// Main package provides entrypoint to use schedule parser as executable.

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/1asagne/scheduleparser"
)

// Main function reads command-line arguments,
// parses input and output files paths and uses ParseFile from scheduleparser package.
func main() {
	var inputFilePath, outputFilePath, date string
	flag.StringVar(&inputFilePath, "i", "", "Input pdf file path")
	flag.StringVar(&outputFilePath, "o", "", "Output json file path")
	flag.StringVar(&date, "d", "", "Initial date in 'dd.MM.YYYY' format to determine years of events dates")
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

	err = scheduleparser.ParseFile(inputFilePath, outputFilePath, initialDate)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Parsing completed successfully.\nOutput JSON file: %v\n", outputFilePath)
}
