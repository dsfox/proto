// Package emit writes the generated files of proto/mtproto from the schema.
// Each file has one function; the tests in ice9/mtprotoc compare every file
// with the tree byte for byte.
package emit

import (
	"path/filepath"
	"sort"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// Model is the schema as the emitters read it: the three files, joined in
// the order the generated code lists everything, with each entity's name in
// that code.
type Model struct {
	Top      int
	API      *schema.Schema // layers.tl: Telegram's API, every layer
	Service  *schema.Schema // mtproto.tl: the MTProto service types, layer 0
	Teamgram *schema.Schema // teamgram.tl: teamgram's own types and methods, layer 0
	All      *schema.Schema // API, then Service, then Teamgram
	Names    map[*schema.Entity]string
}

// Load reads the schema directory.
func Load(dir string) (*Model, error) {
	m := &Model{All: &schema.Schema{}}
	for _, part := range []struct {
		file string
		into **schema.Schema
	}{{"layers.tl", &m.API}, {"mtproto.tl", &m.Service}, {"teamgram.tl", &m.Teamgram}} {
		s, err := schema.ParseFile(filepath.Join(dir, part.file))
		if err != nil {
			return nil, err
		}
		*part.into = s
		m.All.Append(s)
	}
	m.Top = m.All.Layer
	m.Names = schema.PredicateNames(m.All)
	return m, nil
}

// IDs is every id the entity ever had, newest first: the id listed at the
// highest layer, then the next one down. That is the order the generated
// files list a predicate's ids in.
func (m *Model) IDs(e *schema.Entity) []uint32 {
	newest := map[uint32]int{}
	for layer, id := range e.Listing(m.Top) {
		if last, seen := newest[id]; !seen || layer > last {
			newest[id] = layer
		}
	}
	ids := make([]uint32, 0, len(newest))
	for id := range newest {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return newest[ids[i]] > newest[ids[j]] })
	return ids
}

// Version is the version of the entity that has this id.
func (m *Model) Version(e *schema.Entity, id uint32) *schema.Version {
	for i := range e.Versions {
		if e.Versions[i].ID == id {
			return &e.Versions[i]
		}
	}
	return nil
}

// Constructors and Methods split a schema's entities, keeping their order.
func Constructors(s *schema.Schema) []*schema.Entity { return filter(s, false) }
func Methods(s *schema.Schema) []*schema.Entity      { return filter(s, true) }

func filter(s *schema.Schema, method bool) []*schema.Entity {
	var out []*schema.Entity
	for _, e := range s.Entities {
		if e.Method == method {
			out = append(out, e)
		}
	}
	return out
}
