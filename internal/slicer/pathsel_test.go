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
	sel := slicer.CreatePathSelection()

	sel.AddPath("/var/log/messages")

	c.Assert(sel.IsPathSelected("/"), Equals, true)
	c.Assert(sel.IsPathSelected("/var/"), Equals, true)
	c.Assert(sel.IsPathSelected("/var/log/"), Equals, true)
	c.Assert(sel.IsPathSelected("/var/log/messages"), Equals, true)

	c.Assert(sel.IsPathSelected(""), Equals, false)
	c.Assert(sel.IsPathSelected("./"), Equals, false)
	c.Assert(sel.IsPathSelected("./var"), Equals, false)
	c.Assert(sel.IsPathSelected("//var"), Equals, false)
	c.Assert(sel.IsPathSelected("/var"), Equals, false)
	c.Assert(sel.IsPathSelected("/var/./"), Equals, false)
	c.Assert(sel.IsPathSelected("/var//"), Equals, false)
	c.Assert(sel.IsPathSelected("/var/log"), Equals, false)
	c.Assert(sel.IsPathSelected("/var/log/dmesg"), Equals, false)
	c.Assert(sel.IsPathSelected("/var/log/messages/"), Equals, false)
	c.Assert(sel.IsPathSelected("/zzz"), Equals, false)
}

func (s *S) TestPathSelectionFewPaths(c *C) {
	sel := slicer.CreatePathSelection()

	sel.AddPath("/a/b/c1/d/")
	sel.AddPath("/a/b/c1/d/e")
	sel.AddPath("/a/bbb/c/d/")
	sel.AddPath("/a/b/c1/d/eee")
	sel.AddPath("/a/b/c2/d/e")
	//sel.DumpTree()

	c.Assert(sel.IsPathSelected("/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c1/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c1/d/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c1/d/e"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c1/d/eee"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c2/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c2/d/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/b/c2/d/e"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/bbb/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/bbb/c/"), Equals, true)
	c.Assert(sel.IsPathSelected("/a/bbb/c/d/"), Equals, true)

	c.Assert(sel.IsPathSelected("/a"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/b/c"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/b/c/"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/b/c1/d"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/b/c1/d/e/"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/b/c1/d/ee"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/b/c2"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/bb/"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/bbb"), Equals, false)
	c.Assert(sel.IsPathSelected("/a/bbbb/"), Equals, false)
}
