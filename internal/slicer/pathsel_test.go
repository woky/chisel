package slicer_test

import (
	. "gopkg.in/check.v1"

	"github.com/canonical/chisel/internal/slicer"
)

func (s *S) TestLongestCommonPrefix(c *C) {
	runTest := func(a, b, expPrefix, expASuffix, expBSuffix string) {
		prefix, aSuffix, bSuffix := slicer.LongestCommonPrefix(a, b)
		c.Assert(prefix, Equals, expPrefix)
		c.Assert(aSuffix, Equals, expASuffix)
		c.Assert(bSuffix, Equals, expBSuffix)
	}

	runTest("abcd", "abab", "ab", "cd", "ab")
	runTest("abcd", "ab", "ab", "cd", "")
	runTest("ab", "ab", "ab", "", "")
	runTest("ab", "ab", "ab", "", "")
	runTest("ab", "cd", "", "ab", "cd")
	runTest("", "", "", "", "")
	runTest("", "", "", "", "")
	runTest("", "ab", "", "", "ab")
	runTest("ab/c", "ab/c", "ab/c", "", "")
	runTest("/a", "/a", "/a", "", "")
	runTest("./aa", "./ab", "./a", "a", "b")
	runTest("aa", "aa/", "aa", "", "/")
}

func (s *S) TestCleanPathPrefix(c *C) {
	runTest := func(prefix, expResult string, errCheck Checker) {
		result, err := slicer.CleanPathPrefix(prefix)
		c.Assert(result, Equals, expResult)
		c.Assert(err, errCheck)
	}

	runTest("", "", IsNil)
	runTest("/", "", IsNil)
	runTest("abc", "abc", IsNil)
	runTest("a/b/c", "a/b/c", IsNil)
	runTest("/a/b/c", "a/b/c", IsNil)
	runTest("./a/b/c", "a/b/c", IsNil)
	runTest("/./a/b/c", "a/b/c", IsNil)
	runTest("//a/b/c", "a/b/c", IsNil)
	runTest("/////a/b/c", "a/b/c", IsNil)
	runTest("./././a/b/c", "a/b/c", IsNil)
	runTest("/./././a/b/c", "a/b/c", IsNil)
	runTest(".//.///././/a/b/c", "a/b/c", IsNil)
	runTest("./a/./b/c", "a/./b/c", IsNil)
	runTest("a///b/./c", "a///b/./c", IsNil)
	runTest("/../a/b/c", "", NotNil)
	runTest("././//./.././a/b/c", "", NotNil)
	runTest("////../a/b/c", "", NotNil)
	runTest("..", "", NotNil)
	runTest("../", "", NotNil)
	runTest("...", "...", IsNil)
	runTest("./...//", "...//", IsNil)
	runTest("./.", "", IsNil)
	runTest(".", "", IsNil)
	runTest(".///./", "", IsNil)
	runTest(".foo", ".foo", IsNil)
	runTest("..foo", "..foo", IsNil)
	runTest(".////foo", "foo", IsNil)
}

func (s *S) TestPathTreeContainsOnePath(c *C) {
	tree := slicer.PathTree[bool, any]{}
	tree.Init()

	assertContains := func(path string, expContains bool) {
		c.Assert(tree.Contains(path), Equals, expContains)
	}

	tree.Insert("/var/log/messages", nil)

	assertContains("", true)
	assertContains("/", true)
	assertContains("/var", true)
	assertContains("/var/", true)
	assertContains("/var/log", true)
	assertContains("/var/log/", true)
	assertContains("/var/log/messages", true)
	assertContains("/var/log/messages/", false)
	assertContains("var", true)
	assertContains("var/", true)
	assertContains("var/log", true)
	assertContains("var/log/", true)
	assertContains("var/log/messages", true)
	assertContains("var/log/messages/", false)
	assertContains("./var/", true)
	assertContains("./var/.", true)
	assertContains("./var/./", true)
	assertContains("./var/./log/", true)
	assertContains("./", true)
	assertContains(".//.///", true)
	assertContains("/var/./", true)
	assertContains("//var", true)
	assertContains("/var//", true)
	assertContains("./var/../log/", false)
	assertContains("./var/.../log/", false)
	assertContains("/var/log/dmesg", false)
	assertContains("/zzz", false)
}

func (s *S) TestPathTreeContainsMorePaths(c *C) {
	tree := slicer.PathTree[bool, any]{}
	tree.Init()

	assertContains := func(path string, expContains bool) {
		c.Assert(tree.Contains(path), Equals, expContains)
	}

	tree.Insert("/a/b/c1/d/", nil)
	tree.Insert("/a/b/c1/d/e", nil)
	tree.Insert("/a/bbb/c/d/", nil)
	tree.Insert("/a/b/c1/d/eee", nil)
	tree.Insert("/a/b/c2/d/e", nil)

	assertContains("/", true)
	assertContains("/a", true)
	assertContains("/a/", true)
	assertContains("/a/b/", true)
	assertContains("/a/b/c", false)
	assertContains("/a/b/c/", false)
	assertContains("/a/b/c1/", true)
	assertContains("/a/b/c1/d", true)
	assertContains("/a/b/c1/d/", true)
	assertContains("/a/b/c1/d/e", true)
	assertContains("/a/b/c1/d/e/", false)
	assertContains("/a/b/c1/d/ee", false)
	assertContains("/a/b/c1/d/eee", true)
	assertContains("/a/b/c2", true)
	assertContains("/a/b/c2/", true)
	assertContains("/a/b/c2/d/", true)
	assertContains("/a/b/c2/d/e", true)
	assertContains("/a/bb/", false)
	assertContains("/a/bbb", true)
	assertContains("/a/bbb/", true)
	assertContains("/a/bbb/c/", true)
	assertContains("/a/bbb/c/d/", true)
	assertContains("/a/bbbb/", false)

}

func (s *S) TestPathTreeReplaceValue(c *C) {
	tree := slicer.PathTree[string, string]{}
	tree.Root.Path = "/"
	tree.UpdateNode = slicer.ReplaceValue[string]
	tree.UpdateImplicitNode = slicer.ReplaceValue[string]
	tree.Init()

	checkValue := func(insertValue, expValue string) {
		path := "/a/b/c"
		node, err := tree.Insert(path, insertValue)
		c.Assert(err, IsNil)
		c.Assert(node, NotNil)
		c.Assert(node.Path, Equals, path)
		c.Assert(node.Value, Equals, expValue)
	}

	checkValue("A", "A")
	checkValue("B", "B")
	// modyfing hooks after Init() is unsupported, but test it anyway
	tree.UpdateNode = nil
	checkValue("C", "B")
}

func (s *S) TestPathTreeContainsGlobs(c *C) {
	tree := slicer.PathTree[bool, any]{}
	tree.Init()

	assertContains := func(path string, expContains bool) {
		c.Assert(tree.Contains(path), Equals, expContains)
	}

	// TODO test **
	tree.Insert("/foo*", nil)

	assertContains("/", true)
	assertContains("/fo", false)
	assertContains("/foo", true)
	assertContains("/fooo", true)
	assertContains("/foo/", false)
	assertContains("/fooo/", false)

	tree.Insert("/fo*", nil)

	assertContains("/fo", true)
	assertContains("/foo", true)
	assertContains("/fooo", true)
	assertContains("/fo/", false)
	assertContains("/foo/", false)
	assertContains("/fooo/", false)

	tree.Insert("/foo", nil)

	assertContains("/fo", true)
	assertContains("/foo", true)
	assertContains("/fooo", true)
	assertContains("/fo/", false)
	assertContains("/foo/", false)
	assertContains("/fooo/", false)

	tree.Insert("/fo/bar", nil)

	assertContains("/fo", true)
	assertContains("/foo", true)
	assertContains("/fooo", true)
	assertContains("/fo/", true)
	assertContains("/foo/", false)
	assertContains("/fooo/", false)
}

//func (s *S) TestPathSelectionSearch(c *C) {
//	var value *slicer.PathValue[bool]
//	tree := slicer.PathSelection[bool, any]{}
//	tree.Init()
//
//	tree.Insert("/a/b/c", nil)
//
//	value = tree.Search("/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/")
//	c.Assert(value.Implicit, Equals, true)
//
//	value = tree.Search("/a/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/")
//	c.Assert(value.Implicit, Equals, true)
//
//	value = tree.Search("/a/b/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/")
//	c.Assert(value.Implicit, Equals, true)
//
//	value = tree.Search("/a/b/c")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/c")
//	c.Assert(value.Implicit, Equals, false)
//
//	tree.Insert("/a/b/cc/", nil)
//	tree.Insert("/a/bb/c", nil)
//
//	value = tree.Search("/a/b/cc/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/cc/")
//	c.Assert(value.Implicit, Equals, false)
//
//	value = tree.Search("/a/bb/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/bb/")
//	c.Assert(value.Implicit, Equals, true)
//
//	value = tree.Search("/a/bb/c")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/bb/c")
//	c.Assert(value.Implicit, Equals, false)
//
//	value = tree.Search("/a/bb/c")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/bb/c")
//	c.Assert(value.Implicit, Equals, false)
//
//	c.Assert(tree.Search("/a"), IsNil)
//	c.Assert(tree.Search("/a/b"), IsNil)
//	c.Assert(tree.Search("/a/bb"), IsNil)
//	c.Assert(tree.Search("/a/b/c/"), IsNil)
//	c.Assert(tree.Search("/zzz"), IsNil)
//
//	tree.Insert("/a/b*", nil)
//	tree.Insert("/a/bbb", nil)
//
//	value = tree.Search("/a/bbb")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/bbb")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.PathIsGlob, Equals, false)
//
//	value = tree.Search("/a/b/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/")
//	c.Assert(value.Implicit, Equals, true)
//	c.Assert(value.PathIsGlob, Equals, false)
//
//	value = tree.Search("/a/bb/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/bb/")
//	c.Assert(value.Implicit, Equals, true)
//	c.Assert(value.PathIsGlob, Equals, false)
//
//	value = tree.Search("/a/bbbb")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b*")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.PathIsGlob, Equals, true)
//
//	c.Assert(tree.Search("/a/bbbb/"), IsNil)
//
//	tree.Insert("/a**/b", nil)
//
//	value = tree.Search("/a/b/")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/")
//	c.Assert(value.Implicit, Equals, true)
//	c.Assert(value.PathIsGlob, Equals, false)
//
//	value = tree.Search("/aa/b")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a**/b")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.PathIsGlob, Equals, true)
//
//	value = tree.Search("/aa/b/c/b")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a**/b")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.PathIsGlob, Equals, true)
//
//	c.Assert(tree.Search("/aa/b/"), IsNil)
//	c.Assert(tree.Search("/aa/b/c"), IsNil)
//}
//
//func (s *S) TestPathSelectionParent(c *C) {
//	var value *slicer.PathValue[string]
//	tree := slicer.PathSelection[string, string]{}
//	tree.UpdateUserData = slicer.ReplaceUserData[string]
//	tree.UpdateImplicitUserData = tree.UpdateUserData
//	tree.Init()
//
//	tree.Insert("/a/b", "A")
//
//	value, _ = tree.Insert("/x/y/z", "X")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/x/y/z")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.UserData, Equals, "X")
//
//	value = value.Parent
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/x/y/")
//	c.Assert(value.Implicit, Equals, true)
//	c.Assert(value.UserData, Equals, "X")
//
//	value = value.Parent
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/x/")
//	c.Assert(value.Implicit, Equals, true)
//	c.Assert(value.UserData, Equals, "X")
//
//	value = value.Parent
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/")
//	c.Assert(value.Implicit, Equals, true)
//	c.Assert(value.UserData, Equals, "X")
//
//	value = value.Parent
//	c.Assert(value, IsNil)
//
//	value, _ = tree.Insert("/x/y/", "Z")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/x/y/")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.UserData, Equals, "Z")
//}
//
//func (s *S) TestPathSelectionUserData(c *C) {
//	type PathData struct {
//		initCount           int
//		updateCount         int
//		implicitUpdateCount int
//	}
//	var value *slicer.PathValue[PathData]
//	tree := slicer.PathSelection[PathData, any]{}
//	tree.InitUserData = func(value *slicer.PathValue[PathData]) {
//		value.UserData.initCount += 1
//	}
//	tree.UpdateUserData = func(value *slicer.PathValue[PathData], _ any) {
//		value.UserData.updateCount += 1
//	}
//	tree.UpdateImplicitUserData = func(value *slicer.PathValue[PathData], _ any) {
//		value.UserData.implicitUpdateCount += 1
//	}
//	tree.Init()
//
//	tree.Insert("/a/b/c", 10)
//	tree.Insert("/a/b/c/d", 1)
//	tree.Insert("/a/b/cc/d", 1)
//	tree.Insert("/a/b/", 1)
//	tree.Insert("/", 1)
//	tree.Insert("/a/b/", 1)
//
//	value = tree.Search("/")
//	c.Assert(value, NotNil)
//	//c.Assert(value.UserData.initCount, Equals, 0)
//	//c.Assert(value.UserData.updateCount, Equals, 1)
//	//c.Assert(value.UserData.implicitUpdateCount, Equals, 5)
//}
//
//func (s *S) TestPathSelectionOddities(c *C) {
//	tree := slicer.PathSelection[any, any]{}
//	tree.Init()
//
//	tree.Insert("/foo/bar", nil)
//	tree.Insert("/foo/bar/", nil)
//	//tree.DumpTree()
//	//c.Assert(true, Equals, false)
//
//	//tree.Insert("", nil)
//}
//
//// TODO oddities
//// Empty path

// TODO test insert on directory

// TODO
//tree.Insert("/var/logg/.../././//./messages", nil)
//tree.Insert("/var/log/./././//./messages", nil)
//tree.Root.Print()
