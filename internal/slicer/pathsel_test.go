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

func (s *S) TestPathSelectionSinglePath(c *C) {
	sel := slicer.CreatePathSelection[bool, any]()

	sel.AddPath("/var/log/messages")

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
	sel := slicer.CreatePathSelection[bool, any]()

	sel.AddPath("/a/b/c1/d/")
	sel.AddPath("/a/b/c1/d/e")
	sel.AddPath("/a/bbb/c/d/")
	sel.AddPath("/a/b/c1/d/eee")
	sel.AddPath("/a/b/c2/d/e")
	//sel.DumpTree()

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
	sel := slicer.CreatePathSelection[bool, any]()

	sel.AddPath("/foo*")

	c.Assert(sel.ContainsPath("/"), Equals, true)
	c.Assert(sel.ContainsPath("/fo"), Equals, false)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	sel.AddPath("/fo*")

	c.Assert(sel.ContainsPath("/fo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/fo/"), Equals, false)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	sel.AddPath("/foo")

	c.Assert(sel.ContainsPath("/fo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/fo/"), Equals, false)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	sel.AddPath("/fo/bar")

	c.Assert(sel.ContainsPath("/fo"), Equals, true)
	c.Assert(sel.ContainsPath("/foo"), Equals, true)
	c.Assert(sel.ContainsPath("/fooo"), Equals, true)
	c.Assert(sel.ContainsPath("/fo/"), Equals, true)
	c.Assert(sel.ContainsPath("/foo/"), Equals, false)
	c.Assert(sel.ContainsPath("/fooo/"), Equals, false)

	//sel.DumpTree()
}

func (s *S) TestePathSelectionFindPath(c *C) {
	sel := slicer.CreatePathSelection[bool, any]()
	var value *slicer.PathValue[bool]

	sel.AddPath("/a/b/c")

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

	sel.AddPath("/a/b/cc/")
	sel.AddPath("/a/bb/c")

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

	sel.AddPath("/a/b*")
	sel.AddPath("/a/bbb")

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

	sel.AddPath("/a**/b")

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
