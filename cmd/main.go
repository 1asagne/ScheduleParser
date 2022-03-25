// Main package provides entrypoint to use schedule parser as executable.

package main

import (
	"flag"
	"fmt"

	"github.com/1asagne/scheduleparser"
)

// Main function reads command-line arguments,
// parses input and output files paths and uses ParseFile from scheduleparser package.
func main() {
	var inputFilePath, outputFilePath string
	flag.StringVar(&inputFilePath, "i", "", "Input pdf file path")
	flag.StringVar(&outputFilePath, "o", "", "Output json file path")
	flag.Parse()

	if inputFilePath == "" || outputFilePath == "" {
		flag.PrintDefaults()
		return
	}

	err := scheduleparser.ParseFile(inputFilePath, outputFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Parsing completed successfully.\nOutput JSON file: %v\n", outputFilePath)
}
