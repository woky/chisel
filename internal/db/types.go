package db

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"
)

type Package struct {
	Name    string
	Version string
	SHA256  string
	Arch    string
}

type jsonPackage struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	Arch    string `json:"arch"`
}

func (p Package) MarshalJSON() ([]byte, error) {
	m := jsonPackage{
		Kind:    "package",
		Name:    p.Name,
		Version: p.Version,
		SHA256:  p.SHA256,
		Arch:    p.Arch,
	}
	return json.Marshal(m)
}

func (p *Package) UnmarshalJSON(data []byte) error {
	m := jsonPackage{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m.Kind != "package" {
		return fmt.Errorf("invalid kind %#v: must be \"package\"", m.Kind)
	}
	p.Name = m.Name
	p.Version = m.Version
	p.SHA256 = m.SHA256
	p.Arch = m.Arch
	return nil
}

type Slice struct {
	Name string
}

type jsonSlice struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func (s Slice) MarshalJSON() ([]byte, error) {
	m := jsonSlice{
		Kind: "slice",
		Name: s.Name,
	}
	return json.Marshal(m)
}

func (s *Slice) UnmarshalJSON(data []byte) error {
	m := jsonSlice{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m.Kind != "slice" {
		return fmt.Errorf("invalid kind %#v: must be \"slice\"", m.Kind)
	}
	s.Name = m.Name
	return nil
}

type Path struct {
	Path        string
	Mode        fs.FileMode
	Slices      StringSortedSet
	SHA256      string
	FinalSHA256 string
	Size        int64
	Link        string
}

type jsonPath struct {
	Kind        string   `json:"kind"`
	Path        string   `json:"path"`
	Mode        string   `json:"mode"`
	Slices      []string `json:"slices"`
	SHA256      string   `json:"sha256,omitempty"`
	FinalSHA256 string   `json:"final_sha256,omitempty"`
	Size        *int64   `json:"size,omitempty"`
	Link        string   `json:"link,omitempty"`
}

func (p Path) MarshalJSON() ([]byte, error) {
	m := jsonPath{
		Kind:        "path",
		Path:        p.Path,
		Mode:        fmt.Sprintf("%#o", p.Mode),
		Slices:      p.Slices,
		SHA256:      p.SHA256,
		FinalSHA256: p.FinalSHA256,
		Link:        p.Link,
	}
	if p.Slices == nil {
		m.Slices = []string{}
	}
	if p.SHA256 != "" {
		m.Size = &p.Size
	}
	return json.Marshal(m)
}

func (p *Path) UnmarshalJSON(data []byte) error {
	m := jsonPath{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m.Kind != "path" {
		return fmt.Errorf("invalid kind %#v: must be \"path\"", m.Kind)
	}
	mode, err := strconv.ParseUint(m.Mode, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid mode %#v: %w", m.Mode, err)
	}
	p.Path = m.Path
	p.Mode = fs.FileMode(mode)
	p.Slices = m.Slices
	if p.Slices != nil && len(p.Slices) == 0 {
		p.Slices = nil
	}
	p.SHA256 = m.SHA256
	p.FinalSHA256 = m.FinalSHA256
	p.Size = 0
	if m.Size != nil {
		p.Size = *m.Size
	}
	p.Link = m.Link
	return nil
}

type Content struct {
	Slice string
	Path  string
}

type jsonContent struct {
	Kind  string `json:"kind"`
	Slice string `json:"slice"`
	Path  string `json:"path"`
}

func (p Content) MarshalJSON() ([]byte, error) {
	m := jsonContent{"content", p.Slice, p.Path}
	return json.Marshal(m)
}

func (p *Content) UnmarshalJSON(data []byte) error {
	m := jsonContent{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m.Kind != "content" {
		return fmt.Errorf("invalid kind %#v: must be \"content\"", m.Kind)
	}
	p.Slice = m.Slice
	p.Path = m.Path
	return nil
}
