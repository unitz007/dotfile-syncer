package main

import (
	"fmt"
)

func Info(v ...string) {
	s := ""
	for _, c := range v {
		s = s + " " + c
	}

	fmt.Println("INFO:", s)
}

func Error(v ...string) {
	s := ""
	for _, c := range v {
		s = s + " " + c
	}

	fmt.Println("ERROR:", s)
}
