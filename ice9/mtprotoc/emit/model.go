// Package emit writes the generated files of proto/mtproto from the schema.
// Each file has one function; the tests in ice9/mtprotoc compare every file
// with the tree byte for byte.
package emit

import (
	"bytes"
	"go/format"
	"os/exec"
	"path/filepath"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// Model is the schema as the emitters read it: the three files, read in the
// order the generated code lists everything, with each entity's name in that
// code.
type Model struct {
	Top      int
	All      *schema.Schema
	API      []*schema.Entity // layers.tl: Telegram's API, every layer
	Service  []*schema.Entity // mtproto.tl: the MTProto service types, layer 0
	Teamgram []*schema.Entity // teamgram.tl: teamgram's own types and methods, layer 0
	Names    map[*schema.Entity]string
}

// Load reads the schema directory.
func Load(dir string) (*Model, error) {
	paths := []string{filepath.Join(dir, "layers.tl"), filepath.Join(dir, "mtproto.tl"), filepath.Join(dir, "teamgram.tl")}
	s, err := schema.ParseFiles(paths...)
	if err != nil {
		return nil, err
	}
	m := &Model{Top: s.Top, All: s, Names: schema.PredicateNames(s)}
	for _, e := range s.Entities {
		switch e.File {
		case paths[0]:
			m.API = append(m.API, e)
		case paths[1]:
			m.Service = append(m.Service, e)
		case paths[2]:
			m.Teamgram = append(m.Teamgram, e)
		}
	}
	return m, nil
}

// Constructors and Methods split entities, keeping their order.
func Constructors(entities []*schema.Entity) []*schema.Entity { return filter(entities, false) }
func Methods(entities []*schema.Entity) []*schema.Entity      { return filter(entities, true) }

func filter(entities []*schema.Entity, method bool) []*schema.Entity {
	var out []*schema.Entity
	for _, e := range entities {
		if e.Method == method {
			out = append(out, e)
		}
	}
	return out
}

// File is one generated file: how it is emitted and how it is formatted.
type File struct {
	Name  string
	Emit  func(*Model) []byte
	Proto bool // formatted by clang-format with the tree's style; otherwise gofmt
}

// Files is every file the generator writes.
var Files = []File{
	{"schema.tl.crc32.proto", CRC32Proto, true},
	{"class_name_registers.go", ClassNameRegisters, false},
}

// Formatted is the file as it is written: emitted, then formatted as the
// tree is. The path is where it will land, which tells clang-format which
// .clang-format applies.
func (f File) Formatted(m *Model, path string) ([]byte, error) {
	text := f.Emit(m)
	if !f.Proto {
		return format.Source(text)
	}
	cmd := exec.Command("clang-format", "-assume-filename="+path)
	cmd.Stdin = bytes.NewReader(text)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, &formatError{err: err, stderr: stderr.String()}
	}
	return out, nil
}

type formatError struct {
	err    error
	stderr string
}

func (e *formatError) Error() string { return "clang-format: " + e.err.Error() + ": " + e.stderr }
