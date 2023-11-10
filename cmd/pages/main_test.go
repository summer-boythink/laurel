package main

import (
	"fmt"
	"github.com/summer-boythink/laurel"
	"os"
	"reflect"
	"testing"
)

func runScript(t *testing.T, commands []string, isPrevDel bool) []string {
	var output []string
	var file *os.File

	if _, err := os.Stat("test.db"); os.IsNotExist(err) {
		file, err = os.Create("test.db")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return nil
		}
	} else if err == nil && isPrevDel {
		err := os.Remove("test.db")
		if err != nil {
			fmt.Println("Error deleting file:", err)
			return nil
		}
		file, err = os.Create("test.db")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return nil
		}
	} else {
		file, err = os.Open("test.db")
		if err != nil {
			fmt.Println("Error open file:", err)
			return nil
		}
	}

	tempTable := laurel.DBopen(file.Name())
	input_buffer := new_input_buffer()
	cmd := make(chan string)
	go func() {
		defer close(cmd)
		for _, v := range commands {
			cmd <- v
		}
	}()

	opt := laurel.Options{}
	opt.ResMsg = make(chan string)
	opt.IsTestCmd = true
	opt.InputCmd = cmd

	for msg := range laurel.Run(tempTable, input_buffer, laurel.WithOptions(opt)) {
		output = append(output, msg)
	}
	return output
}
func TestInsertAndRetrieveRow(t *testing.T) {
	result := runScript(t, []string{
		"insert 1 user1 person1@example.com",
		"select",
		".exit",
	}, true)
	expected := []string{
		"Executed.\n",
		"(1, user1, person1@example.com)\n",
		"Executed.\n",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestInsertAndRetrieveRow failed, got: %v, want: %v", result, expected)
	}
}

func TestKeepDataAfterClosingConnection(t *testing.T) {
	result1 := runScript(t, []string{
		"insert 1 user1 person1@example.com",
		".exit",
	}, true)
	expected1 := []string{
		"Executed.\n",
	}
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("TestKeepDataAfterClosingConnection (part 1) failed, got: %v, want: %v", result1, expected1)
	}

	result2 := runScript(t, []string{
		"select",
		".exit",
	}, false)
	expected2 := []string{
		"(1, user1, person1@example.com)\n",
		"Executed.\n",
	}
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("TestKeepDataAfterClosingConnection (part 2) failed, got: %v, want: %v", result2, expected2)
	}
}
