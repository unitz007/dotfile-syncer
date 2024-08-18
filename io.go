package main

import (
	"fmt"
	"time"
)

func Info(v ...string) {
	s := ""
	for _, c := range v {
		s = s + " " + c
	}

	fmt.Println(time.Now().Format(time.RFC3339), "INFO:", s)
}

func Error(v ...string) {
	s := ""
	for _, c := range v {
		s = s + " " + c
	}

	fmt.Println(time.Now().Format(time.RFC3339), "ERROR:", s)
}
