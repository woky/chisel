package db_test

import (
	"encoding/json"
	"reflect"

	"github.com/canonical/chisel/internal/db"
	. "gopkg.in/check.v1"
)

type roundTripTestCase struct {
	value any
	json  string
}

var roundTripTestCases = []roundTripTestCase{{
	db.Package{"foo", "bar", "", ""},
	`{"kind":"package","name":"foo","version":"bar","sha256":"","arch":""}`,
}, {
	db.Package{"coolutils", "1.2-beta", "bbbaa816dedc5c5d58e30e7ed600fe62ea0208ef2d6fac4da8b312d113401958", "all"},
	`{"kind":"package","name":"coolutils","version":"1.2-beta","sha256":"bbbaa816dedc5c5d58e30e7ed600fe62ea0208ef2d6fac4da8b312d113401958","arch":"all"}`,
}, {
	db.Package{"badcowboys", "7", "d0aba6d028cd4a3fd153eb5e0bfb35c33f4d5674b80a7a827917df40e1192424", "amd64"},
	`{"kind":"package","name":"badcowboys","version":"7","sha256":"d0aba6d028cd4a3fd153eb5e0bfb35c33f4d5674b80a7a827917df40e1192424","arch":"amd64"}`,
}, {
	db.Slice{"elitestrike_bins"},
	`{"kind":"slice","name":"elitestrike_bins"}`,
}, {
	db.Slice{"invalid but unmarshals"},
	`{"kind":"slice","name":"invalid but unmarshals"}`,
}, {
	db.Path{
		Path:   "/bin/snake",
		Mode:   0755,
		Slices: []string{"snake_bins"},
		SHA256: "a01bab26f08ba87b86736368b0684f0849a365ab4c5ec546d973ca87c815f682",
		Size:   13,
	},
	`{"kind":"path","path":"/bin/snake","mode":"0755","slices":["snake_bins"],"sha256":"a01bab26f08ba87b86736368b0684f0849a365ab4c5ec546d973ca87c815f682","size":13}`,
}, {
	db.Path{
		Path:   "/etc/default/",
		Mode:   0750,
		Slices: []string{"someconfig_data", "mytoo_data"},
	},
	`{"kind":"path","path":"/etc/default/","mode":"0750","slices":["someconfig_data","mytoo_data"]}`,
}, {
	db.Path{
		Path:        "/var/lib/matt/index.data",
		Mode:        0600,
		Slices:      []string{"daemon_data"},
		SHA256:      "0682c5f2076f099c34cfdd15a9e063849ed437a49677e6fcc5b4198c76575be5",
		FinalSHA256: "d7d5dcc369426e2e5f8dcb89af4308b0daed6e55910d53395ce38bd6dd1a9456",
		Size:        999,
	},
	`{"kind":"path","path":"/var/lib/matt/index.data","mode":"0600","slices":["daemon_data"],"sha256":"0682c5f2076f099c34cfdd15a9e063849ed437a49677e6fcc5b4198c76575be5","final_sha256":"d7d5dcc369426e2e5f8dcb89af4308b0daed6e55910d53395ce38bd6dd1a9456","size":999}`,
}, {
	db.Path{
		Path:   "/lib",
		Mode:   0777,
		Slices: []string{"libc6_libs", "zlib1g_libs"},
		Link:   "/usr/lib/",
	},
	`{"kind":"path","path":"/lib","mode":"0777","slices":["libc6_libs","zlib1g_libs"],"link":"/usr/lib/"}`,
}, {
	db.Path{},
	`{"kind":"path","path":"","mode":"0","slices":[]}`,
}, {
	db.Path{Mode: 077777},
	`{"kind":"path","path":"","mode":"077777","slices":[]}`,
}, {
	db.Content{"foo_sl", "/a/b/c"},
	`{"kind":"content","slice":"foo_sl","path":"/a/b/c"}`,
}}

func (s *S) TestMarshalUnmarshalRoundTrip(c *C) {
	for i, test := range roundTripTestCases {
		c.Logf("Test #%d", i)
		data, err := json.Marshal(test.value)
		c.Assert(err, IsNil)
		c.Assert(string(data), DeepEquals, test.json)
		ptrOut := reflect.New(reflect.ValueOf(test.value).Type())
		err = json.Unmarshal(data, ptrOut.Interface())
		c.Assert(err, IsNil)
		c.Assert(ptrOut.Elem().Interface(), DeepEquals, test.value)
	}
}

type unmarshalTestCase struct {
	value any
	json  string
	error string
}

var unmarshalTestCases = []unmarshalTestCase{{
	value: db.Package{"pkg", "1.1", "d0aba6d028cd4a3fd153eb5e0bfb35c33f4d5674b80a7a827917df40e1192424", "all"},
	json:  `{"kind":"package","name":"pkg","version":"1.1","sha256":"d0aba6d028cd4a3fd153eb5e0bfb35c33f4d5674b80a7a827917df40e1192424","arch":"all"}`,
}, {
	value: db.Slice{"a"},
	json:  `{"kind":"slice","name":"a"}`,
}, {
	value: db.Path{
		Path:        "/x/y/z",
		Mode:        0644,
		Slices:      []string{"pkg1_data", "pkg2_data"},
		SHA256:      "f177b37f18f5bc6596878f074721d796c2333d95f26ce1e45c5a096c350a1c07",
		FinalSHA256: "61bd495076999a77f75288fcfcdd76073ec4aa114632a310b3b3263c498e12f7",
		Size:        123,
	},
	json: `{"kind":"path","path":"/x/y/z","mode":"0644","slices":["pkg1_data","pkg2_data"],"sha256":"f177b37f18f5bc6596878f074721d796c2333d95f26ce1e45c5a096c350a1c07","final_sha256":"61bd495076999a77f75288fcfcdd76073ec4aa114632a310b3b3263c498e12f7","size":123}`,
}, {
	value: db.Path{Path: "/x/y/z", Mode: 0777, Link: "/home"},
	json:  `{"kind":"path","path":"/x/y/z","mode":"0777","slices":[],"link":"/home"}`,
}, {
	value: db.Path{Path: "/x/y/z"},
	json:  `{"kind":"path","path":"/x/y/z","mode":"0","slices":null}`,
}, {
	value: db.Path{Path: "/x/y/z"},
	json:  `{"kind":"path","path":"/x/y/z","mode":"0"}`,
}, {
	value: db.Content{Slice: "pkg_sl", Path: "/a/b/c"},
	json:  `{"kind":"content","slice":"pkg_sl","path":"/a/b/c"}`,
}, {
	value: db.Path{},
	json:  `{"kind":"path","path":"","mode":"90909"}`,
	error: `invalid mode "90909": strconv.ParseUint: parsing "90909": invalid syntax`,
}, {
	value: db.Path{},
	json:  `{"kind":"path","path":"","mode":"077777777777"}`,
	error: `invalid mode "077777777777": strconv.ParseUint: parsing "077777777777": value out of range`,
}, {
	value: db.Slice{},
	json:  `{"kind":"package","name":"foo_libs"}`,
	error: `invalid kind "package": must be "slice"`,
}}

func (s *S) TestUnmarshal(c *C) {
	for i, test := range unmarshalTestCases {
		c.Logf("Test #%d", i)
		ptrOut := reflect.New(reflect.ValueOf(test.value).Type())
		err := json.Unmarshal([]byte(test.json), ptrOut.Interface())
		if test.error != "" {
			c.Assert(err, ErrorMatches, test.error)
		} else {
			c.Assert(ptrOut.Elem().Interface(), DeepEquals, test.value)
		}
	}
}
