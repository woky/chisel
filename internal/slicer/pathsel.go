package slicer

import (
	"fmt"
	"sort"
	"strings"
)

type PathTree[V any, A any] struct {
	Root               PathNode[V, A]
	InitNode           func(node *PathNode[V, A])
	UpdateNode         func(node *PathNode[V, A], arg A)
	UpdateImplicitNode func(node *PathNode[V, A], arg A)
}

func ReplaceValue[V any](node *PathNode[V, V], arg V) {
	node.Value = arg
}

const (
	_FL_INTERNAL = 1 << iota
	_FL_DIRECTORY
	_FL_IMPLICIT
	_FL_GLOB
)

type PathNode[V any, A any] struct {
	Tree     *PathTree[V, A]
	Parent   *PathNode[V, A]
	Path     string
	Value    V
	flags    int
	children map[byte]*_Edge[V, A]
}

type _Edge[V any, A any] struct {
	label  string
	target *PathNode[V, A]
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

func (sel *PathTree[V, A]) Init() {
	sel.Root.Tree = sel
	sel.Root.flags = _FL_DIRECTORY | _FL_IMPLICIT
	sel.Root.children = make(map[byte]*_Edge[V, A])
	if sel.InitNode != nil {
		sel.InitNode(&sel.Root)
	}
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

type _InsertContext[V any, A any] struct {
	path     string
	slash    bool
	parent   *PathNode[V, A]
	fullPath strings.Builder
	arg      A
}

func (node *PathNode[V, A]) insertInternal(ctx *_InsertContext[V, A]) (*PathNode[V, A], error) {
	var err error

	if node.flags&_FL_GLOB != 0 {
		return nil, fmt.Errorf("bug: glob node in tree")
	}

	ctx.path, err = cleanPathPrefix(ctx.path)
	if err != nil {
		return nil, err
	}

	if ctx.slash {
		if node.flags&_FL_INTERNAL != 0 {
			node.flags = _FL_DIRECTORY
		}

	}

	// if path is empty, this is insert into existing node
	if ctx.path == "" {
		if ctx.slash {
			if node.flags&_FL_INTERNAL != 0 {
				if node.flags&(_FL_DIRECTORY|_FL_IMPLICIT) != 0 {
					return nil, fmt.Errorf("bug: internal node is implicit and/or a directory")
				}
				node.flags = _FL_DIRECTORY
				ctx.fullPath.WriteByte('/') // TODO
				node.Path = ctx.fullPath.String()
				node.Parent = ctx.Parent
				node.initNode()
			} else if node.flags&_FL_DIRECTORY == 0 {
				return nil, fmt.Errorf("cannot insert a directory into a file")
			}
		} else {
		}
		if node.flags&_FL_INTERNAL != 0 {
			node.flags &^= _FL_INTERNAL
			node.Path = ctx.fullPath.String()
			node.Parent = ctx.parent // TODO test after bridge node is added
			node.initNode()
		}
		if node.flags&_FL_IMPLICIT != 0 {
			if node.flags&_FL_DIRECTORY == 0 {
				return nil, fmt.Errorf("bug: implicit node is not a directory node")
			}
			node.flags &^= _FL_IMPLICIT
		}
		node.updateNode(ctx.arg)
		return node, nil

	}

	if node.flags&_FL_DIRECTORY != 0 {
		// TODO strip leading empty paths
		if node.flags&_FL_INTERNAL != 0 {
			return nil, fmt.Errorf("bug: directory node is an internal node")
		}
		node.updateImplicitNode(ctx.arg)
		ctx.parent = node
	}

	edge := node.children[path[0]]

	// true if no edge.label shares a common prefix with path
	if edge == nil {
		var err error
		head, tail, tailIsGlob := splitPath(path)
		ctx.fullPath.WriteString(head)
		if !tailIsGlob {
			if head == "" || head == "." {
				return nil, fmt.Errorf("bug: tail is not glob but head is empty")
			}
			newNode := &PathNode[V, A]{
				Tree:     node.Tree,
				Parent:   ctx.parent,
				children: map[byte]*_Edge[V, A]{}, // TODO lazy init
			}
			node.children[path[0]] = &_Edge[V, A]{
				label:  head,
				target: newNode,
			}
			if tail == "" {
				newNode.Path = ctx.fullPath.String()
				newNode.initNode()
				newNode.updateNode(ctx.arg)
				node = newNode
			} else {
				ctx.fullPath.WriteByte('/')
				newNode.Path = ctx.fullPath.String()
				newNode.flags = _FL_DIRECTORY | _FL_IMPLICIT
				newNode.initNode()
				newNode.updateImplicitNode(ctx.arg)
				ctx.parent = newNode
				node, err = newNode.insertInternal(tail, ctx)
			}
		} else {
			if tail == "" {
				return nil, fmt.Errorf("bug: tail is empty glob")
			}
			if head != "" {
				newNode := &PathNode[V, A]{
					Tree:     node.Tree,
					flags:    _FL_INTERNAL,
					children: map[byte]*_Edge[V, A]{}, // TODO lazy init
				}
				node.children[path[0]] = &_Edge[V, A]{
					label:  head,
					target: newNode,
				}
				node = newNode
			}
			//node, err = node.AddGlob(tail, ctx) // TODO
		}
		return node, err
	}

	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
	ctx.fullPath.WriteString(prefix)

	// true if edge.label is a prefix of path
	if edgeSuffix == "" {
		return edge.target.insertInternal(pathSuffix, ctx)
	}

	// edge.label and path share a common prefix
	bridge := &PathNode[V, A]{
		Tree:     node.Tree,
		flags:    _FL_INTERNAL,
		children: map[byte]*_Edge[V, A]{}, // TODO better init
	}
	node.children[path[0]] = &_Edge[V, A]{
		label:  prefix,
		target: bridge,
	}
	bridge.children[edgeSuffix[0]] = &_Edge[V, A]{
		label:  edgeSuffix,
		target: edge.target,
	}
	return bridge.insertInternal(pathSuffix, ctx)
}

func (node *PathNode[V, A]) searchInternal(path string) *PathNode[V, A] {
	var result *PathNode[V, A]
	path, err := cleanPathPrefix(path)
	if err != nil {
		return nil
	}
	if path == "" {
		if node.flags&_FL_INTERNAL == 0 {
			result = node
		}
	} else {
		if edge, ok := node.children[path[0]]; ok {
			_, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
			if edgeSuffix == "" {
				next := node.children[path[0]].target
				result = next.searchInternal(pathSuffix)
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
	typeList := make([]string, 0, 4)
	if node.flags&_FL_INTERNAL != 0 {
		typeList = append(typeList, "INTERNAL")
	}
	if node.flags&_FL_DIRECTORY != 0 {
		typeList = append(typeList, "DIRECTORY")
	}
	if node.flags&_FL_IMPLICIT != 0 {
		typeList = append(typeList, "IMPLICIT")
	}
	if node.flags&_FL_GLOB != 0 {
		typeList = append(typeList, "GLOB")
	}
	if len(typeList) == 0 {
		typeList = append(typeList, "FILE")
	}
	fmt.Printf("% *sNODE %v %#v\n", indent, "", typeList, node.Path)

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
	if node.flags&_FL_INTERNAL != 0 {
		return nil, fmt.Errorf("bug: Insert() on internal node")
	}
	ctx := _InsertContext[V, A]{
		path:  path,
		slash: slash,
		arg:   arg,
	}
	ctx.fullPath.WriteString(node.Path)
	return node.insertInternal(&ctx)
}

func (node *PathNode[V, A]) Search(path string) *PathNode[V, A] {
	//TODO
	path, err := cleanPathPrefix(path)
	if err != nil {
		return nil
	}
	return node.searchInternal(path)
}

func (node *PathNode[V, A]) Contains(path string) bool {
	return node.Search(path) != nil
}

func (node *PathNode[V, A]) Print() {
	node.printInternal(0)
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
