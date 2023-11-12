package laurel

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
)

type PrepareResult int

const (
	PREPARE_SUCCESS PrepareResult = iota
	PREPARE_NEGATIVE_ID
	PREPARE_STRING_TOO_LONG
	PREPARE_SYNTAX_ERROR
	PREPARE_UNRECOGNIZED_STATEMENT
)

type MetaCommandResult int

const (
	META_COMMAND_SUCCESS MetaCommandResult = iota
	META_COMMAND_UNRECOGNIZED_COMMAND
	META_COMMAND_EXIT
)

type ExecuteResult int

const (
	EXECUTE_SUCCESS ExecuteResult = iota
	EXECUTE_DUPLICATE_KEY
	EXECUTE_TABLE_FULL
)

type InputBuffer struct {
	buffer        []byte
	buffer_length int
	input_length  int
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{}
}

func (i *InputBuffer) clear() {
	i.buffer = nil
	i.buffer_length = 0
	i.input_length = 0
}

func prepare_insert(input_buffer *InputBuffer, statement *Statement) PrepareResult {
	statement.stype = STATEMENT_INSERT

	tokens := bytes.Split(input_buffer.buffer, []byte(" "))
	if len(tokens) < 4 {
		return PREPARE_SYNTAX_ERROR
	}

	id, err := strconv.Atoi(string(tokens[1]))
	if err != nil || id < 0 {
		return PREPARE_NEGATIVE_ID
	}
	username := tokens[2]
	email := tokens[3]

	if len(username) > COLUMN_USERNAME_SIZE || len(email) > COLUMN_EMAIL_SIZE {
		return PREPARE_STRING_TOO_LONG
	}

	statement.rowToInsert.id = uint32(id)
	copy(statement.rowToInsert.username[:], username)
	copy(statement.rowToInsert.email[:], email)

	return PREPARE_SUCCESS
}

func prepare_statement(input_buffer *InputBuffer, statement *Statement) PrepareResult {
	if bytes.HasPrefix(input_buffer.buffer, []byte("insert")) {
		return prepare_insert(input_buffer, statement)
	}
	if bytes.Equal(input_buffer.buffer, []byte("select")) {
		statement.stype = STATEMENT_SELECT
		return PREPARE_SUCCESS
	}

	return PREPARE_UNRECOGNIZED_STATEMENT
}

func (input_buffer *InputBuffer) ReadInput() {
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		fmt.Printf("Error reading input %v\n", err)
		os.Exit(1)
	}

	input_buffer.buffer = line
	input_buffer.input_length = len(line)
	input_buffer.buffer_length = len(line)
}

func (i *InputBuffer) SetInput(s string) {
	i.buffer = []byte(s)
	i.input_length = len(s)
	i.buffer_length = len(s)
}

func print_prompt() { PrintMsgf("db > ") }

func doMetaCommand(inputBuffer *InputBuffer, table *Table) MetaCommandResult {
	if bytes.Equal(inputBuffer.buffer, []byte(".exit")) {
		table.db_close()
		return META_COMMAND_EXIT
	} else if bytes.Equal(inputBuffer.buffer, []byte(".btree")) {
		PrintMsgf("Tree:\n")
		table.pager.printTree(0, 0)
		return META_COMMAND_SUCCESS
	} else if bytes.Equal(inputBuffer.buffer, []byte(".constants")) {
		PrintMsgf("Constants:\n")
		PrintMsgf(printConstants())
		return META_COMMAND_SUCCESS
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}
}

var (
	ResMsg = make(chan string)
)

func Run(table *Table, input_buffer *InputBuffer, opt ...Option) <-chan string {
	// process options
	opts := loadOptions(opt...)
	if opts.ResMsg != nil {
		ResMsg = opts.ResMsg
	}

	go func() {
		defer close(ResMsg)
	loop:
		for {
			if opts.IsTestCmd {
				input_buffer.SetInput(<-opts.InputCmd)
			} else {
				print_prompt()
				input_buffer.ReadInput()
			}
			if input_buffer.buffer[0] == '.' {
				switch doMetaCommand(input_buffer, table) {
				case META_COMMAND_SUCCESS:
					continue
				case META_COMMAND_UNRECOGNIZED_COMMAND:
					PrintMsgf("Unrecognized command '%s'\n", input_buffer.buffer)
					continue
				case META_COMMAND_EXIT:
					break loop
				}
			}

			statement := &Statement{}
			switch prepare_statement(input_buffer, statement) {
			case PREPARE_SUCCESS:
				// break
			case PREPARE_NEGATIVE_ID:
				PrintMsgf("ID must be positive.\n")
				continue
			case PREPARE_STRING_TOO_LONG:
				PrintMsgf("String is too long.\n")
				continue
			case PREPARE_SYNTAX_ERROR:
				PrintMsgf("Syntax error. Could not parse statement.\n")
				continue
			case PREPARE_UNRECOGNIZED_STATEMENT:
				PrintMsgf("Unrecognized keyword at start of '%s'.\n",
					input_buffer.buffer)
				continue
			default:
				PrintMsgf("err '%v'.\n", statement.stype)
				os.Exit(1)
			}

			switch statement.execute_statement(table) {
			case EXECUTE_SUCCESS:

			case EXECUTE_TABLE_FULL:
				PrintMsgf("Error: Table full.\n")
			case EXECUTE_DUPLICATE_KEY:
				PrintMsgf("Error: Duplicate key.\n")
				continue
			default:
				PrintMsgf("Error executing statement.\n")
				os.Exit(1)
			}
			PrintMsgf("Executed.\n")
			input_buffer.clear()
		}
	}()
	return ResMsg
}
