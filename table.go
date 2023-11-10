package laurel

import (
	"fmt"
	"os"
)

type Table struct {
	pager    *Pager
	num_rows uint32
}

func (table *Table) row_slot(row_num uint32) []byte {
	page_num := row_num / ROWS_PER_PAGE
	page := table.pager.get_page(page_num)
	row_offset := row_num % ROWS_PER_PAGE
	byte_offset := row_offset * ROW_SIZE

	return page[byte_offset : byte_offset+ROW_SIZE]
}

func (table *Table) db_close() {
	pager := table.pager
	num_full_pages := table.num_rows / ROWS_PER_PAGE

	for i := uint32(0); i < num_full_pages; i++ {
		if IsZeroPage(pager.pages[i]) {
			continue
		}
		pager.pager_flush(i, PAGE_SIZE)
		SetZeroPage(pager.pages[i])
	}

	num_additional_rows := table.num_rows % ROWS_PER_PAGE
	if num_additional_rows > 0 {
		page_num := num_full_pages
		if !IsZeroPage(pager.pages[page_num]) {
			pager.pager_flush(page_num, num_additional_rows*ROW_SIZE)
			SetZeroPage(pager.pages[page_num])
		}
	}

	err := pager.file_descriptor.Close()
	if err != nil {
		fmt.Printf("Error closing db file.\n")
		os.Exit(1)
	}
	for i := uint32(0); i < TABLE_MAX_PAGES; i++ {
		page := pager.pages[i]
		if !IsZeroPage(page) {
			SetZeroPage(page)
		}
	}
}

func DBopen(filename string) *Table {
	pager := pager_open(filename)
	num_rows := pager.file_length / ROW_SIZE

	table := &Table{
		pager:    pager,
		num_rows: num_rows,
	}

	return table
}
