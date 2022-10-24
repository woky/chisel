package main

import (
	"fmt"

	"github.com/canonical/chisel/internal/slicer"
)

func main() {
	var node *slicer.PathNode[bool, any]
	var err error
	tree := slicer.PathTree[bool, any]{}
	tree.Init()
	node, err = tree.Insert("/foo*", nil)
	fmt.Println(tree.Contains("/foo/"))

	_ = node
	_ = err
}
