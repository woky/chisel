package slicer_test

import (
	. "gopkg.in/check.v1"

	"github.com/canonical/chisel/internal/slicer"
)

func (s *S) TestLongestCommonPrefix(c *C) {
	var prefix, aSuffix, bSuffix string

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("abcd", "abab")
	c.Assert(prefix, Equals, "ab")
	c.Assert(aSuffix, Equals, "cd")
	c.Assert(bSuffix, Equals, "ab")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("abcd", "ab")
	c.Assert(prefix, Equals, "ab")
	c.Assert(aSuffix, Equals, "cd")
	c.Assert(bSuffix, Equals, "")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("ab", "ab")
	c.Assert(prefix, Equals, "ab")
	c.Assert(aSuffix, Equals, "")
	c.Assert(bSuffix, Equals, "")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("ab", "cd")
	c.Assert(prefix, Equals, "")
	c.Assert(aSuffix, Equals, "ab")
	c.Assert(bSuffix, Equals, "cd")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("", "")
	c.Assert(prefix, Equals, "")
	c.Assert(aSuffix, Equals, "")
	c.Assert(bSuffix, Equals, "")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("", "ab")
	c.Assert(prefix, Equals, "")
	c.Assert(aSuffix, Equals, "")
	c.Assert(bSuffix, Equals, "ab")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("ab/c", "ab/c")
	c.Assert(prefix, Equals, "ab/c")
	c.Assert(aSuffix, Equals, "")
	c.Assert(bSuffix, Equals, "")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("ab/c", "ab/")
	c.Assert(prefix, Equals, "ab/")
	c.Assert(aSuffix, Equals, "c")
	c.Assert(bSuffix, Equals, "")

	prefix, aSuffix, bSuffix = slicer.LongestCommonPrefix("/a/b/", "/a/b/")
	c.Assert(prefix, Equals, "/a/b/")
	c.Assert(aSuffix, Equals, "")
	c.Assert(bSuffix, Equals, "")
}

func (s *S) TestCleanPathPrefix(c *C) {
	var result string
	var err error

	result, err = slicer.CleanPathPrefix("")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "")

	result, err = slicer.CleanPathPrefix("/")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "/")

	result, err = slicer.CleanPathPrefix("abc")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "abc")

	result, err = slicer.CleanPathPrefix("a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("/a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "/a/b/c")

	result, err = slicer.CleanPathPrefix("./a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("/./a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("//a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("/////a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("./././a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("/./././a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix(".//.///././/a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.CleanPathPrefix("./a/./b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/./b/c")

	result, err = slicer.CleanPathPrefix("a///b/./c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a///b/./c")

	result, err = slicer.CleanPathPrefix("../a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.CleanPathPrefix("/../a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.CleanPathPrefix("././//./.././a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.CleanPathPrefix("////../a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.CleanPathPrefix("..")
	c.Assert(err, NotNil)

	result, err = slicer.CleanPathPrefix("../")
	c.Assert(err, NotNil)

	result, err = slicer.CleanPathPrefix("...")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "...")

	result, err = slicer.CleanPathPrefix("./...//")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "...//")

	result, err = slicer.CleanPathPrefix("./.")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "")

	result, err = slicer.CleanPathPrefix(".")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "")

	result, err = slicer.CleanPathPrefix("")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "")

	result, err = slicer.CleanPathPrefix(".///./")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "")

	result, err = slicer.CleanPathPrefix(".foo")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, ".foo")

	result, err = slicer.CleanPathPrefix("..foo")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "..foo")

	result, err = slicer.CleanPathPrefix(".////foo")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "")
}

//TODO
//c.Assert(tree.Contains(""), Equals, true)

func (s *S) TestPathTreeSinglePath(c *C) {
	tree := slicer.PathTree[bool, any]{}
	tree.Init()

	//tree.Insert("/var/log/messages", nil)
	//_, err := tree.Insert("/var/log/messages", nil)
	//c.Assert(err, IsNil)
	//_, err = tree.Insert("/var/loag/messages", nil)
	//c.Assert(err, IsNil)
	//tree.Insert("/var/logg/.../././//./messages", nil)
	//tree.Insert("/var/log/./././//./messages", nil)
	tree.Insert("/a/b", nil)
	tree.Insert("/a/bc", nil)
	tree.Insert("/a/bc/", nil)
	tree.Root.Print()

	c.Assert(tree.Contains(""), Equals, true)
	c.Assert(tree.Contains("/"), Equals, true)
	c.Assert(tree.Contains("/var"), Equals, true)
	c.Assert(tree.Contains("/var/"), Equals, true)
	c.Assert(tree.Contains("/var/log"), Equals, true)
	c.Assert(tree.Contains("/var/log/"), Equals, true)
	c.Assert(tree.Contains("/var/log/messages"), Equals, true)
	c.Assert(tree.Contains("/var/log/messages/"), Equals, true)
	c.Assert(tree.Contains("var"), Equals, false)
	c.Assert(tree.Contains("var/"), Equals, true)
	c.Assert(tree.Contains("var/log"), Equals, false)
	c.Assert(tree.Contains("var/log/"), Equals, true)
	c.Assert(tree.Contains("var/log/messages"), Equals, true)
	c.Assert(tree.Contains("var/log/messages/"), Equals, true)

	c.Assert(tree.Contains("./var/"), Equals, true)
	c.Assert(tree.Contains("./var/."), Equals, true)
	c.Assert(tree.Contains("./var/./"), Equals, true)
	c.Assert(tree.Contains("./var/./log/"), Equals, true)

	c.Assert(tree.Contains("./"), Equals, true)
	c.Assert(tree.Contains(".//.///"), Equals, true)
	c.Assert(tree.Contains("/var/./"), Equals, true)
	c.Assert(tree.Contains("//var"), Equals, true)
	c.Assert(tree.Contains("/var//"), Equals, true)

	c.Assert(tree.Contains("./var/../log/"), Equals, false)
	c.Assert(tree.Contains("./var/.../log/"), Equals, false)
	c.Assert(tree.Contains("/var/log/dmesg"), Equals, false)
	c.Assert(tree.Contains("/zzz"), Equals, false)
}

//func (s *S) TestPathSelectionFewPaths(c *C) {
//	tree := slicer.PathSelection[bool, any]{}
//	tree.Init()
//
//	tree.Insert("/a/b/c1/d/", nil)
//	tree.Insert("/a/b/c1/d/e", nil)
//	tree.Insert("/a/bbb/c/d/", nil)
//	tree.Insert("/a/b/c1/d/eee", nil)
//	tree.Insert("/a/b/c2/d/e", nil)
//
//	c.Assert(tree.Contains("/"), Equals, true)
//	c.Assert(tree.Contains("/a/"), Equals, true)
//	c.Assert(tree.Contains("/a/b/"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c1/"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c1/d/"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c1/d/e"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c1/d/eee"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c2/"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c2/d/"), Equals, true)
//	c.Assert(tree.Contains("/a/b/c2/d/e"), Equals, true)
//	c.Assert(tree.Contains("/a/bbb/"), Equals, true)
//	c.Assert(tree.Contains("/a/bbb/c/"), Equals, true)
//	c.Assert(tree.Contains("/a/bbb/c/d/"), Equals, true)
//
//	c.Assert(tree.Contains("/a"), Equals, false)
//	c.Assert(tree.Contains("/a/b/c"), Equals, false)
//	c.Assert(tree.Contains("/a/b/c/"), Equals, false)
//	c.Assert(tree.Contains("/a/b/c1/d"), Equals, false)
//	c.Assert(tree.Contains("/a/b/c1/d/e/"), Equals, false)
//	c.Assert(tree.Contains("/a/b/c1/d/ee"), Equals, false)
//	c.Assert(tree.Contains("/a/b/c2"), Equals, false)
//	c.Assert(tree.Contains("/a/bb/"), Equals, false)
//	c.Assert(tree.Contains("/a/bbb"), Equals, false)
//	c.Assert(tree.Contains("/a/bbbb/"), Equals, false)
//}
//
//func (s *S) TestPathSelectionGlobs(c *C) {
//	tree := slicer.PathSelection[bool, any]{}
//	tree.Init()
//
//	tree.Insert("/foo*", nil)
//
//	c.Assert(tree.Contains("/"), Equals, true)
//	c.Assert(tree.Contains("/fo"), Equals, false)
//	c.Assert(tree.Contains("/foo"), Equals, true)
//	c.Assert(tree.Contains("/fooo"), Equals, true)
//	c.Assert(tree.Contains("/foo/"), Equals, false)
//	c.Assert(tree.Contains("/fooo/"), Equals, false)
//
//	tree.Insert("/fo*", nil)
//
//	c.Assert(tree.Contains("/fo"), Equals, true)
//	c.Assert(tree.Contains("/foo"), Equals, true)
//	c.Assert(tree.Contains("/fooo"), Equals, true)
//	c.Assert(tree.Contains("/fo/"), Equals, false)
//	c.Assert(tree.Contains("/foo/"), Equals, false)
//	c.Assert(tree.Contains("/fooo/"), Equals, false)
//
//	tree.Insert("/foo", nil)
//
//	c.Assert(tree.Contains("/fo"), Equals, true)
//	c.Assert(tree.Contains("/foo"), Equals, true)
//	c.Assert(tree.Contains("/fooo"), Equals, true)
//	c.Assert(tree.Contains("/fo/"), Equals, false)
//	c.Assert(tree.Contains("/foo/"), Equals, false)
//	c.Assert(tree.Contains("/fooo/"), Equals, false)
//
//	tree.Insert("/fo/bar", nil)
//
//	c.Assert(tree.Contains("/fo"), Equals, true)
//	c.Assert(tree.Contains("/foo"), Equals, true)
//	c.Assert(tree.Contains("/fooo"), Equals, true)
//	c.Assert(tree.Contains("/fo/"), Equals, true)
//	c.Assert(tree.Contains("/foo/"), Equals, false)
//	c.Assert(tree.Contains("/fooo/"), Equals, false)
//}
//
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
//func (s *S) TestPathSelectionReturnValue(c *C) {
//	var value *slicer.PathValue[string]
//	tree := slicer.PathSelection[string, string]{}
//	tree.UpdateUserData = slicer.ReplaceUserData[string]
//	tree.Init()
//
//	value, _ = tree.Insert("/a/b/c", "A")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/c")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.UserData, Equals, "A")
//
//	value, _ = tree.Insert("/a/b/c", "B")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/c")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.UserData, Equals, "B")
//
//	// modyfing hooks after Init() is unsupported, but test it anyway
//	tree.UpdateUserData = nil
//
//	value, _ = tree.Insert("/a/b/c", "C")
//	c.Assert(value, NotNil)
//	c.Assert(value.Path, Equals, "/a/b/c")
//	c.Assert(value.Implicit, Equals, false)
//	c.Assert(value.UserData, Equals, "B")
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
