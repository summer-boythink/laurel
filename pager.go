package laurel

import (
	"fmt"
	"os"
)

type Pager struct {
	file_descriptor *os.File
	file_length     uint32
	pages           [TABLE_MAX_PAGES][PAGE_SIZE]byte
	numPages        uint32
}

func (pager *Pager) getPage(page_num uint32) *[PAGE_SIZE]byte {
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
		if page_num >= pager.numPages {
			pager.numPages = page_num + 1
		}
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

func (pager *Pager) getUnusedPageNum() uint32 {
	return pager.numPages
}

func (pager *Pager) printTree(pageNum uint32, indentationLevel uint32) {
	node := pager.getPage(pageNum)
	numKeys, child := uint32(0), uint32(0)

	switch getNodeType(node) {
	case NODE_LEAF:
		numKeys = *leafNodeNumCells(node)
		indent(indentationLevel)
		PrintMsgf("- leaf (size %d)\n", numKeys)
		for i := uint32(0); i < numKeys; i++ {
			indent(indentationLevel + 1)
			PrintMsgf("- %d\n", *leafNodeKey(node, i))
		}
	case NODE_INTERNAL:
		numKeys = *internalNodeNumKeys(node)
		indent(indentationLevel)
		PrintMsgf("- internal (size %d)\n", numKeys)
		for i := uint32(0); i < numKeys; i++ {
			child = *internalNodeChild(node, i)
			pager.printTree(child, indentationLevel+1)

			indent(indentationLevel + 1)
			PrintMsgf("- key %d\n", *internalNodeKey(node, i))
		}
		child = *internalNodeRightChild(node)
		pager.printTree(child, indentationLevel+1)
	}
}

func pager_open(filename string) (*Pager, error) {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Unable to open file\n")
		return nil, err
	}

	fileInfo, err := fd.Stat()
	if err != nil {
		fmt.Printf("Unable to get file stats\n")
		return nil, err
	}

	fileLength := fileInfo.Size()

	pager := &Pager{
		file_descriptor: fd,
		file_length:     uint32(fileLength),
		numPages:        uint32(fileLength) / PAGE_SIZE,
	}

	return pager, nil
}
