# SPDX-License-Identifier: MIT
# Copyright (c) 2018 QRI, Inc.
# Copyright (c) 2022 Canonical Ltd.

load("assert.star", "assert")
load("re.star", "re")

def test_regex(fn, pattern, args, expect):
    assert.eq(getattr(re, fn)(pattern, *args), expect)
    assert.eq(getattr(re.compile(pattern), fn)(*args), expect)

test_regex("find",
    r"(\w*)\s*(ADD|REM|DEL|EXT|TRF)\s*(.*)\s*(NAT|INT)\s*(.*)\s*(\(\w{2}\))\s*(.*)",
    ["EDM ADD FROM INJURED NAT Jordan BEAULIEU (DB) Western University"],
    (
      "EDM ADD FROM INJURED NAT Jordan BEAULIEU (DB) Western University",
      "EDM",
      "ADD",
      "FROM INJURED ",
      "NAT",
      "Jordan BEAULIEU ",
      "(DB)",
      "Western University",
    )
)

test_regex("split", r"\s+",
    ["  foo bar   baz\tqux quux ", 5],
    ("", "foo", "bar", "baz", "qux quux ")
)

test_regex("findall", r"\b([A-Z])\w+",
    ["Alpha Bravo charlie DELTA Echo F Hotel"],
    [
        ("Alpha", "A"),
        ("Bravo", "B"),
        ("DELTA", "D"),
        ("Echo", "E"),
        ("Hotel", "H"),
    ]
)
