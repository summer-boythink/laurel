package laurel

import (
	"fmt"
	"unsafe"
)

type NodeType uint8

const (
	NODE_INTERNAL NodeType = iota
	NODE_LEAF
)

const (
	/*
	 * Common Node Header Layout
	 */
	NODE_TYPE_SIZE          = 1
	NODE_TYPE_OFFSET        = 0
	IS_ROOT_SIZE            = 1
	IS_ROOT_OFFSET          = NODE_TYPE_SIZE
	PARENT_POINTER_SIZE     = 4
	PARENT_POINTER_OFFSET   = IS_ROOT_OFFSET + IS_ROOT_SIZE
	COMMON_NODE_HEADER_SIZE = NODE_TYPE_SIZE + IS_ROOT_SIZE + PARENT_POINTER_SIZE
	/*
	 * Internal Node Header Layout
	 */
	INTERNAL_NODE_NUM_KEYS_SIZE      = 4
	INTERNAL_NODE_NUM_KEYS_OFFSET    = COMMON_NODE_HEADER_SIZE
	INTERNAL_NODE_RIGHT_CHILD_SIZE   = 4
	INTERNAL_NODE_RIGHT_CHILD_OFFSET = INTERNAL_NODE_NUM_KEYS_OFFSET + INTERNAL_NODE_NUM_KEYS_SIZE
	INTERNAL_NODE_HEADER_SIZE        = COMMON_NODE_HEADER_SIZE + INTERNAL_NODE_NUM_KEYS_SIZE + INTERNAL_NODE_RIGHT_CHILD_SIZE
	/*
	 * Internal Node Body Layout
	 */
	INTERNAL_NODE_KEY_SIZE   = 4
	INTERNAL_NODE_CHILD_SIZE = 4
	INTERNAL_NODE_CELL_SIZE  = INTERNAL_NODE_CHILD_SIZE + INTERNAL_NODE_KEY_SIZE
	/*
	 * Leaf Node Header Layout
	 */
	LEAF_NODE_NUM_CELLS_SIZE   = 4
	LEAF_NODE_NUM_CELLS_OFFSET = COMMON_NODE_HEADER_SIZE
	LEAF_NODE_NEXT_LEAF_SIZE   = 4
	LEAF_NODE_NEXT_LEAF_OFFSET = LEAF_NODE_NUM_CELLS_OFFSET + LEAF_NODE_NUM_CELLS_SIZE
	LEAF_NODE_HEADER_SIZE      = COMMON_NODE_HEADER_SIZE + LEAF_NODE_NUM_CELLS_SIZE + LEAF_NODE_NEXT_LEAF_SIZE

	/*
	 * Leaf Node Body Layout
	 */
	LEAF_NODE_KEY_SIZE          = 4
	LEAF_NODE_KEY_OFFSET        = 0
	LEAF_NODE_VALUE_SIZE        = ROW_SIZE
	LEAF_NODE_VALUE_OFFSET      = LEAF_NODE_KEY_OFFSET + LEAF_NODE_KEY_SIZE
	LEAF_NODE_CELL_SIZE         = LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE
	LEAF_NODE_SPACE_FOR_CELLS   = PAGE_SIZE - LEAF_NODE_HEADER_SIZE
	LEAF_NODE_MAX_CELLS         = LEAF_NODE_SPACE_FOR_CELLS / LEAF_NODE_CELL_SIZE
	LEAF_NODE_RIGHT_SPLIT_COUNT = (LEAF_NODE_MAX_CELLS + 1) / 2
	LEAF_NODE_LEFT_SPLIT_COUNT  = (LEAF_NODE_MAX_CELLS + 1) - LEAF_NODE_RIGHT_SPLIT_COUNT
)

func getNodeType(node *[PAGE_SIZE]byte) NodeType {
	value := node[NODE_TYPE_OFFSET]
	return NodeType(value)
}

func setNodeType(node *[PAGE_SIZE]byte, typ NodeType) {
	node[NODE_TYPE_OFFSET] = uint8(typ)
}

func isNodeRoot(node *[PAGE_SIZE]byte) bool {
	value := node[IS_ROOT_OFFSET]
	return value != 0
}

func setNodeRoot(node *[PAGE_SIZE]byte, isRoot bool) {
	if isRoot {
		node[IS_ROOT_OFFSET] = 1
	} else {
		node[IS_ROOT_OFFSET] = 0
	}
}

func internalNodeNumKeys(node *[PAGE_SIZE]byte) *uint32 {
	return (*uint32)(unsafe.Pointer(&node[INTERNAL_NODE_NUM_KEYS_OFFSET]))
}

func internalNodeRightChild(node *[PAGE_SIZE]byte) *uint32 {
	return (*uint32)(unsafe.Pointer(&node[INTERNAL_NODE_RIGHT_CHILD_OFFSET]))
}

func internalNodeCell(node *[PAGE_SIZE]byte, cellNum uint32) *uint32 {
	offset := INTERNAL_NODE_HEADER_SIZE + (cellNum * INTERNAL_NODE_CELL_SIZE)
	return (*uint32)(unsafe.Pointer(&node[offset]))
}

func internalNodeChild(node *[PAGE_SIZE]byte, childNum uint32) *uint32 {
	numKeys := *internalNodeNumKeys(node)
	if childNum > numKeys {
		// handle error
	}
	if childNum == numKeys {
		return internalNodeRightChild(node)
	}
	return internalNodeCell(node, childNum)
}

func internalNodeKey(node *[PAGE_SIZE]byte, keyNum uint32) *uint32 {
	cellPtr := internalNodeCell(node, keyNum)
	return (*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(cellPtr)) + uintptr(INTERNAL_NODE_CHILD_SIZE)))
}

func leafNodeNumCells(node *[PAGE_SIZE]byte) *uint32 {
	return (*uint32)(unsafe.Pointer(&node[LEAF_NODE_NUM_CELLS_OFFSET]))
}

func leafNodeCell(node *[PAGE_SIZE]byte, cellNum uint32) *[LEAF_NODE_CELL_SIZE]byte {
	offset := LEAF_NODE_HEADER_SIZE + (cellNum * LEAF_NODE_CELL_SIZE)
	ptr := unsafe.Pointer(&node[offset])
	return (*[LEAF_NODE_CELL_SIZE]byte)(ptr)
}

func leafNodeKey(node *[PAGE_SIZE]byte, cellNum uint32) *uint32 {
	cell := leafNodeCell(node, cellNum)
	return (*uint32)(unsafe.Pointer(&cell[0]))
}

func leafNodeValue(node *[PAGE_SIZE]byte, cellNum uint32) *[LEAF_NODE_CELL_SIZE - LEAF_NODE_KEY_SIZE]byte {
	cell := leafNodeCell(node, cellNum)
	return (*[LEAF_NODE_CELL_SIZE - LEAF_NODE_KEY_SIZE]byte)(unsafe.Pointer(&cell[LEAF_NODE_KEY_SIZE]))
}

func leafNodeNextLeaf(node *[PAGE_SIZE]byte) *uint32 {
	return (*uint32)(unsafe.Pointer(&node[LEAF_NODE_NEXT_LEAF_OFFSET]))
}

func getNodeMaxKey(node *[PAGE_SIZE]byte) uint32 {
	switch getNodeType(node) {
	case NODE_INTERNAL:
		return *internalNodeKey(node, *internalNodeNumKeys(node)-1)
	case NODE_LEAF:
		return *leafNodeKey(node, *leafNodeNumCells(node)-1)
	default:
		// handle unknown node type
	}
	return 0
}

func printConstants() string {
	s := ""
	s += fmt.Sprintf("ROW_SIZE: %d\n", ROW_SIZE)
	s += fmt.Sprintf("COMMON_NODE_HEADER_SIZE: %d\n", COMMON_NODE_HEADER_SIZE)
	s += fmt.Sprintf("LEAF_NODE_HEADER_SIZE: %d\n", LEAF_NODE_HEADER_SIZE)
	s += fmt.Sprintf("LEAF_NODE_CELL_SIZE: %d\n", LEAF_NODE_CELL_SIZE)
	s += fmt.Sprintf("LEAF_NODE_SPACE_FOR_CELLS: %d\n", LEAF_NODE_SPACE_FOR_CELLS)
	s += fmt.Sprintf("LEAF_NODE_MAX_CELLS: %d\n", LEAF_NODE_MAX_CELLS)
	return s
}

func initializeLeafNode(node *[PAGE_SIZE]byte) {
	setNodeType(node, NODE_LEAF)
	setNodeRoot(node, false)
	*leafNodeNumCells(node) = 0
	*leafNodeNextLeaf(node) = 0
}

func initializeInternalNode(node *[PAGE_SIZE]byte) {
	setNodeType(node, NODE_INTERNAL)
	setNodeRoot(node, false)
	*internalNodeNumKeys(node) = 0
}
