package laurel

import (
	"fmt"
	"os"
)

type Pager struct {
	file_descriptor *os.File
	file_length     uint32
	pages           [TABLE_MAX_PAGES][PAGE_SIZE]byte
}

func (pager *Pager) get_page(page_num uint32) *[PAGE_SIZE]byte {
	if page_num > TABLE_MAX_PAGES {
		fmt.Printf("Tried to fetch page number out of bounds. %d > %d\n", page_num,
			TABLE_MAX_PAGES)
		os.Exit(1)
	}

	if IsZeroPage(pager.pages[page_num]) {
		page := make([]byte, PAGE_SIZE)
		num_pages := pager.file_length / PAGE_SIZE

		if pager.file_length%PAGE_SIZE != 0 {
			num_pages += 1
		}

		if page_num <= uint32(num_pages) {
			pager.file_descriptor.Seek(int64(page_num)*PAGE_SIZE, 0)
			bytes_read, err := pager.file_descriptor.Read(page)
			if bytes_read == -1 {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
			}
		}
		CopyPage(&pager.pages[page_num], page[:])
	}

	return &pager.pages[page_num]
}

func (pager *Pager) pager_flush(page_num uint32, size uint32) {
	if IsZeroPage(pager.pages[page_num]) {
		fmt.Printf("Tried to flush null page\n")
		os.Exit(1)
	}

	pager.file_descriptor.Seek(int64(page_num)*PAGE_SIZE, 0)

	bytes_written, err := pager.file_descriptor.Write(pager.pages[page_num][:size])
	if err != nil || bytes_written == -1 {
		fmt.Printf("Error writing: %v\n", err)
		os.Exit(1)
	}
}

func pager_open(filename string) *Pager {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Unable to open file\n")
		os.Exit(1)
	}

	fileInfo, err := fd.Stat()
	if err != nil {
		fmt.Printf("Unable to get file stats\n")
		os.Exit(1)
	}

	file_length := fileInfo.Size()

	pager := &Pager{
		file_descriptor: fd,
		file_length:     uint32(file_length),
	}

	return pager
}
