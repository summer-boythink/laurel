package laurel

import "fmt"

type Cursor struct {
	table      *Table
	pageNum    uint32
	cellNum    uint32
	endOfTable bool // Indicates a position one past the last element
}

func (cursor *Cursor) leafNodeInsert(key uint32, value *Row) {
	node := cursor.table.pager.getPage(cursor.pageNum)

	numCells := *leafNodeNumCells(node)
	if numCells >= LEAF_NODE_MAX_CELLS {
		// Node full
		cursor.leafNodeSplitAndInsert(key, value)
		return
	}

	if cursor.cellNum < numCells {
		// Make room for new cell
		for i := numCells; i > cursor.cellNum; i-- {
			copy(leafNodeCell(node, i)[:], leafNodeCell(node, i-1)[:])
		}
	}

	*leafNodeNumCells(node) += 1
	*leafNodeKey(node, cursor.cellNum) = key
	serializeRow(value, leafNodeValue(node, cursor.cellNum)[:])
}

func (cursor *Cursor) leafNodeSplitAndInsert(key uint32, value *Row) {
	// Create a new node and move half the cells over.
	// Insert the new value in one of the two nodes.
	// Update parent or create a new parent.

	oldNode := cursor.table.pager.getPage(cursor.pageNum)
	newPageNum := cursor.table.pager.getUnusedPageNum()
	newNode := cursor.table.pager.getPage(newPageNum)
	initializeLeafNode(newNode)

	// All existing keys plus new key should be divided
	// evenly between old (left) and new (right) nodes.
	// Starting from the right, move each key to correct position.
	for i := int(LEAF_NODE_MAX_CELLS); i >= 0; i-- {
		var destinationNode *[PAGE_SIZE]byte
		if i >= int(LEAF_NODE_LEFT_SPLIT_COUNT) {
			destinationNode = newNode
		} else {
			destinationNode = oldNode
		}
		indexWithinNode := i % int(LEAF_NODE_LEFT_SPLIT_COUNT)
		destination := leafNodeCell(destinationNode, uint32(indexWithinNode))

		if i == int(cursor.cellNum) {
			serializeRow(value, destination[:])
		} else if i > int(cursor.cellNum) {
			copy(destination[:], leafNodeCell(oldNode, uint32(i-1))[:])
		} else {
			copy(destination[:], leafNodeCell(oldNode, uint32(i))[:])
		}
	}

	// Update cell count on both leaf nodes
	*leafNodeNumCells(oldNode) = LEAF_NODE_LEFT_SPLIT_COUNT
	*leafNodeNumCells(newNode) = LEAF_NODE_RIGHT_SPLIT_COUNT

	if isNodeRoot(oldNode) {
		cursor.table.createNewRoot(newPageNum)
	} else {
		fmt.Println("Need to implement updating parent after split")
		// You can handle updating parent logic here
	}
}

func (cursor *Cursor) cursorAdvance() {
	pageNum := cursor.pageNum
	node := cursor.table.pager.getPage(pageNum)

	cursor.cellNum += 1
	if cursor.cellNum >= *leafNodeNumCells(node) {
		cursor.endOfTable = true
	}
}

func (cursor *Cursor) cursorValue() *[291]byte {
	pageNum := cursor.pageNum
	page := cursor.table.pager.getPage(pageNum)
	return leafNodeValue(page, cursor.cellNum)
}
