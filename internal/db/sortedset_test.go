package db_test

import (
	"github.com/canonical/chisel/internal/db"
	. "gopkg.in/check.v1"
)

type stringStortedSetAddStringTest struct {
	set    db.StringSortedSet
	add    string
	added  bool
	result db.StringSortedSet
}

var stringStortedSetAddStringTests = []stringStortedSetAddStringTest{
	{[]string{"a", "b", "c"}, "", true, []string{"", "a", "b", "c"}},
	{[]string{"a", "b", "c"}, "a", false, []string{"a", "b", "c"}},
	{[]string{"b", "d"}, "a", true, []string{"a", "b", "d"}},
	{[]string{"b", "d"}, "c", true, []string{"b", "c", "d"}},
	{[]string{"b", "d"}, "e", true, []string{"b", "d", "e"}},
	{[]string{"a", "b", "b", "c"}, "b", false, []string{"a", "b", "b", "c"}},
	{[]string{}, "a", true, []string{"a"}},
	{nil, "a", true, []string{"a"}},
}

func (s *S) TestStringSortedSetAddString(c *C) {
	for _, test := range stringStortedSetAddStringTests {
		result, added := test.set.AddString(test.add)
		c.Assert(result, DeepEquals, test.result)
		c.Assert(added, DeepEquals, test.added)
	}
}

type stringStortedSetAddStringsTest struct {
	set    db.StringSortedSet
	add    []string
	result db.StringSortedSet
}

var stringStortedSetAddStringsTests = []stringStortedSetAddStringsTest{
	{[]string{"b", "d"}, []string{}, []string{"b", "d"}},
	{[]string{"b", "d"}, nil, []string{"b", "d"}},
	{[]string{}, []string{}, []string{}},
	{nil, []string{}, nil},
	{nil, nil, nil},
	{[]string{}, []string{"a", "b"}, []string{"a", "b"}},
	{nil, []string{"a", "b"}, []string{"a", "b"}},
	{[]string{"b", "d"}, []string{"c", "a"}, []string{"a", "b", "c", "d"}},
	{[]string{"b", "d"}, []string{"c", "c"}, []string{"b", "c", "d"}},
	{[]string{"b", "d"}, []string{"b", "a", "b"}, []string{"a", "b", "d"}},
}

func (s *S) TestStringSortedSetAddStrings(c *C) {
	for _, test := range stringStortedSetAddStringsTests {
		result := test.set.AddStrings(test.add...)
		c.Assert(result, DeepEquals, test.result)
	}
}
