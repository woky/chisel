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

func (s *S) TestStripLeadingEmptyPath(c *C) {
	var result string
	var err error

	result, err = slicer.StripLeadingEmptyPath("abc")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "abc")

	result, err = slicer.StripLeadingEmptyPath("a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("/a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("./a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("/./a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("//a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("/////a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("./././a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("/./././a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath(".//.///././/a/b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/b/c")

	result, err = slicer.StripLeadingEmptyPath("./a/./b/c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a/./b/c")

	result, err = slicer.StripLeadingEmptyPath("a///b/./c")
	c.Assert(err, IsNil)
	c.Assert(result, Equals, "a///b/./c")

	result, err = slicer.StripLeadingEmptyPath("../a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.StripLeadingEmptyPath("/../a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.StripLeadingEmptyPath("././//./.././a/b/c")
	c.Assert(err, NotNil)

	result, err = slicer.StripLeadingEmptyPath("////../a/b/c")
	c.Assert(err, NotNil)
}

func (s *S) TestPathSelectionSinglePath(c *C) {
	sel := slicer.PathSelection[bool, any]{}
	sel.Init()

	sel.AddPath("/var/log/messages", nil)

	c.Assert(sel.ContainsPath("/"), Equals, true)
	c.Assert(sel.ContainsPath("/var/"), Equals, true)
	c.Assert(sel.ContainsPath("/var/log/"), Equals, true)
	c.Assert(sel.ContainsPath("/var/log/messages"), Equals, true)

	c.Assert(sel.ContainsPath(""), Equals, false)
	c.Assert(sel.ContainsPath("./"), Equals, false)
	c.Assert(sel.ContainsPath("./var"), Equals, false)
	c.Assert(sel.ContainsPath("//var"), Equals, false)
	c.Assert(sel.ContainsPath("/var"), Equals, false)
	c.Assert(sel.ContainsPath("/var/./"), Equals, false)
	c.Assert(sel.ContainsPath("/var//"), Equals, false)
	c.Assert(sel.ContainsPath("/var/log"), Equals, false)
	c.Assert(sel.ContainsPath("/var/log/dmesg"), Equals, false)
	c.Assert(sel.ContainsPath("/var/log/messages/"), Equals, false)
	c.Assert(sel.ContainsPath("/zzz"), Equals, false)
}

func (s *S) TestPathSelectionFewPaths(c *C) {
	sel := slicer.PathSelection[bool, any]{}
	sel.Init()

	sel.AddPath("/a/b/c1/d/", nil)
	sel.AddPath("/a/b/c1/d/e", nil)
	sel.AddPath("/a/bbb/c/d/", nil)
	sel.AddPath("/a/b/c1/d/eee", nil)
	sel.AddPath("/a/b/c2/d/e", nil)

	c.Assert(sel.ContainsPath("/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c1/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c1/d/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c1/d/e"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c1/d/eee"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c2/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c2/d/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/b/c2/d/e"), Equals, true)
	c.Assert(sel.ContainsPath("/a/bbb/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/bbb/c/"), Equals, true)
	c.Assert(sel.ContainsPath("/a/bbb/c/d/"), Equals, true)

	c.Assert(sel.ContainsPath("/a"), Equals, false)
	c.Assert(sel.ContainsPath("/a/b/c"), Equals, false)
	c.Assert(sel.ContainsPath("/a/b/c/"), Equals, false)
	c.Assert(sel.ContainsPath("/a/b/c1/d"), Equals, false)
	c.Assert(sel.ContainsPath("/a/b/c1/d/e/"), Equals, false)
	c.Assert(sel.ContainsPath("/a/b/c1/d/ee"), Equals, false)
	c.Assert(sel.ContainsPath("/a/b/c2"), Equals, false)
	c.Assert(sel.ContainsPath("/a/bb/"), Equals, false)
	c.Assert(sel.ContainsPath("/a/bbb"), Equals, false)
	c.Assert(sel.ContainsPath("/a/bbbb/"), Equals, false)
}

func (s *S) TestPathSelectionGlobs(c *C) {
	sel := slicer.PathSelection[bool, any]{}
	sel.Init()

	sel.AddPath("/foo*", nil)

	c.Assert(sel.ContainsPath("/"), Equals, true)
	c.Assert(sel.ContainsPath("/fo"), Equals, false)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	sel.AddPath("/fo*", nil)

	c.Assert(sel.ContainsPath("/fo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/fo/"), Equals, false)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	sel.AddPath("/foo", nil)

	c.Assert(sel.ContainsPath("/fo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/fo/"), Equals, false)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	sel.AddPath("/fo/bar", nil)

	c.Assert(sel.ContainsPath("/fo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/fo/"), Equals, true)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)
}

func (s *S) TestPathSelectionFindPath(c *C) {
	var value *slicer.PathValue[bool]
	sel := slicer.PathSelection[bool, any]{}
	sel.Init()

	sel.AddPath("/a/b/c", nil)

	value = sel.FindPath("/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/")
	c.Assert(value.Implicit, Equals, true)

	value = sel.FindPath("/a/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/")
	c.Assert(value.Implicit, Equals, true)

	value = sel.FindPath("/a/b/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/")
	c.Assert(value.Implicit, Equals, true)

	value = sel.FindPath("/a/b/c")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/c")
	c.Assert(value.Implicit, Equals, false)

	sel.AddPath("/a/b/cc/", nil)
	sel.AddPath("/a/bb/c", nil)

	value = sel.FindPath("/a/b/cc/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/cc/")
	c.Assert(value.Implicit, Equals, false)

	value = sel.FindPath("/a/bb/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/bb/")
	c.Assert(value.Implicit, Equals, true)

	value = sel.FindPath("/a/bb/c")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/bb/c")
	c.Assert(value.Implicit, Equals, false)

	value = sel.FindPath("/a/bb/c")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/bb/c")
	c.Assert(value.Implicit, Equals, false)

	c.Assert(sel.FindPath("/a"), IsNil)
	c.Assert(sel.FindPath("/a/b"), IsNil)
	c.Assert(sel.FindPath("/a/bb"), IsNil)
	c.Assert(sel.FindPath("/a/b/c/"), IsNil)
	c.Assert(sel.FindPath("/zzz"), IsNil)

	sel.AddPath("/a/b*", nil)
	sel.AddPath("/a/bbb", nil)

	value = sel.FindPath("/a/bbb")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/bbb")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.PathIsGlob, Equals, false)

	value = sel.FindPath("/a/b/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/")
	c.Assert(value.Implicit, Equals, true)
	c.Assert(value.PathIsGlob, Equals, false)

	value = sel.FindPath("/a/bb/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/bb/")
	c.Assert(value.Implicit, Equals, true)
	c.Assert(value.PathIsGlob, Equals, false)

	value = sel.FindPath("/a/bbbb")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b*")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.PathIsGlob, Equals, true)

	c.Assert(sel.FindPath("/a/bbbb/"), IsNil)

	sel.AddPath("/a**/b", nil)

	value = sel.FindPath("/a/b/")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/")
	c.Assert(value.Implicit, Equals, true)
	c.Assert(value.PathIsGlob, Equals, false)

	value = sel.FindPath("/aa/b")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a**/b")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.PathIsGlob, Equals, true)

	value = sel.FindPath("/aa/b/c/b")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a**/b")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.PathIsGlob, Equals, true)

	c.Assert(sel.FindPath("/aa/b/"), IsNil)
	c.Assert(sel.FindPath("/aa/b/c"), IsNil)
}

func (s *S) TestPathSelectionReturnValue(c *C) {
	var value *slicer.PathValue[string]
	sel := slicer.PathSelection[string, string]{}
	sel.UpdateUserData = slicer.ReplaceUserData[string]
	sel.Init()

	value, _ = sel.AddPath("/a/b/c", "A")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/c")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.UserData, Equals, "A")

	value, _ = sel.AddPath("/a/b/c", "B")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/c")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.UserData, Equals, "B")

	// modyfing hooks after Init() is unsupported, but test it anyway
	sel.UpdateUserData = nil

	value, _ = sel.AddPath("/a/b/c", "C")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/a/b/c")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.UserData, Equals, "B")
}

func (s *S) TestPathSelectionParent(c *C) {
	var value *slicer.PathValue[string]
	sel := slicer.PathSelection[string, string]{}
	sel.UpdateUserData = slicer.ReplaceUserData[string]
	sel.UpdateImplicitUserData = sel.UpdateUserData
	sel.Init()

	sel.AddPath("/a/b", "A")

	value, _ = sel.AddPath("/x/y/z", "X")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/x/y/z")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.UserData, Equals, "X")

	value = value.Parent
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/x/y/")
	c.Assert(value.Implicit, Equals, true)
	c.Assert(value.UserData, Equals, "X")

	value = value.Parent
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/x/")
	c.Assert(value.Implicit, Equals, true)
	c.Assert(value.UserData, Equals, "X")

	value = value.Parent
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/")
	c.Assert(value.Implicit, Equals, true)
	c.Assert(value.UserData, Equals, "X")

	value = value.Parent
	c.Assert(value, IsNil)

	value, _ = sel.AddPath("/x/y/", "Z")
	c.Assert(value, NotNil)
	c.Assert(value.Path, Equals, "/x/y/")
	c.Assert(value.Implicit, Equals, false)
	c.Assert(value.UserData, Equals, "Z")
}

func (s *S) TestPathSelectionUserData(c *C) {
	type PathData struct {
		initCount           int
		updateCount         int
		implicitUpdateCount int
	}
	var value *slicer.PathValue[PathData]
	sel := slicer.PathSelection[PathData, any]{}
	sel.InitUserData = func(value *slicer.PathValue[PathData]) {
		value.UserData.initCount += 1
	}
	sel.UpdateUserData = func(value *slicer.PathValue[PathData], _ any) {
		value.UserData.updateCount += 1
	}
	sel.UpdateImplicitUserData = func(value *slicer.PathValue[PathData], _ any) {
		value.UserData.implicitUpdateCount += 1
	}
	sel.Init()

	sel.AddPath("/a/b/c", 10)
	sel.AddPath("/a/b/c/d", 1)
	sel.AddPath("/a/b/cc/d", 1)
	sel.AddPath("/a/b/", 1)
	sel.AddPath("/", 1)
	sel.AddPath("/a/b/", 1)

	value = sel.FindPath("/")
	c.Assert(value, NotNil)
	//c.Assert(value.UserData.initCount, Equals, 0)
	//c.Assert(value.UserData.updateCount, Equals, 1)
	//c.Assert(value.UserData.implicitUpdateCount, Equals, 5)
}

func (s *S) TestPathSelectionOddities(c *C) {
	sel := slicer.PathSelection[any, any]{}
	sel.Init()

	sel.AddPath("/foo/bar", nil)
	sel.AddPath("/foo/bar/", nil)
	//sel.DumpTree()
	//c.Assert(true, Equals, false)

	//sel.AddPath("", nil)
}

// TODO oddities
// Empty path
