package slicer

import (
	"fmt"
	"sort"
	"strings"
)

func ReplaceValue[V any](node *PathNode[V, V], arg V) {
	node.Value = arg
}

type PathTree[V any, A any] struct {
	Root               PathNode[V, A]
	InitNode           func(node *PathNode[V, A])
	UpdateNode         func(node *PathNode[V, A], arg A)
	UpdateImplicitNode func(node *PathNode[V, A], arg A)
}

const (
	PATHNODE_KIND_FILE = iota
	PATHNODE_KIND_INTERNAL
	PATHNODE_KIND_DIRECTORY
	PATHNODE_KIND_GLOB
)
const (
	PATHNODE_FLAG_IMPLICIT = 1 << iota
)

type PathNode[V any, A any] struct {
	Tree     *PathTree[V, A]
	Parent   *PathNode[V, A]
	Path     string
	Value    V
	Kind     int
	Flags    int
	children map[byte]*_Edge[V, A]
}

type _Edge[V any, A any] struct {
	label  string
	target *PathNode[V, A]
}

func longestCommonPrefix(a, b string) (prefix, aSuffix, bSuffix string, slash bool) {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	i := 0
	for ; i < limit; i++ {
		if a[i] == '/' {
			slash = true
		}
		if a[i] != b[i] {
			break
		}
	}
	return a[:i], a[i:], b[i:], true
}

func cleanPathPrefix(path string) (string, error) {
	const (
		S_INIT = iota
		S_DOT1
		S_DOT2
	)
	s := S_INIT
	j := 0
out:
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '/':
			switch s {
			case S_INIT:
				j++
			case S_DOT1:
				j += 2
				s = S_INIT
			case S_DOT2:
				break out

			}
		case '.':
			switch s {
			case S_INIT:
				s = S_DOT1
			case S_DOT1:
				s = S_DOT2
			case S_DOT2:
				s = S_INIT
				break out
			}
		default:
			s = S_INIT
			break out
		}
	}
	switch s {
	case S_DOT1:
		j++
	case S_DOT2:
		return "", fmt.Errorf("path backreferences (\"..\") are not supported")
	}
	return path[j:], nil
}

func splitPath(path string) (head string, tail string, tailIsGlob bool) {
	i := 0
	tailIsGlob = false
	for ; i < len(path); i++ {
		if path[i] == '/' {
			break
		}
		if path[i] == '*' || path[i] == '?' {
			tailIsGlob = true
			break
		}
	}
	head, tail = path[:i], path[i:]
	return
}

func (node *PathNode[V, _]) initNode() {
	if node.Tree.InitNode != nil {
		node.Tree.InitNode(node)
	}
}

func (node *PathNode[V, A]) updateNode(arg A) {
	if node.Tree.UpdateNode != nil {
		node.Tree.UpdateNode(node, arg)
	}
}

func (node *PathNode[V, A]) updateImplicitNode(arg A) {
	if node.Tree.UpdateImplicitNode != nil {
		node.Tree.UpdateImplicitNode(node, arg)
	}
}

func (node PathNode[V, A]) initEdges() {
	node.children = make(map[byte]*_Edge[V, A])
}

func (node PathNode[V, A]) getEdge(c byte) *_Edge[V, A] {
	if node.children == nil {
		return nil
	}
	return node.children[c]
}

func (node PathNode[V, A]) setEdge(c byte, label string, target *PathNode[V, A]) {
	if node.children == nil {
		node.initEdges()
	}
	node.children[c] = &_Edge[V, A]{label: label, target: target}
}

type _InsertContext[V any, A any] struct {
	path     string
	slash    bool
	parent   *PathNode[V, A]
	fullPath strings.Builder
	arg      A
}

func (node *PathNode[V, A]) _insert(ctx *_InsertContext[V, A]) (*PathNode[V, A], error) {
	var newNode *PathNode[V, A]
	var err error

	switch node.Kind {
	case PATHNODE_KIND_FILE:
	case PATHNODE_KIND_INTERNAL:
	case PATHNODE_KIND_DIRECTORY:
	case PATHNODE_KIND_GLOB:
		panic("bug: glob node in tree")
	default:
		panic("bug: unknown node kind")
	}

	ctx.path, err = cleanPathPrefix(ctx.path)
	if err != nil {
		return nil, err
	}

	if ctx.path != "" {
		newNode = node
		if ctx.slash {
			switch node.Kind {
			case PATHNODE_KIND_FILE:
				return nil, fmt.Errorf("trying to insert directory at existing file")
			case PATHNODE_KIND_INTERNAL:
				node.Kind = PATHNODE_KIND_DIRECTORY
				ctx.fullPath.WriteByte('/')
				node.Path = ctx.fullPath.String()
				node.Parent = ctx.parent
				node.initNode()
				node.updateNode(ctx.arg)
			case PATHNODE_KIND_DIRECTORY:
				node.Flags &^= PATHNODE_FLAG_IMPLICIT
				node.updateNode(ctx.arg)
			}
		} else {
			switch node.Kind {
			case PATHNODE_KIND_FILE:
				node.updateNode(ctx.arg)
			case PATHNODE_KIND_INTERNAL:
				node.Kind = PATHNODE_KIND_FILE
				node.Path = ctx.fullPath.String()
				node.Parent = ctx.parent
				node.initNode()
				node.updateNode(ctx.arg)
			case PATHNODE_KIND_DIRECTORY:
				return nil, fmt.Errorf("trying to insert file at existing directory")
			}
		}
	} else {
		if ctx.slash {
			switch node.Kind {
			case PATHNODE_KIND_FILE:
				return nil, fmt.Errorf("trying to insert directory at existing file")
			case PATHNODE_KIND_INTERNAL:
				node.Kind = PATHNODE_KIND_DIRECTORY
				ctx.fullPath.WriteByte('/')
				node.Path = ctx.fullPath.String()
				node.Parent = ctx.parent
				node.initNode()
				node.updateImplicitNode(ctx.arg)
			case PATHNODE_KIND_DIRECTORY:
				node.updateImplicitNode(ctx.arg)
			}
		}
		c := ctx.path[0]
		edge := node.getEdge(c)
		// true if no edge.label shares a common prefix with path
		if edge == nil {
			head, tail, tailIsGlob := splitPath(ctx.path)
			ctx.fullPath.WriteString(head)
			if !tailIsGlob {
				if head == "" || head == "." {
					panic("bug: tail is not glob but head is empty")
				}
				newNode = &PathNode[V, A]{
					Tree:   node.Tree,
					Parent: ctx.parent,
				}
				node.setEdge(c, head, newNode)
				if tail == "" {
					newNode.Kind = PATHNODE_KIND_FILE
					newNode.Path = ctx.fullPath.String()
					newNode.initNode()
					newNode.updateNode(ctx.arg)
					node = newNode
				} else {
					newNode.Kind = PATHNODE_KIND_DIRECTORY
					ctx.fullPath.WriteByte('/')
					newNode.Path = ctx.fullPath.String()
					newNode.Flags = PATHNODE_FLAG_IMPLICIT
					newNode.initNode()
					newNode.updateImplicitNode(ctx.arg)
					ctx.path = tail
					ctx.parent = newNode
					newNode, err = newNode._insert(ctx)
				}
			} else {
				if tail == "" {
					panic("bug: tail is empty glob")
				}
				if head == "" {
					newNode = node
				} else {
					newNode = &PathNode[V, A]{
						Tree: node.Tree,
						Kind: PATHNODE_KIND_INTERNAL,
					}
					node.setEdge(c, head, newNode)
				}
				//newNode, err = newNode._insertGlob(ctx, tail) // TODO
			}
		} else {
			prefix, pathSuffix, edgeSuffix, _ := longestCommonPrefix(ctx.path, edge.label)
			ctx.path = pathSuffix
			ctx.fullPath.WriteString(prefix)
			// true if edge.label is a prefix of path
			if edgeSuffix == "" {
				newNode, err = edge.target._insert(ctx)
			} else {
				// edge.label and path share a common prefix
				bridge := &PathNode[V, A]{
					Tree: node.Tree,
					Kind: PATHNODE_KIND_INTERNAL,
				}
				node.setEdge(c, prefix, bridge)
				bridge.setEdge(edgeSuffix[0], edgeSuffix, edge.target)
				newNode, err = bridge._insert(ctx)
			}
		}
	}
	return newNode, err
}

func (node *PathNode[V, A]) _search(path string, slash bool) *PathNode[V, A] {
	var result *PathNode[V, A]
	path, err := cleanPathPrefix(path)
	if err != nil {
		return nil
	}
	if path == "" {
		if (node.Kind == PATHNODE_KIND_FILE && !slash) || node.Kind == PATHNODE_KIND_DIRECTORY {
			result = node
		}
	} else {
		if edge := node.getEdge(path[0]); edge != nil {
			_, pathSuffix, edgeSuffix, slash := longestCommonPrefix(path, edge.label)
			if edgeSuffix == "" {
				result = edge.target._search(pathSuffix, slash)
			}
		}
	}
	/* TODO
	if result == nil && node.globs != nil {
		for _, globEntry := range node.globs {
			if strdist.GlobPath(globEntry.glob, path) {
				value = &globEntry.value
				break
			}
		}
	}
	*/
	return result
}

func (node *PathNode[V, A]) printInternal(indent int) {
	var kindStr string
	switch node.Kind {
	case PATHNODE_KIND_FILE:
		kindStr = "FILE"
	case PATHNODE_KIND_INTERNAL:
		kindStr = "BRIDGE"
	case PATHNODE_KIND_DIRECTORY:
		kindStr = "DIR"
		if node.Flags&PATHNODE_FLAG_IMPLICIT != 0 {
			kindStr = "DIR*"
		} else {
			kindStr = "DIR"
		}
	case PATHNODE_KIND_GLOB:
		kindStr = "GLOB"
	}
	fmt.Printf("% *sNODE %s %#v\n", indent, "", kindStr, node.Path)

	/* TODO
	for _, glob := range node.globs {
		fmt.Printf("% *sG %#v\n", indent, "", glob)
	}
	*/

	if len(node.children) > 0 {
		keys := make([]int, 0, len(node.children))
		for k := range node.children {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		for _, k := range keys {
			edge := node.children[byte(k)]
			fmt.Printf("% *sEDGE %#v\n", indent+4, "", edge.label)
			next := edge.target
			next.printInternal(indent + 8)
		}
	}
}

func (node *PathNode[V, A]) Insert(path string, arg A) (*PathNode[V, A], error) {
	if node.Kind != PATHNODE_KIND_DIRECTORY {
		return nil, fmt.Errorf("bug: Insert() on non-directory node")
	}
	ctx := _InsertContext[V, A]{
		path:  path,
		slash: true,
		arg:   arg,
	}
	ctx.fullPath.WriteString(node.Path)
	return node._insert(&ctx)
}

func (node *PathNode[V, A]) Search(path string) *PathNode[V, A] {
	return node._search(path, true)
}

func (node *PathNode[V, A]) Contains(path string) bool {
	return node.Search(path) != nil
}

func (node *PathNode[V, A]) Print() {
	node.printInternal(0)
}

func (tree *PathTree[V, A]) Init() {
	tree.Root.Tree = tree
	tree.Root.Kind = PATHNODE_KIND_DIRECTORY
	tree.Root.Flags = PATHNODE_FLAG_IMPLICIT
	if tree.InitNode != nil {
		tree.InitNode(&tree.Root)
	}
}

func (tree *PathTree[V, A]) Insert(path string, arg A) (*PathNode[V, A], error) {
	return tree.Root.Insert(path, arg)
}

func (tree *PathTree[V, A]) Search(path string) *PathNode[V, A] {
	return tree.Root.Search(path)
}

func (tree *PathTree[V, A]) Contains(path string) bool {
	return tree.Root.Contains(path)
}

//
//type PathValue[V any] struct {
//	Path       string
//	PathIsGlob bool
//	Implicit   bool
//	Parent     *PathValue[V]
//	UserData   V
//}
//
//type _GlobEntry[V any] struct {
//	glob  string
//	value PathValue[V]
//}
//
//type _Edge[V any] struct {
//	label       string
//	target *_Node[V]
//}
//
//type _Node[V any] struct {
//	value    *PathValue[V]
//	children map[byte]*_Edge[V]
//	globs    []*_GlobEntry[V]
//}
//
//
//
//
//func (sel *PathSelection[V, A]) addGlob(node *_Node[V], glob string, ctx *_InsertContext[V, A]) *PathValue[V] {
//	ctx.fullPath.WriteString(glob)
//	entry := &_GlobEntry[V]{
//		glob: glob,
//		value: PathValue[V]{
//			Path:       ctx.fullPath.String(),
//			PathIsGlob: true,
//		},
//	}
//	if node.globs == nil {
//		node.globs = make([]*_GlobEntry[V], 0, 1)
//	} else {
//		for _, otherEntry := range node.globs {
//			if otherEntry.glob == entry.glob {
//				sel.updateUserData(&otherEntry.value, ctx.arg)
//				return &otherEntry.value
//			}
//		}
//	}
//	// keep in insertion order for "predicatble" matching
//	node.globs = append(node.globs, entry)
//	sel.initUserData(&entry.value)
//	sel.updateUserData(&entry.value, ctx.arg)
//	return &entry.value
//}
//
//func (sel *PathSelection[V, A]) insertPath(node *_Node[V], path string, ctx *_InsertContext[V, A]) *PathValue[V] {
//	if path == "" {
//		if node.value == nil {
//			node.value = &PathValue[V]{
//				Path:   ctx.fullPath.String(),
//				Parent: ctx.parent,
//			}
//		} else if node.value.Implicit {
//			node.value.Implicit = false
//		}
//		sel.updateUserData(node.value, ctx.arg)
//		return node.value
//	}
//
//	if node.value != nil {
//		sel.updateImplicitUserData(node.value, ctx.arg)
//		ctx.parent = node.value
//	}
//
//	edge := node.children[path[0]]
//
//	// If no edge.label shares a common prefix with path...
//	if edge == nil {
//		head, tail, tailIsGlob := splitPath(path)
//
//		// head can be empty when path starts with a glob character
//		// ("*" or "?"). In that case, tail is non-empty.
//		if head != "" {
//			var value *PathValue[V]
//
//			if !tailIsGlob {
//				ctx.fullPath.WriteString(head)
//				value = &PathValue[V]{
//					Path:   ctx.fullPath.String(),
//					Parent: ctx.parent,
//				}
//				if tail != "" {
//					value.Implicit = true
//				}
//				sel.initUserData(value)
//				if tail != "" {
//					sel.updateImplicitUserData(value, ctx.arg)
//					ctx.parent = value
//				} else {
//					sel.updateUserData(value, ctx.arg)
//				}
//			}
//
//			newNode := makeNode(value)
//			node.children[path[0]] = &_Edge[V]{
//				label:       head,
//				target: newNode,
//			}
//			node = newNode
//		}
//
//		if tail != "" {
//			if tailIsGlob {
//				return sel.addGlob(node, tail, ctx)
//			}
//			return sel.insertPath(node, tail, ctx)
//		}
//		return node.value
//	}
//
//	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
//	ctx.fullPath.WriteString(prefix)
//
//	// If edge.label is a prefix of path...
//	if edgeSuffix == "" {
//		return sel.insertPath(edge.target, pathSuffix, ctx)
//	}
//
//	// Else, edge.label and path share a common prefix.
//	bridge := makeNode[V](nil)
//	node.children[path[0]] = &_Edge[V]{
//		label:       prefix,
//		target: bridge,
//	}
//	bridge.children[edgeSuffix[0]] = &_Edge[V]{
//		label:       edgeSuffix,
//		target: edge.target,
//	}
//	return sel.insertPath(bridge, pathSuffix, ctx)
//}
//
//func (sel *PathSelection[V, _]) searchPath(node *_Node[V], path string) *PathValue[V] {
//	var value *PathValue[V]
//
//	if path == "" {
//		value = node.value
//	} else {
//		if edge, ok := node.children[path[0]]; ok {
//			_, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
//			if edgeSuffix == "" {
//				next := node.children[path[0]].target
//				value = sel.searchPath(next, pathSuffix)
//			}
//		}
//	}
//
//	if value == nil && node.globs != nil {
//		for _, globEntry := range node.globs {
//			if strdist.GlobPath(globEntry.glob, path) {
//				value = &globEntry.value
//				break
//			}
//		}
//	}
//
//	return value
//}
//
//func (sel *PathSelection[V, _]) dumpTree(node *_Node[V], indent int) {
//	for _, glob := range node.globs {
//		fmt.Printf("% *sG %#v\n", indent, "", glob)
//	}
//	keys := make([]int, 0, len(node.children))
//	for k := range node.children {
//		keys = append(keys, int(k))
//	}
//	sort.Ints(keys)
//	for _, k := range keys {
//		edge := node.children[byte(k)]
//		next := edge.target
//		value := '0'
//		if next.value != nil {
//			value = '1'
//		}
//		fmt.Printf("% *s%c <== %#v\n", indent, "", value, edge.label)
//		sel.dumpTree(next, indent+4)
//	}
//}
