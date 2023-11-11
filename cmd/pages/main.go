package main

import (
	"fmt"
	"os"

	"github.com/summer-boythink/laurel"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Must supply a database filename.\n")
		os.Exit(1)
	}
	filename := os.Args[1]
	table, err := laurel.DBopen(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	input_buffer := laurel.NewInputBuffer()

	for msg := range laurel.Run(table, input_buffer) {
		fmt.Print(msg)
	}
}
