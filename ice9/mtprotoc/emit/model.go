// Package emit writes the generated files of proto/mtproto from the schema.
// Each file has one function; the tests in ice9/mtprotoc compare every file
// with the tree byte for byte.
package emit

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	Services []*Service         // services.txt: the gRPC service each method is answered from
	Extras   map[string][]Field // extra_fields.txt: fields teamgram adds, by message
	owners   map[*schema.Version]*schema.Entity
}

// Service is one gRPC service and the methods it answers, in order.
type Service struct {
	Name    string
	Methods []*schema.Entity
}

// Load reads the schema directory.
func Load(dir string) (*Model, error) {
	paths := []string{filepath.Join(dir, "layers.tl"), filepath.Join(dir, "mtproto.tl"), filepath.Join(dir, "teamgram.tl")}
	s, err := schema.ParseFiles(paths...)
	if err != nil {
		return nil, err
	}
	m := &Model{Top: s.Top, All: s, Names: schema.PredicateNames(s)}
	byPredicate := map[string]*schema.Entity{}
	for _, e := range s.Entities {
		byPredicate[m.Names[e]] = e
		switch e.File {
		case paths[0]:
			m.API = append(m.API, e)
		case paths[1]:
			m.Service = append(m.Service, e)
		case paths[2]:
			m.Teamgram = append(m.Teamgram, e)
		}
	}
	if err := m.loadServices(filepath.Join(dir, "services.txt"), byPredicate); err != nil {
		return nil, err
	}
	extras, err := loadExtras(filepath.Join(dir, "extra_fields.txt"))
	if err != nil {
		return nil, err
	}
	m.Extras = extras
	return m, nil
}

// loadServices reads "RPCService method" lines, in order.
func (m *Model) loadServices(path string, byPredicate map[string]*schema.Entity) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var current *Service
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return fmt.Errorf("%s: a line is a service and a method: %s", path, line)
		}
		e := byPredicate[fields[1]]
		if e == nil || !e.Method {
			return fmt.Errorf("%s: %s is not a method of the schema", path, fields[1])
		}
		if current == nil || current.Name != fields[0] {
			current = &Service{Name: fields[0]}
			m.Services = append(m.Services, current)
		}
		current.Methods = append(current.Methods, e)
	}
	return scanner.Err()
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
	{"class_id_registers.tl.go", ClassIDRegisters, false},
	{"rpc_client_registers.go", RPCClientRegisters, false},
	{"schema.tl.core_types.proto", CoreTypesProto, true},
	{"schema.tl.handshake.proto", HandshakeProto, true},
	{"schema.tl.transport.proto", TransportProto, true},
	{"schema.tl.sync.proto", SyncProto, true},
	{"schema.tl.handshake_service.proto", HandshakeServiceProto, true},
	{"schema.tl.transport_service.proto", TransportServiceProto, true},
	{"schema.tl.sync_service.proto", SyncServiceProto, true},
	{"schema.tl.crc32_helper.pb.go", CRC32Helper, false},
	{"codec_schema.tl.pb.go", Codec, false},
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
