package main

import (
	"log"
	"os"
)

var (
	Logger    = log.New(os.Stdout, "shh: ", log.LstdFlags)
	ErrLogger = log.New(os.Stderr, "shh: ", log.LstdFlags)
)
