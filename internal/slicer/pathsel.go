package slicer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/canonical/chisel/internal/strdist"
)

type PathSelection[U any, A any] struct {
	root                   *_Node[U]
	Relative               bool
	InitUserData           func(value *PathValue[U])
	UpdateUserData         func(value *PathValue[U], arg A)
	UpdateImplicitUserData func(value *PathValue[U], arg A)
}

func ReplaceUserData[U any](value *PathValue[U], arg U) {
	value.UserData = arg
}

type PathValue[U any] struct {
	Path       string
	PathIsGlob bool
	Implicit   bool
	Parent     *PathValue[U]
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

func stripLeadingSeparator(path string) (string, error) {
	i := 0
	for i < len(path) {
		c := path[i]
		if c == '/' {
			i++
			continue
		}
		if c == '.' {
			j := i + 1
			if j < len(path) {
				c := path[j]
				j++
				if c == '.' {
					if j == len(path) || path[j] == '/' {
						return "", fmt.Errorf("double dot pahts (../) are not supported")
					}
					break
				}
				if c == '/' {
					i = j
					continue
				}
			}
		}
		break
	}
	return path[i:], nil
}

type _InsertContext[U any, A any] struct {
	arg      A
	fullPath strings.Builder
	parent   *PathValue[U]
}

func (sel *PathSelection[U, A]) initUserData(value *PathValue[U]) {
	if sel.InitUserData != nil {
		sel.InitUserData(value)
	}
}

func (sel *PathSelection[U, A]) updateUserData(value *PathValue[U], arg A) {
	if sel.UpdateUserData != nil {
		sel.UpdateUserData(value, arg)
	}
}

func (sel *PathSelection[U, A]) updateImplicitUserData(value *PathValue[U], arg A) {
	if sel.UpdateImplicitUserData != nil {
		sel.UpdateImplicitUserData(value, arg)
	}
}

func (sel *PathSelection[U, A]) addGlob(node *_Node[U], glob string, ctx *_InsertContext[U, A]) *PathValue[U] {
	ctx.fullPath.WriteString(glob)
	entry := &_GlobEntry[U]{
		glob: glob,
		value: PathValue[U]{
			Path:       ctx.fullPath.String(),
			PathIsGlob: true,
		},
	}
	if node.globs == nil {
		node.globs = make([]*_GlobEntry[U], 0, 1)
	} else {
		for _, otherEntry := range node.globs {
			if otherEntry.glob == entry.glob {
				sel.updateUserData(&otherEntry.value, ctx.arg)
				return &otherEntry.value
			}
		}
	}
	// keep in insertion order for "predicatble" matching
	node.globs = append(node.globs, entry)
	sel.initUserData(&entry.value)
	sel.updateUserData(&entry.value, ctx.arg)
	return &entry.value
}

func (sel *PathSelection[U, A]) insertPath(node *_Node[U], path string, ctx *_InsertContext[U, A]) *PathValue[U] {
	if path == "" {
		if node.value == nil {
			node.value = &PathValue[U]{
				Path:   ctx.fullPath.String(),
				Parent: ctx.parent,
			}
		} else if node.value.Implicit {
			node.value.Implicit = false
		}
		sel.updateUserData(node.value, ctx.arg)
		return node.value
	}

	if node.value != nil {
		sel.updateImplicitUserData(node.value, ctx.arg)
		ctx.parent = node.value
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
				ctx.fullPath.WriteString(head)
				value = &PathValue[U]{
					Path:   ctx.fullPath.String(),
					Parent: ctx.parent,
				}
				if tail != "" {
					value.Implicit = true
				}
				sel.initUserData(value)
				if tail != "" {
					sel.updateImplicitUserData(value, ctx.arg)
					ctx.parent = value
				} else {
					sel.updateUserData(value, ctx.arg)
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
			if tailIsGlob {
				return sel.addGlob(node, tail, ctx)
			}
			return sel.insertPath(node, tail, ctx)
		}
		return node.value
	}

	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
	ctx.fullPath.WriteString(prefix)

	// If edge.label is a prefix of path...
	if edgeSuffix == "" {
		return sel.insertPath(edge.destination, pathSuffix, ctx)
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
	return sel.insertPath(bridge, pathSuffix, ctx)
}

func (sel *PathSelection[U, _]) searchPath(node *_Node[U], path string) *PathValue[U] {
	var value *PathValue[U]

	if path == "" {
		value = node.value
	} else {
		if edge, ok := node.children[path[0]]; ok {
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

func (sel *PathSelection[U, A]) Init() {
	sel.root = makeNode[U](nil)
}

func (sel *PathSelection[U, A]) AddPath(path string, arg A) (*PathValue[U], error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	ctx := _InsertContext[U, A]{arg: arg}
	return sel.insertPath(sel.root, path, &ctx), nil
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
