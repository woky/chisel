package slicer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/canonical/chisel/internal/strdist"
)

type _Edge struct {
	label       string
	destination *_Node
}

type _Node struct {
	value    bool
	children map[byte]*_Edge
	globs    []string
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

func splitPath(path string) (head string, tail string, tailIsGlob bool) {
	i := 0
	tailIsGlob = false
	for ; i < len(path); i++ {
		if path[i] == '/' {
			i++
			break
		}
		if path[i] == '*' || path[i] == '?' {
			tailIsGlob = true
			break
		}
	}
	head = path[:i]
	tail = path[i:]
	return
}

func addGlob(node *_Node, glob string) {
	if node.globs == nil {
		node.globs = make([]string, 1)
		node.globs[0] = glob
	} else {
		// keep in insertion order for "predicatble" matching
		for _, other := range node.globs {
			if other == glob {
				return
			}
		}
		node.globs = append(node.globs, glob)
	}
}

type _InsertContext struct {
	parentPath strings.Builder
}

func insertPath(node *_Node, path string, ctx *_InsertContext) {
	if path == "" {
		node.value = true
		return
	}
	edge := node.children[path[0]]
	// If no edge.label shares a common prefix with path...
	if edge == nil {
		head, tail, tailIsGlob := splitPath(path)
		// head can be empty when path starts with a glob character
		// ("*" or "?"). In that case, tail is non-empty.
		if head != "" {
			newNode := makeNode(!tailIsGlob)
			node.children[path[0]] = &_Edge{
				label:       head,
				destination: newNode,
			}
			node = newNode
		}
		if tail != "" {
			if !tailIsGlob {
				insertPath(node, tail, ctx)
			} else {
				addGlob(node, tail)
			}
		}
		return
	}
	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
	// If edge.label is a prefix of path...
	if edgeSuffix == "" {
		insertPath(edge.destination, pathSuffix, ctx)
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
	insertPath(bridge, pathSuffix, ctx)
}

func searchPath(node *_Node, path string) bool {
	value := false
	if path == "" {
		value = node.value
	} else {
		edge := node.children[path[0]]
		if edge != nil {
			_, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
			if edgeSuffix == "" {
				next := node.children[path[0]].destination
				value = searchPath(next, pathSuffix)
			}
		}
	}
	if !value && node.globs != nil {
		for _, glob := range node.globs {
			if strdist.GlobPath(glob, path) {
				value = true
				break
			}
		}
	}
	return value
}

func dumpTree(node *_Node, indent int) {
	for _, glob := range node.globs {
		fmt.Printf("% *sG %#v\n", indent, "", glob)
	}
	keys := make([]int, 0, len(node.children))
	for k := range node.children {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		edge := node.children[byte(k)]
		next := edge.destination
		value := '0'
		if next.value {
			value = '1'
		}
		fmt.Printf("% *s%c <== %#v\n", indent, "", value, edge.label)
		dumpTree(next, indent+4)
	}
}

type PathSelection struct {
	root *_Node
}

func CreatePathSelection() PathSelection {
	return PathSelection{root: makeNode(false)}
}

func (sel *PathSelection) AddPath(path string) {
	var ctx _InsertContext
	insertPath(sel.root, path, &ctx)
}

func (sel *PathSelection) IsPathSelected(path string) bool {
	return searchPath(sel.root, path)
}

func (sel *PathSelection) DumpTree() {
	dumpTree(sel.root, 0)
}
