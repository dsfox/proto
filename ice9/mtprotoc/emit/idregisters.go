package emit

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// handshakeTypes are the types of the key exchange, which have a .proto of
// their own (schema.tl.handshake.proto); the other service types are the
// transport's.
var handshakeTypes = map[string]bool{
	"ResPQ":                       true,
	"P_Q_inner_data":              true,
	"BindAuthKeyInner":            true,
	"Server_DH_Params":            true,
	"Server_DH_inner_data":        true,
	"Client_DH_Inner_Data":        true,
	"Set_client_DH_params_answer": true,
	"DestroyAuthKeyRes":           true,
}

// handshakeMethods are the key exchange's methods (schema.tl.handshake_service.proto).
var handshakeMethods = map[string]bool{
	"req_pq":               true,
	"req_pq_multi":         true,
	"req_DH_params":        true,
	"set_client_DH_params": true,
	"destroy_auth_key":     true,
}

// typeGroup is one result type with its constructors in schema order.
type typeGroup struct {
	name         string // the TL type
	constructors []*schema.Entity
}

// typesOf groups constructors by their result type, in the order the types
// are first seen; constructors keep their order within a type.
func typesOf(constructors []*schema.Entity) []*typeGroup {
	var groups []*typeGroup
	byName := map[string]*typeGroup{}
	for _, e := range constructors {
		name := e.Versions[0].Result
		g := byName[name]
		if g == nil {
			g = &typeGroup{name: name}
			byName[name] = g
			groups = append(groups, g)
		}
		g.constructors = append(g.constructors, e)
	}
	return groups
}

// sortedByGoName orders types as protoc-gen-go names them, byte by byte.
func sortedByGoName(groups []*typeGroup) []*typeGroup {
	sorted := append([]*typeGroup(nil), groups...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return GoName(ProtoType(sorted[i].name)) < GoName(ProtoType(sorted[j].name))
	})
	return sorted
}

// ClassIDRegisters is class_id_registers.tl.go before gofmt: a factory per
// id. Constructors come by type - the handshake's types, the transport's,
// then every other - each part sorted by the type's Go name, a type's
// constructors in schema order, a constructor's ids newest first; methods
// come in schema order, the service's first.
func ClassIDRegisters(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(goHeader("2024-present"))
	b.WriteString(`
// ConstructorList
// RequestList

package mtproto

var clazzIdRegisters2 = map[int32]func() TLObject{
	// parsed_manually_types
	1538843921: func() TLObject { // CRC32_message2
		return &TLMessage2{}
	},
	1945237724: func() TLObject { // CRC32_msg_container
		return &TLMsgContainer{}
	},
	530561358: func() TLObject { // CRC32_msg_copy
		return &TLMsgCopy{}
	},
	812830625: func() TLObject { // CRC32_gzip_packed
		return &TLGzipPacked{}
	},
	-212046591: func() TLObject { // CRC32_rpc_result
		return &TLRpcResult{}
	},

	// Constructor
`)
	var handshake, transport, rest []*typeGroup
	for _, g := range typesOf(Constructors(m.Service)) {
		if handshakeTypes[g.name] {
			handshake = append(handshake, g)
		} else {
			transport = append(transport, g)
		}
	}
	rest = typesOf(append(append([]*schema.Entity(nil), Constructors(m.API)...), Constructors(m.Teamgram)...))
	for _, part := range [][]*typeGroup{handshake, transport, rest} {
		for _, g := range sortedByGoName(part) {
			for _, e := range g.constructors {
				for _, id := range e.IDs() {
					fmt.Fprintf(&b, "\t%d: func() TLObject { // 0x%x\n\t\to := Make%s(nil)\n\t\to.Data2.Constructor = %d\n\t\treturn o\n\t},\n",
						int32(id), id, GoName("TL_"+m.Names[e]), int32(id))
				}
			}
		}
	}
	b.WriteString("\n\t// Method\n")
	for _, group := range [][]*schema.Entity{Methods(m.Service), Methods(m.API), Methods(m.Teamgram)} {
		for _, e := range group {
			for _, id := range e.IDs() {
				fmt.Fprintf(&b, "\t%d: func() TLObject { // 0x%x\n\t\treturn &%s{\n\t\t\tConstructor: %d,\n\t\t}\n\t},\n",
					int32(id), id, GoName("TL_"+m.Names[e]), int32(id))
			}
		}
	}
	b.WriteString(`}

func NewTLObjectByClassID(classId int32) TLObject {
	f, ok := clazzIdRegisters2[classId]
	if !ok {
		return nil
	}
	return f()
}

func CheckClassID(classId int32) (ok bool) {
	_, ok = clazzIdRegisters2[classId]
	return
}
`)
	return b.Bytes()
}
