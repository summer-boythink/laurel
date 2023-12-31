package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/summer-boythink/laurel"
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

	tempTable, err := laurel.DBopen(file.Name())
	if err != nil {
		fmt.Println("laurel open error:", err)
		return nil
	}
	input_buffer := laurel.NewInputBuffer()
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

// TODO
// func TestTableIsFullErrorMessage(t *testing.T) {
// 	insertCommands := []string{}
// 	for i := 1; i <= 66; i++ {
// 		insertCommands = append(insertCommands, fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i))
// 	}
// 	insertCommands = append(insertCommands, ".exit")

// 	result := runScript(t, insertCommands, true)
// 	expectedErrorMessage := "Error: Table full.\n"
// 	if result[len(result)-2] != expectedErrorMessage {
// 		t.Errorf("TestTableIsFullErrorMessage failed, got: %v, want: %v", result[len(result)-2], expectedErrorMessage)
// 	}
// }

func TestInsertingMaxLengthStrings(t *testing.T) {
	longUsername := strings.Repeat("a", 32)
	longEmail := strings.Repeat("a", 255)

	result := runScript(t, []string{
		fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
		"select",
		".exit",
	}, true)
	expected := []string{
		"Executed.\n",
		fmt.Sprintf("(1, %s, %s)\n", longUsername, longEmail),
		"Executed.\n",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestInsertingMaxLengthStrings failed, got: %v, want: %v", result, expected)
	}
}

func TestStringTooLongErrorMessage(t *testing.T) {
	longUsername := strings.Repeat("a", 33)
	longEmail := strings.Repeat("a", 256)

	result := runScript(t, []string{
		fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
		"select",
		".exit",
	}, true)
	expected := []string{
		"String is too long.\n",
		"Executed.\n",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestStringTooLongErrorMessage failed, got: %v, want: %v", result, expected)
	}
}

func TestNegativeIdErrorMessage(t *testing.T) {
	result := runScript(t, []string{
		"insert -1 cstack foo@bar.com",
		"select",
		".exit",
	}, true)
	expected := []string{
		"ID must be positive.\n",
		"Executed.\n",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestNegativeIdErrorMessage failed, got: %v, want: %v", result, expected)
	}
}

func TestDuplicateIdErrorMessage(t *testing.T) {
	result := runScript(t, []string{
		"insert 1 user1 person1@example.com",
		"insert 1 user1 person1@example.com",
		"select",
		".exit",
	}, true)
	expected := []string{
		"Executed.\n",
		"Error: Duplicate key.\n",
		"(1, user1, person1@example.com)\n",
		"Executed.\n",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestDuplicateIdErrorMessage failed, got: %v, want: %v", result, expected)
	}
}

func TestPrintingOneNodeBTreeStructure(t *testing.T) {
	insertCommands := []string{"insert 1 user1 person1@example.com", "insert 2 user2 person2@example.com", "insert 3 user3 person3@example.com", ".btree", ".exit"}

	result := runScript(t, insertCommands, true)
	expected := []string{
		"Executed.\n",
		"Executed.\n",
		"Executed.\n",
		"Tree:\n",
		"- leaf (size 3)\n",
		"  - 1\n",
		"  - 2\n",
		"  - 3\n",
	}
	if strings.Join(result, "") != strings.Join(expected, "") {
		t.Errorf("TestPrintingOneNodeBTreeStructure failed, got: %v, want: %v", result, expected)
	}
}

func TestPrintingThreeLeafNodeBTreeStructure(t *testing.T) {
	insertCommands := make([]string, 0)
	for i := 1; i <= 14; i++ {
		insertCommands = append(insertCommands, fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i))
	}
	insertCommands = append(insertCommands, ".btree")
	insertCommands = append(insertCommands, "insert 15 user15 person15@example.com")
	insertCommands = append(insertCommands, ".exit")

	result := runScript(t, insertCommands, true)
	expected := []string{
		"Tree:\n",
		"- internal (size 1)\n",
		"  - leaf (size 7)\n",
		"    - 1\n",
		"    - 2\n",
		"    - 3\n",
		"    - 4\n",
		"    - 5\n",
		"    - 6\n",
		"    - 7\n",
		"  - key 7\n",
		"  - leaf (size 7)\n",
		"    - 8\n",
		"    - 9\n",
		"    - 10\n",
		"    - 11\n",
		"    - 12\n",
		"    - 13\n",
		"    - 14\n",
		"Executed.\n",
	}

	if strings.Join(result[14:], "") != strings.Join(expected, "") {
		t.Errorf("TestPrintingThreeLeafNodeBTreeStructure failed, got: %v, want: %v", result[14:], expected)
	}
}

func TestPrintingConstants(t *testing.T) {
	insertCommands := []string{".constants", ".exit"}

	result := runScript(t, insertCommands, true)
	expected := []string{
		"Constants:\n",
		"ROW_SIZE: 291\n",
		"COMMON_NODE_HEADER_SIZE: 6\n",
		"LEAF_NODE_HEADER_SIZE: 14\n",
		"LEAF_NODE_CELL_SIZE: 295\n",
		"LEAF_NODE_SPACE_FOR_CELLS: 4082\n",
		"LEAF_NODE_MAX_CELLS: 13\n",
	}
	if strings.Join(result, "") != strings.Join(expected, "") {
		t.Errorf("TestPrintingConstants failed, got: %v, want: %v", result, expected)
	}
}

func TestPrintBTreeStructure(t *testing.T) {
	insertCommands := []string{
		"insert 18 user18 person18@example.com",
		"insert 7 user7 person7@example.com",
		"insert 10 user10 person10@example.com",
		"insert 29 user29 person29@example.com",
		"insert 23 user23 person23@example.com",
		"insert 4 user4 person4@example.com",
		"insert 14 user14 person14@example.com",
		"insert 30 user30 person30@example.com",
		"insert 15 user15 person15@example.com",
		"insert 26 user26 person26@example.com",
		"insert 22 user22 person22@example.com",
		"insert 19 user19 person19@example.com",
		"insert 2 user2 person2@example.com",
		"insert 1 user1 person1@example.com",
		"insert 21 user21 person21@example.com",
		"insert 11 user11 person11@example.com",
		"insert 6 user6 person6@example.com",
		"insert 20 user20 person20@example.com",
		"insert 5 user5 person5@example.com",
		"insert 8 user8 person8@example.com",
		"insert 9 user9 person9@example.com",
		"insert 3 user3 person3@example.com",
		"insert 12 user12 person12@example.com",
		"insert 27 user27 person27@example.com",
		"insert 17 user17 person17@example.com",
		"insert 16 user16 person16@example.com",
		"insert 13 user13 person13@example.com",
		"insert 24 user24 person24@example.com",
		"insert 25 user25 person25@example.com",
		"insert 28 user28 person28@example.com",
		".btree",
		".exit",
	}
	result := runScript(t, insertCommands, true)
	expected := []string{
		"Tree:\n",
		"- internal (size 3)\n",
		"  - leaf (size 7)\n",
		"    - 1\n",
		"    - 2\n",
		"    - 3\n",
		"    - 4\n",
		"    - 5\n",
		"    - 6\n",
		"    - 7\n",
		"  - key 7\n",
		"  - leaf (size 8)\n",
		"    - 8\n",
		"    - 9\n",
		"    - 10\n",
		"    - 11\n",
		"    - 12\n",
		"    - 13\n",
		"    - 14\n",
		"    - 15\n",
		"  - key 15\n",
		"  - leaf (size 7)\n",
		"    - 16\n",
		"    - 17\n",
		"    - 18\n",
		"    - 19\n",
		"    - 20\n",
		"    - 21\n",
		"    - 22\n",
		"  - key 22\n",
		"  - leaf (size 8)\n",
		"    - 23\n",
		"    - 24\n",
		"    - 25\n",
		"    - 26\n",
		"    - 27\n",
		"    - 28\n",
		"    - 29\n",
		"    - 30\n",
	}
	if strings.Join(result[30:], "") != strings.Join(expected, "") {
		t.Errorf("TestPrintBTreeStructure failed, got: %v, want: %v", result[30:], expected)
	}
}
