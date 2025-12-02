package main

import (
	"log"
	"os"
)

func init() {
	app.logger = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.errLogger = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
}
