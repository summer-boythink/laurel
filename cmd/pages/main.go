package main

import (
	"fmt"
	"os"

	"github.com/summer-boythink/laurel"
)

func new_input_buffer() *laurel.InputBuffer {
	return &laurel.InputBuffer{}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Must supply a database filename.\n")
		os.Exit(1)
	}

	filename := os.Args[1]
	table := laurel.DBopen(filename)

	input_buffer := new_input_buffer()

	for msg := range laurel.Run(table, input_buffer) {
		fmt.Print(msg)
	}
}
