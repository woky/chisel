package db

import "sort"

type StringSortedSet []string

func (s StringSortedSet) AddString(x string) (StringSortedSet, bool) {
	if s == nil {
		return []string{x}, true
	}
	i := sort.SearchStrings(s, x)
	if i == len(s) {
		s = append(s, x)
	} else if s[i] != x {
		s = append(s[:i], append([]string{x}, s[i:]...)...)
	} else {
		return s, false
	}
	return s, true
}

func (s StringSortedSet) AddStrings(xs ...string) StringSortedSet {
	for _, x := range xs {
		s, _ = s.AddString(x)
	}
	return s
}
