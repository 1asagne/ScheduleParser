package logger

import (
	"log"
	"os"
)

const flag = log.Ldate | log.Ltime | log.Lshortfile

var (
	Info    = log.New(os.Stdout, "INFO: ", flag)
	Warning = log.New(os.Stdout, "WARNING: ", flag)
	Error   = log.New(os.Stderr, "ERROR: ", flag)
)
