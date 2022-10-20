package slicer

import (
	"fmt"
	"sort"

	"github.com/canonical/chisel/internal/strdist"
)

type PathSelection[U any, A any] struct {
	root *_Node[U]
}

type PathValue[U any] struct {
	Path       string
	PathIsGlob bool
	Implicit   bool
	UserData   U
}

type _GlobEntry[U any] struct {
	glob  string
	value PathValue[U]
}

type _Edge[U any] struct {
	label       string
	destination *_Node[U]
}

type _Node[U any] struct {
	value    *PathValue[U]
	children map[byte]*_Edge[U]
	globs    []*_GlobEntry[U]
}

func makeNode[U any](value *PathValue[U]) *_Node[U] {
	return &_Node[U]{
		value:    value,
		children: make(map[byte]*_Edge[U]),
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

func (sel *PathSelection[U, _]) addGlob(node *_Node[U], glob string, ctx *_InsertContext) {
	entry := &_GlobEntry[U]{
		glob: glob,
		value: PathValue[U]{
			Path:       ctx.origPath,
			PathIsGlob: true,
		},
	}
	if node.globs == nil {
		node.globs = make([]*_GlobEntry[U], 1)
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

func (sel *PathSelection[U, _]) insertPath(node *_Node[U], path string, ctx *_InsertContext) {
	if path == "" {
		node.value = &PathValue[U]{
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
			var value *PathValue[U]
			if !tailIsGlob {
				ctx.origPathOffset += len(head)
				value = &PathValue[U]{
					Path:     ctx.origPath[0:ctx.origPathOffset],
					Implicit: tail != "",
				}
			}
			newNode := makeNode(value)
			node.children[path[0]] = &_Edge[U]{
				label:       head,
				destination: newNode,
			}
			node = newNode
		}
		if tail != "" {
			if !tailIsGlob {
				sel.insertPath(node, tail, ctx)
			} else {
				sel.addGlob(node, tail, ctx)
			}
		}
		return
	}
	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
	ctx.origPathOffset += len(prefix)
	// If edge.label is a prefix of path...
	if edgeSuffix == "" {
		sel.insertPath(edge.destination, pathSuffix, ctx)
		return
	}
	// Else, edge.label and path share a common prefix.
	bridge := makeNode[U](nil)
	node.children[path[0]] = &_Edge[U]{
		label:       prefix,
		destination: bridge,
	}
	bridge.children[edgeSuffix[0]] = &_Edge[U]{
		label:       edgeSuffix,
		destination: edge.destination,
	}
	sel.insertPath(bridge, pathSuffix, ctx)
}

func (sel *PathSelection[U, _]) searchPath(node *_Node[U], path string) *PathValue[U] {
	var value *PathValue[U]
	if path == "" {
		value = node.value
	} else {
		edge := node.children[path[0]]
		if edge != nil {
			_, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
			if edgeSuffix == "" {
				next := node.children[path[0]].destination
				value = sel.searchPath(next, pathSuffix)
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

func (sel *PathSelection[U, _]) dumpTree(node *_Node[U], indent int) {
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
		sel.dumpTree(next, indent+4)
	}
}

func CreatePathSelection[U any, A any]() PathSelection[U, A] {
	return PathSelection[U, A]{root: makeNode[U](nil)}
}

func (sel *PathSelection[U, A]) AddPath(path string) {
	ctx := _InsertContext{origPath: path}
	sel.insertPath(sel.root, path, &ctx)
}

func (sel *PathSelection[U, _]) FindPath(path string) *PathValue[U] {
	return sel.searchPath(sel.root, path)
}

func (sel *PathSelection[_, _]) ContainsPath(path string) bool {
	return sel.FindPath(path) != nil
}

func (sel *PathSelection[U, A]) DumpTree() {
	sel.dumpTree(sel.root, 0)
}
