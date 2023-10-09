package log

import (
	"log"
)

func Print(args ...interface{}) {
	log.Print(args...)
}

func Printf(template string, args ...interface{}) {
	log.Printf(template, args...)
}
