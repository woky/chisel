package scripts

import (
	"regexp"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var reModuleName = "re.star"

var reModule = starlark.StringDict{
	"re": &starlarkstruct.Module{
		Name: "re",
		Members: starlark.StringDict{
			"compile": starlark.NewBuiltin("compile", compile),
			"find":    starlark.NewBuiltin("find", find),
			"findall": starlark.NewBuiltin("findall", findAll),
			"split":   starlark.NewBuiltin("split", split),
		},
	},
}

func unmarshalInt(in starlark.Int) (int, error) {
	var out int = 0
	err := starlark.AsInt(in, &out)
	return out, err
}

func marshalStringArray(match []string) starlark.Value {
	if match == nil {
		return starlark.None
	}
	resultArray := make([]starlark.Value, len(match))
	for i, str := range match {
		resultArray[i] = starlark.String(str)
	}
	return starlark.Tuple(resultArray)
}

func regexFind(regex *regexp.Regexp, str starlark.String) (starlark.Value, error) {
	match := regex.FindStringSubmatch(string(str))
	return marshalStringArray(match), nil
}

func regexFindAll(regex *regexp.Regexp, str starlark.String, limit starlark.Int) (starlark.Value, error) {
	limitInt, err := unmarshalInt(limit)
	if err != nil {
		return nil, err
	}
	matches := regex.FindAllStringSubmatch(string(str), limitInt)
	resultArray := make([]starlark.Value, len(matches))
	for i, match := range matches {
		resultArray[i] = marshalStringArray(match)
	}
	return starlark.NewList(resultArray), nil
}

func regexSplit(regex *regexp.Regexp, str starlark.String, limit starlark.Int) (starlark.Value, error) {
	limitInt, err := unmarshalInt(limit)
	if err != nil {
		return nil, err
	}
	parts := regex.Split(string(str), limitInt)
	return marshalStringArray(parts), nil
}

func compile(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pattern starlark.String
	if err := starlark.UnpackArgs("compile", args, kwargs, "pattern", &pattern); err != nil {
		return starlark.None, err
	}
	return newRegex(pattern)
}

func find(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pattern, str starlark.String
	if err := starlark.UnpackArgs("find", args, kwargs, "pattern", &pattern, "string", &str); err != nil {
		return starlark.None, err
	}
	regex, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, err
	}
	return regexFind(regex, str)
}

func findAll(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pattern, str starlark.String
	var limit = starlark.MakeInt(-1)
	if err := starlark.UnpackArgs("findall", args, kwargs, "pattern", &pattern, "string", &str, "limit?", &limit); err != nil {
		return starlark.None, err
	}
	regex, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, err
	}
	return regexFindAll(regex, str, limit)
}

func split(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pattern, str starlark.String
	var limit = starlark.MakeInt(-1)
	if err := starlark.UnpackArgs("split", args, kwargs, "pattern", &pattern, "string", &str, "limit?", &limit); err != nil {
		return starlark.None, err
	}
	regex, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, err
	}
	return regexSplit(regex, str, limit)
}

type Regex struct {
	regex *regexp.Regexp
}

func newRegex(pattern starlark.String) (*Regex, error) {
	re, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, err
	}
	return &Regex{regex: re}, nil
}

func (r *Regex) String() string {
	return r.regex.String()
}

func (r *Regex) Type() string {
	return "Regex"
}

func (r *Regex) Freeze() {
}

func (r *Regex) Hash() (uint32, error) {
	return starlark.String(r.regex.String()).Hash()
}

func (r *Regex) Truth() starlark.Bool {
	return true
}

func (r *Regex) Attr(name string) (starlark.Value, error) {
	switch name {
	case "find":
		return starlark.NewBuiltin("Regex.find", r.find), nil
	case "findall":
		return starlark.NewBuiltin("Regex.findall", r.findAll), nil
	case "split":
		return starlark.NewBuiltin("Regex.split", r.split), nil
	}
	return nil, nil
}

func (r *Regex) AttrNames() []string {
	return []string{"find", "findall", "split"}
}

func (r *Regex) find(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str starlark.String
	if err := starlark.UnpackArgs("Regex.find", args, kwargs, "string", &str); err != nil {
		return starlark.None, err
	}
	return regexFind(r.regex, str)
}

func (r *Regex) findAll(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str starlark.String
	var limit = starlark.MakeInt(-1)
	if err := starlark.UnpackArgs("Regex.findall", args, kwargs, "string", &str, "limit?", &limit); err != nil {
		return starlark.None, err
	}
	return regexFindAll(r.regex, str, limit)
}

func (r *Regex) split(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str starlark.String
	var limit = starlark.MakeInt(-1)
	if err := starlark.UnpackArgs("Regex.split", args, kwargs, "string", &str, "limit?", &limit); err != nil {
		return starlark.None, err
	}
	return regexSplit(r.regex, str, limit)
}
