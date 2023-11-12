package laurel

import (
	"fmt"
	"os"
)

type Table struct {
	pager       *Pager
	rootPageNum uint32
}

func (t *Table) row_slot(row_num uint32) []byte {
	page_num := row_num / ROWS_PER_PAGE
	page := t.pager.getPage(page_num)
	row_offset := row_num % ROWS_PER_PAGE
	byte_offset := row_offset * ROW_SIZE

	return page[byte_offset : byte_offset+ROW_SIZE]
}

func (t *Table) db_close() {
	pager := t.pager
	for i := uint32(0); i < t.pager.numPages; i++ {
		if !IsZeroPage(pager.pages[i]) {
			pager.pager_flush(i, PAGE_SIZE)
		}
		SetZeroPage(pager.pages[i])
	}

	err := pager.file_descriptor.Close()
	if err != nil {
		fmt.Printf("Error closing db file.\n")
		os.Exit(1)
	}
	for i := uint32(0); i < TABLE_MAX_PAGES; i++ {
		page := pager.pages[i]
		SetZeroPage(page)
	}
}

func (t *Table) tableFind(key uint32) *Cursor {
	rootPageNum := t.rootPageNum
	rootNode := t.pager.getPage(rootPageNum)

	if getNodeType(rootNode) == NODE_LEAF {
		return t.leafNodeFind(rootPageNum, key)
	} else {
		return t.internalNodeFind(rootPageNum, key)
	}
}

func (table *Table) internalNodeFind(pageNum uint32, key uint32) *Cursor {
	node := table.pager.getPage(pageNum)
	numKeys := *internalNodeNumKeys(node)

	// Binary search to find index of child to search
	minIndex := uint32(0)
	maxIndex := numKeys // there is one more child than key

	for minIndex != maxIndex {
		index := (minIndex + maxIndex) / 2
		keyToRight := *internalNodeKey(node, index)
		if keyToRight >= key {
			maxIndex = index
		} else {
			minIndex = index + 1
		}
	}

	childNum := *internalNodeChild(node, minIndex)
	child := table.pager.getPage(childNum)
	switch getNodeType(child) {
	case NODE_LEAF:
		return table.leafNodeFind(childNum, key)
	case NODE_INTERNAL:
		return table.internalNodeFind(childNum, key)
	}
	return nil
}

func (t *Table) leafNodeFind(pageNum uint32, key uint32) *Cursor {
	node := t.pager.getPage(pageNum)
	numCells := *leafNodeNumCells(node)

	cursor := &Cursor{}
	cursor.table = t
	cursor.pageNum = pageNum

	// Binary search
	minIndex := uint32(0)
	onePastMaxIndex := numCells
	for onePastMaxIndex != minIndex {
		index := (minIndex + onePastMaxIndex) / 2
		keyAtIndex := *leafNodeKey(node, index)
		if key == keyAtIndex {
			cursor.cellNum = index
			return cursor
		}
		if key < keyAtIndex {
			onePastMaxIndex = index
		} else {
			minIndex = index + 1
		}
	}

	cursor.cellNum = minIndex
	return cursor
}

func (t *Table) createNewRoot(rightChildPageNum uint32) {
	// Handle splitting the root.
	// Old root copied to new page, becomes left child.
	// Address of right child passed in.
	// Re-initialize root page to contain the new root node.
	// New root node points to two children.

	root := t.pager.getPage(t.rootPageNum)
	_ = t.pager.getPage(rightChildPageNum)
	leftChildPageNum := t.pager.getUnusedPageNum()
	leftChild := t.pager.getPage(leftChildPageNum)

	// Left child has data copied from old root
	copy(leftChild[:], root[:])
	setNodeRoot(leftChild, false)

	// Root node is a new internal node with one key and two children
	initializeInternalNode(root)
	setNodeRoot(root, true)
	*internalNodeNumKeys(root) = 1
	*internalNodeChild(root, 0) = leftChildPageNum
	leftChildMaxKey := getNodeMaxKey(leftChild)
	*internalNodeKey(root, 0) = leftChildMaxKey
	*internalNodeRightChild(root) = rightChildPageNum
}

func (t *Table) tableStart() *Cursor {
	cursor := t.tableFind(0)

	node := t.pager.getPage(cursor.pageNum)
	numCells := leafNodeNumCells(node)
	cursor.endOfTable = *numCells == 0
	return cursor
}

func DBopen(filename string) (*Table, error) {
	pager, err := pager_open(filename)
	if err != nil {
		return nil, err
	}

	t := &Table{pager: pager, rootPageNum: 0}
	if pager.numPages == 0 {
		rootPage := t.pager.getPage(0)
		initializeLeafNode(rootPage)
		setNodeRoot(rootPage, true)
	}

	return t, nil
}
