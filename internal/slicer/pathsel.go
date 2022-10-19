package slicer

import (
	"fmt"
	"sort"
	"strings"
)

type PathValue struct {
	Path  string
	Value *interface{}
}

type _Edge struct {
	label       string
	destination *_Node
}

type _Node struct {
	value    bool
	children map[byte]*_Edge
}

func makeNode(value bool) *_Node {
	return &_Node{
		value:    value,
		children: make(map[byte]*_Edge),
	}
}

func longestCommonPrefix(a, b string) (prefix, aSuffix, bSuffix string) {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	i := 0
	for ; i < limit; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return a[:i], a[i:], b[i:]
}

func splitPath(path string) (head string, tail string) {
	i := strings.Index(path, "/")
	if i < 0 {
		return path, ""
	}
	return path[:i+1], path[i+1:]
}

func insertPath(node *_Node, path string) {
	if path == "" {
		node.value = true
		return
	}

	edge := node.children[path[0]]

	// no edge.label shares a common prefix with path?
	if edge == nil {
		next := makeNode(true)
		head, tail := splitPath(path)
		node.children[path[0]] = &_Edge{
			label:       head,
			destination: next,
		}
		if tail != "" {
			insertPath(next, tail)
		}
		return
	}

	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)

	// edge.label is a prefix of path?
	if edgeSuffix == "" {
		insertPath(edge.destination, pathSuffix)
		return
	}

	// edge.label and path share a common prefix
	bridge := makeNode(false)
	node.children[path[0]] = &_Edge{
		label:       prefix,
		destination: bridge,
	}
	bridge.children[edgeSuffix[0]] = &_Edge{
		label:       edgeSuffix,
		destination: edge.destination,
	}
	insertPath(bridge, pathSuffix)
}

func searchPath(node *_Node, path string) bool {
	if path == "" {
		return node.value
	}
	edge := node.children[path[0]]
	if edge == nil {
		return false
	}
	_, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
	if edgeSuffix == "" {
		return searchPath(node.children[path[0]].destination, pathSuffix)
	}
	return false
}

type PathSelection struct {
	root *_Node
}

func CreatePathSelection() PathSelection {
	return PathSelection{root: makeNode(false)}
}

func (sel *PathSelection) AddPath(path string) {
	insertPath(sel.root, path)
}

func (sel *PathSelection) IsPathSelected(path string) bool {
	return searchPath(sel.root, path)
}

func dumpTree(node *_Node, indent int) {
	keys := make([]int, 0, len(node.children))
	for k := range node.children {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		edge := node.children[byte(k)]
		next := edge.destination
		fmt.Print(strings.Repeat(" ", indent))
		if next.value {
			fmt.Print("1 ")
		} else {
			fmt.Print("0 ")
		}
		fmt.Println(edge.label)
		dumpTree(next, indent+4)
	}
}

func (sel *PathSelection) DumpTree() {
	dumpTree(sel.root, 0)
}
