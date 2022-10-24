package slicer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/canonical/chisel/internal/strdist"
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
	globs    []_Edge[V, A]
}

type _Edge[V any, A any] struct {
	label  string
	target *PathNode[V, A]
}

type _InsertContext[V any, A any] struct {
	path     string
	slash    bool
	parent   *PathNode[V, A]
	fullPath strings.Builder
	arg      A
}

type _FindContext[V any, A any] struct {
	path   string
	first  bool
	slash  bool
	result []*PathNode[V, A]
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

func (node *PathNode[V, A]) initEdges() {
	node.children = make(map[byte]*_Edge[V, A])
}

func (node *PathNode[V, A]) getEdge(c byte) *_Edge[V, A] {
	if node.children == nil {
		return nil
	}
	return node.children[c]
}

func (node *PathNode[V, A]) setEdge(c byte, label string, target *PathNode[V, A]) {
	if node.children == nil {
		node.initEdges()
	}
	node.children[c] = &_Edge[V, A]{label: label, target: target}
}

func (node *PathNode[V, A]) initGlobs() {
	node.globs = make([]_Edge[V, A], 0, 1)
}

func (node *PathNode[V, A]) addGlob(ctx *_InsertContext[V, A], glob string) *PathNode[V, A] {
	if node.globs != nil {
		for _, globEdge := range node.globs {
			if globEdge.label == glob {
				globNode := globEdge.target
				globNode.updateNode(ctx.arg)
				return globNode
			}
		}
	} else {
		node.initGlobs()
	}
	ctx.fullPath.WriteString(glob)
	globNode := &PathNode[V, A]{
		Tree:   node.Tree,
		Parent: ctx.parent,
		Path:   ctx.fullPath.String(),
		Kind:   PATHNODE_KIND_GLOB,
	}
	globEdge := _Edge[V, A]{
		label:  glob,
		target: globNode,
	}
	node.globs = append(node.globs, globEdge)
	globNode.initNode()
	globNode.updateNode(ctx.arg)
	return globNode
}

func (node *PathNode[V, A]) insertFileOnSelf(ctx *_InsertContext[V, A]) error {
	switch node.Kind {
	case PATHNODE_KIND_FILE:
	case PATHNODE_KIND_INTERNAL:
		node.Kind = PATHNODE_KIND_FILE
		node.Parent = ctx.parent
		node.Path = ctx.fullPath.String()
		node.initNode()
	case PATHNODE_KIND_DIRECTORY:
		return fmt.Errorf("trying to insert file at existing directory")
	}
	node.updateNode(ctx.arg)
	return nil
}

func (node *PathNode[V, A]) insertDirOnSelf(ctx *_InsertContext[V, A], implicit bool) error {
	switch node.Kind {
	case PATHNODE_KIND_FILE:
		return fmt.Errorf("trying to insert directory at existing file")
	case PATHNODE_KIND_INTERNAL:
		ctx.fullPath.WriteByte('/')
		node.Kind = PATHNODE_KIND_DIRECTORY
		node.Parent = ctx.parent
		node.Path = ctx.fullPath.String()
		node.initNode()
	case PATHNODE_KIND_DIRECTORY:
		node.Flags &^= PATHNODE_FLAG_IMPLICIT
	}
	if implicit {
		node.updateImplicitNode(ctx.arg)
	} else {
		node.updateNode(ctx.arg)
	}
	return nil
}

func (node *PathNode[V, A]) insertToNewEdge(ctx *_InsertContext[V, A]) (newNode *PathNode[V, A], err error) {
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
		node.setEdge(ctx.path[0], head, newNode)
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
			newNode, err = newNode.insertInTree(ctx)
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
			node.setEdge(ctx.path[0], head, newNode)
		}
		newNode = newNode.addGlob(ctx, tail)
	}
	return
}

func (node *PathNode[V, A]) insertToExistingEdge(ctx *_InsertContext[V, A], edge *_Edge[V, A]) (newNode *PathNode[V, A], err error) {
	prefix, pathSuffix, edgeSuffix := longestCommonPrefix(ctx.path, edge.label)
	ctx.path = pathSuffix
	ctx.slash = strings.HasPrefix(pathSuffix, "/")
	ctx.fullPath.WriteString(prefix)
	// true if edge.label is a prefix of path
	if edgeSuffix == "" {
		newNode, err = edge.target.insertInTree(ctx)
	} else {
		// edge.label and path share a common prefix
		bridge := &PathNode[V, A]{
			Tree: node.Tree,
			Kind: PATHNODE_KIND_INTERNAL,
		}
		node.setEdge(prefix[0], prefix, bridge)
		bridge.setEdge(edgeSuffix[0], edgeSuffix, edge.target)
		newNode, err = bridge.insertInTree(ctx)
	}
	return
}

func (node *PathNode[V, A]) insertInTree(ctx *_InsertContext[V, A]) (newNode *PathNode[V, A], err error) {
	switch node.Kind {
	case PATHNODE_KIND_FILE:
	case PATHNODE_KIND_INTERNAL:
	case PATHNODE_KIND_DIRECTORY:
	case PATHNODE_KIND_GLOB:
		err = fmt.Errorf("cannot insert on glob")
		return
	default:
		panic("bug: invalid node kind")
	}

	ctx.path, err = cleanPathPrefix(ctx.path)
	if err != nil {
		return
	}

	if ctx.path == "" {
		newNode = node
		if ctx.slash {
			err = node.insertDirOnSelf(ctx, false)
		} else {
			err = node.insertFileOnSelf(ctx)
		}
	} else {
		if ctx.slash {
			node.insertDirOnSelf(ctx, true)
		}
		edge := node.getEdge(ctx.path[0])
		if edge == nil {
			newNode, err = node.insertToNewEdge(ctx)
		} else {
			newNode, err = node.insertToExistingEdge(ctx, edge)
		}
	}
	return
}

func (node *PathNode[V, A]) findInTree(ctx *_FindContext[V, A]) (result *PathNode[V, A]) {
	path, err := cleanPathPrefix(ctx.path)
	if err != nil {
		return
	}
	if path != "" {
		if edge := node.getEdge(path[0]); edge != nil {
			_, pathSuffix, edgeSuffix := longestCommonPrefix(path, edge.label)
			if edgeSuffix == "" {
				ctx.path = pathSuffix
				ctx.slash = strings.HasPrefix(pathSuffix, "/")
				result = edge.target.findInTree(ctx)
			}
		}
	} else if (node.Kind == PATHNODE_KIND_FILE && !ctx.slash) || node.Kind == PATHNODE_KIND_DIRECTORY {
		result = node
	}
	if (result == nil || !ctx.first) && node.globs != nil {
		for _, globEdge := range node.globs {
			if strdist.GlobPath(globEdge.label, ctx.path) {
				return globEdge.target
			}
		}
	}
	return
}

func (node *PathNode[V, A]) printInternal(indent int) {
	var kindStr string
	switch node.Kind {
	case PATHNODE_KIND_FILE:
		kindStr = "FILE"
	case PATHNODE_KIND_INTERNAL:
		kindStr = "INTERNAL"
	case PATHNODE_KIND_DIRECTORY:
		kindStr = "DIR"
		if node.Flags&PATHNODE_FLAG_IMPLICIT != 0 {
			kindStr = "DIR*"
		}
	default:
		panic("bug: invalid node kind")
	}
	fmt.Printf("% *sNODE %s %#v\n", indent, "", kindStr, node.Path)
	if node.globs != nil {
		for _, globEdge := range node.globs {
			fmt.Printf("% *sGLOB %#v\n", indent+4, "", globEdge.label)
		}
	}
	if node.children != nil {
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
	ctx := _InsertContext[V, A]{
		path:   path,
		slash:  true,
		parent: node,
		arg:    arg,
	}
	ctx.fullPath.WriteString(node.Path)
	return node.insertInTree(&ctx)
}

func (node *PathNode[V, A]) FindAll(path string) *PathNode[V, A] {
	ctx := _FindContext[V, A]{
		path:   path,
		first:  false,
		slash:  true,
		result: make([]*PathNode[V, A], 0),
	}
	return node.findInTree(&ctx)
}

func (node *PathNode[V, A]) Find(path string) *PathNode[V, A] {
	ctx := _FindContext[V, A]{
		path:   path,
		first:  true,
		slash:  true,
		result: make([]*PathNode[V, A], 0, 1),
	}
	return node.findInTree(&ctx)
}

func (node *PathNode[V, A]) Contains(path string) bool {
	return node.Find(path) != nil
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

func (tree *PathTree[V, A]) Find(path string) *PathNode[V, A] {
	return tree.Root.Find(path)
}

func (tree *PathTree[V, A]) Contains(path string) bool {
	return tree.Root.Contains(path)
}
