package db_test

import (
	"sort"

	"github.com/canonical/chisel/internal/db"
	. "gopkg.in/check.v1"
)

type testEntry struct {
	T string          `json:"t"`
	S string          `json:"s,omitempty"`
	I int64           `json:"i,omitempty"`
	L []string        `json:"l,omitempty"`
	M map[string]bool `json:"m,omitempty"`
	K int             `json:"k,omitempty"`
}

var saveLoadTestCase = []testEntry{
	{"dummy", "hello", -1, nil, nil, 0},
	{"dummy", "", 100, []string{"a", "b"}, nil, 0},
	{"dummy", "", 0, nil, map[string]bool{"a": true, "b": false}, 0},
}

func (s *S) TestSaveLoad(c *C) {
	sortEntries := func(entries []testEntry) {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].K < entries[j].K
		})
	}
	entriesIn := make([]testEntry, len(saveLoadTestCase))
	for i, entry := range saveLoadTestCase {
		entry.K = i
		entriesIn[i] = entry
	}
	sortEntries(entriesIn)

	workDir := c.MkDir()
	dbw := db.New()
	for _, entry := range entriesIn {
		err := dbw.Add(entry)
		c.Assert(err, IsNil)
	}
	err := db.Save(dbw, workDir)
	c.Assert(err, IsNil)

	dbr, err := db.Load(workDir)
	c.Assert(err, IsNil)
	c.Assert(dbr.Schema(), Equals, db.Schema)

	iter, err := dbr.Iterate(&testEntry{T: "dummy"})
	c.Assert(err, IsNil)
	entriesOut := make([]testEntry, 0, len(entriesIn))
	for iter.Next() {
		var entry testEntry
		err := iter.Get(&entry)
		c.Assert(err, IsNil)
		entriesOut = append(entriesOut, entry)
	}
	sortEntries(entriesOut)
	c.Assert(entriesOut, DeepEquals, entriesIn)
}
