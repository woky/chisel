package slicer

import (
	"fmt"
	"sort"

	"github.com/canonical/chisel/internal/strdist"
)

type NodePathInfo struct {
	Path       string
	PathIsGlob bool
	Implicit   bool
}

type _GlobEntry struct {
	glob  string
	value NodePathInfo
}

type _Edge struct {
	label       string
	destination *_Node
}

type _Node struct {
	value    *NodePathInfo
	children map[byte]*_Edge
	globs    []*_GlobEntry
}

func makeNode(value *NodePathInfo) *_Node {
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

type _InsertContext struct {
	origPath       string
	origPathOffset int
}

func addGlob(node *_Node, glob string, ctx *_InsertContext) {
	entry := &_GlobEntry{
		glob: glob,
		value: NodePathInfo{
			Path:       ctx.origPath,
			PathIsGlob: true,
		},
	}
	if node.globs == nil {
		node.globs = make([]*_GlobEntry, 1)
		node.globs[0] = entry
	} else {
		// keep in insertion order for "predicatble" matching
		for _, otherEntry := range node.globs {
			if entry.glob == otherEntry.glob {
				return
			}
		}
		node.globs = append(node.globs, entry)
	}
}

func insertPath(node *_Node, path string, ctx *_InsertContext) {
	if path == "" {
		node.value = &NodePathInfo{
			Path: ctx.origPath[0:ctx.origPathOffset],
		}
		return
	}
	edge := node.children[path[0]]
	// If no edge.label shares a common prefix with path...
	if edge == nil {
		head, tail, tailIsGlob := splitPath(path)
		// head can be empty when path starts with a glob character
		// ("*" or "?"). In that case, tail is non-empty.
		if head != "" {
			var value *NodePathInfo
			if !tailIsGlob {
				ctx.origPathOffset += len(head)
				value = &NodePathInfo{
					Path:     ctx.origPath[0:ctx.origPathOffset],
					Implicit: tail != "",
				}
			}
			newNode := makeNode(value)
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
				addGlob(node, tail, ctx)
			}
		}
		return
	}
	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
	ctx.origPathOffset += len(prefix)
	// If edge.label is a prefix of path...
	if edgeSuffix == "" {
		insertPath(edge.destination, pathSuffix, ctx)
		return
	}
	// Else, edge.label and path share a common prefix.
	bridge := makeNode(nil)
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

func searchPath(node *_Node, path string) *NodePathInfo {
	var value *NodePathInfo
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
	if value == nil && node.globs != nil {
		for _, globEntry := range node.globs {
			if strdist.GlobPath(globEntry.glob, path) {
				value = &globEntry.value
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
		if next.value != nil {
			value = '1'
		}
		fmt.Printf("% *s%c <== %#v\n", indent, "", value, edge.label)
		dumpTree(next, indent+4)
	}
}

type PathSelection[UserData any, UserDataArg any] struct {
	root *_Node
}

type PathValue[UserData any] struct {
	info *NodePathInfo
	data UserData
}

func CreatePathSelection[D any, A any]() PathSelection[D, A] {
	return PathSelection[D, A]{root: makeNode(nil)}
}

func (sel *PathSelection[D, A]) AddPath(path string) {
	ctx := _InsertContext{origPath: path}
	insertPath(sel.root, path, &ctx)
}

func (sel *PathSelection[D, A]) FindPath(path string) *NodePathInfo {
	return searchPath(sel.root, path)
}

func (sel *PathSelection[D, A]) ContainsPath(path string) bool {
	return sel.FindPath(path) != nil
}

func (sel *PathSelection[D, A]) DumpTree() {
	dumpTree(sel.root, 0)
}
