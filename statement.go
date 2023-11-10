package laurel

import (
	"fmt"
	"os"
)

type StatementType int

const (
	STATEMENT_INSERT StatementType = iota
	STATEMENT_SELECT
)

type Statement struct {
	stype         StatementType
	row_to_insert Row // only used by insert statement
}

func (statement *Statement) execute_statement(table *Table) ExecuteResult {
	switch statement.stype {
	case STATEMENT_INSERT:
		return execute_insert(statement, table)
	case STATEMENT_SELECT:
		return execute_select(statement, table)
	default:
		fmt.Printf("Unrecognized keyword at start of '%v'.\n", statement.stype)
		os.Exit(1)
	}
	return EXECUTE_SUCCESS
}

func execute_insert(statement *Statement, table *Table) ExecuteResult {
	if table.num_rows >= TABLE_MAX_ROWS {
		return EXECUTE_TABLE_FULL
	}

	row_to_insert := &statement.row_to_insert

	serialize_row(row_to_insert, table.row_slot(table.num_rows))
	table.num_rows += 1

	return EXECUTE_SUCCESS
}

func execute_select(statement *Statement, table *Table) ExecuteResult {
	row := &Row{}
	for i := uint32(0); i < table.num_rows; i++ {
		deserialize_row(table.row_slot(i), row)
		print_row(row)
	}
	return EXECUTE_SUCCESS
}
