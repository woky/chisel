package main

import (
	"fmt"

	"github.com/canonical/chisel/internal/slicer"
)

func main() {
	tree := slicer.PathTree[bool, any]{}
	tree.Init()
	_, err := tree.Insert("/a/b", nil)
	fmt.Println(err)
	tree.Insert("/a/bc", nil)
	tree.Insert("/a/bc/", nil)
	tree.Root.Print()
}
