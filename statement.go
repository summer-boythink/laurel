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
	stype       StatementType
	rowToInsert Row // only used by insert statement
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
	node := table.pager.getPage(table.rootPageNum)
	numCells := *leafNodeNumCells(node)

	rowToInsert := &statement.rowToInsert
	keyToInsert := rowToInsert.id
	cursor := table.tableFind(keyToInsert)
	if cursor == nil {
		// TODO:
		return EXECUTE_SUCCESS
	}
	if cursor.cellNum < numCells {
		keyAtIndex := *leafNodeKey(node, cursor.cellNum)
		if keyAtIndex == keyToInsert {
			return EXECUTE_DUPLICATE_KEY
		}
	}

	cursor.leafNodeInsert(rowToInsert.id, rowToInsert)

	return EXECUTE_SUCCESS
}

func execute_select(statement *Statement, table *Table) ExecuteResult {
	cursor := table.tableStart()

	var row Row
	for !cursor.endOfTable {
		deserializeRow(cursor.cursorValue()[:], &row)
		print_row(&row)
		cursor.cursorAdvance()
	}

	return EXECUTE_SUCCESS
}
