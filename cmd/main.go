package main

import (
	"flag"
	"fmt"

	parser "github.com/1asagne/scheduleparser"
)

func main() {
	var inputFileName, outputFileName string
	flag.StringVar(&inputFileName, "i", "", "Input pdf file path")
	flag.StringVar(&outputFileName, "o", "", "Output json file path")
	flag.Parse()

	if inputFileName == "" || outputFileName == "" {
		flag.PrintDefaults()
		return
	}

	err := parser.ParseScheduleFile(inputFileName, outputFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Parsing completed successfully.\nOutput JSON file: %v\n", outputFileName)
}
