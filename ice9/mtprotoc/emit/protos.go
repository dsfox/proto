package emit

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// coreTypes have a .proto of their own (schema.tl.core_types.proto).
var coreTypes = map[string]bool{
	"Bool": true, "Double": true, "Error": true, "Int32": true, "Int64": true, "Null": true, "String": true, "True": true,
}

// protoType is one TL type as a message: its constructors, and every
// definition line of theirs in schema order, which is what the message's
// fields are the union of.
type protoType struct {
	name         string // the TL type
	constructors []*schema.Entity
	lines        []*schema.Version
}

// types groups every constructor of the schema by its type, in the order
// the types are first seen; a type's lines are in section order.
func (m *Model) types() []*protoType {
	var types []*protoType
	byName := map[string]*protoType{}
	for _, sec := range m.All.Sections {
		for _, l := range sec.Lines {
			if l.Version == nil || l.Entity.Method {
				continue
			}
			t := byName[l.Version.Result]
			if t == nil {
				t = &protoType{name: l.Version.Result}
				byName[t.name] = t
				types = append(types, t)
			}
			t.lines = append(t.lines, l.Version)
			if len(t.constructors) == 0 || t.constructors[len(t.constructors)-1] != l.Entity {
				known := false
				for _, e := range t.constructors {
					if e == l.Entity {
						known = true
						break
					}
				}
				if !known {
					t.constructors = append(t.constructors, l.Entity)
				}
			}
		}
	}
	return types
}

// typesIn is the types whose constructors come from these entities.
func typesIn(types []*protoType, entities []*schema.Entity) []*protoType {
	in := map[*schema.Entity]bool{}
	for _, e := range entities {
		in[e] = true
	}
	var out []*protoType
	for _, t := range types {
		if in[t.constructors[0]] {
			out = append(out, t)
		}
	}
	return out
}

func sortTypes(types []*protoType, key func(string) string) []*protoType {
	sorted := append([]*protoType(nil), types...)
	sort.SliceStable(sorted, func(i, j int) bool { return key(sorted[i].name) < key(sorted[j].name) })
	return sorted
}

// unionStyle is how a .proto writes a type: with the "//" line closing the
// constructor list and a "// TL_x" line before each constructor message, or
// without (schema.tl.core_types.proto).
type unionStyle struct {
	commented bool
}

// writeUnion writes one type: the message with every constructor's fields,
// then a message per constructor holding it as data2.
func (m *Model) writeUnion(b *bytes.Buffer, t *protoType, style unionStyle) {
	name := ProtoType(t.name)
	fmt.Fprintf(b, "// %s <--\n", name)
	for _, v := range distinctIDs(t.lines) {
		fmt.Fprintf(b, "//  + TL_%s\n", m.Names[m.entityOf(v)])
	}
	if style.commented {
		b.WriteString("//\n")
	}
	fmt.Fprintf(b, "message %s {\n  string predicate_name = 1;\n  TLConstructor constructor = 2;\n", name)
	for _, f := range unionFields(t.lines, 3) {
		fmt.Fprintf(b, "  %s %s = %d;\n", f.Type, f.Name, f.Number)
	}
	for _, f := range m.Extras[name] {
		fmt.Fprintf(b, "  %s %s = %d;\n", f.Type, f.Name, f.Number)
	}
	b.WriteString("}\n")
	for _, e := range t.constructors {
		b.WriteString("\n")
		if style.commented {
			fmt.Fprintf(b, "// TL_%s\n", m.Names[e])
		}
		fmt.Fprintf(b, "message TL_%s {\n  %s data2 = 1;\n}\n", m.Names[e], name)
	}
}

// entityOf is the entity a version belongs to.
func (m *Model) entityOf(v *schema.Version) *schema.Entity {
	if m.owners == nil {
		m.owners = map[*schema.Version]*schema.Entity{}
		for _, e := range m.All.Entities {
			for _, ev := range e.Versions {
				m.owners[ev] = e
			}
		}
	}
	return m.owners[v]
}

func (m *Model) writeUnions(b *bytes.Buffer, types []*protoType, style unionStyle) {
	for i, t := range types {
		if i > 0 {
			b.WriteString("\n")
		}
		m.writeUnion(b, t, style)
	}
}

// SyncProto is schema.tl.sync.proto: every API and teamgram type but the
// core ones, sorted by message name.
func SyncProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(`
import "google/protobuf/wrappers.proto";
import "schema.tl.crc32.proto";
import "schema.tl.core_types.proto";

option go_package = "github.com/teamgram/proto/mtproto";

`)
	var types []*protoType
	for _, t := range typesIn(m.types(), append(append([]*schema.Entity(nil), Constructors(m.API)...), Constructors(m.Teamgram)...)) {
		if !coreTypes[t.name] {
			types = append(types, t)
		}
	}
	m.writeUnions(&b, sortTypes(types, ProtoType), unionStyle{commented: true})
	return b.Bytes()
}

// CoreTypesProto is schema.tl.core_types.proto: Bool, the numbers, String,
// Error, Null and True.
func CoreTypesProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(serviceImports)
	var types []*protoType
	for _, t := range m.types() {
		if coreTypes[t.name] {
			types = append(types, t)
		}
	}
	m.writeUnions(&b, sortTypes(types, ProtoType), unionStyle{})
	return b.Bytes()
}

const serviceImports = `
// import "google/protobuf/wrappers.proto";
import "schema.tl.crc32.proto";

option go_package = "github.com/teamgram/proto/mtproto";

`

// TransportProto is schema.tl.transport.proto: the service types that are
// not the handshake's, sorted by message name.
func TransportProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(serviceImports)
	var types []*protoType
	for _, t := range typesIn(m.types(), Constructors(m.Service)) {
		if !handshakeTypes[t.name] {
			types = append(types, t)
		}
	}
	m.writeUnions(&b, sortTypes(types, ProtoType), unionStyle{commented: true})
	return b.Bytes()
}

// HandshakeProto is schema.tl.handshake.proto: the key exchange's types,
// sorted by message name.
func HandshakeProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(serviceImports)
	var types []*protoType
	for _, t := range typesIn(m.types(), Constructors(m.Service)) {
		if handshakeTypes[t.name] {
			types = append(types, t)
		}
	}
	m.writeUnions(&b, sortTypes(types, ProtoType), unionStyle{commented: true})
	return b.Bytes()
}
