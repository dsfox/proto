package emit

import (
	"bytes"
	"fmt"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

const divider = "//--------------------------------------------------------------------------------------------\n"

// writeRequest writes a method's request message: its constructor and the
// union of its versions' fields, from 3.
func (m *Model) writeRequest(b *bytes.Buffer, e *schema.Entity) {
	fmt.Fprintf(b, "message TL_%s {\n  TLConstructor constructor = 1;\n", m.Names[e])
	for _, f := range unionFields(e.Versions, 3) {
		fmt.Fprintf(b, "  %s %s = %d;\n", f.Type, f.Name, f.Number)
	}
	b.WriteString("}\n")
}

// replyProtoType is the message a method's result is: Vector_<element> for
// a vector, else the type's message.
func replyProtoType(result string) string {
	if elem, ok := vectorElement(result); ok {
		return "Vector_" + vectorElementName(elem)
	}
	return ProtoType(result)
}

// vectorElementProto is the element type inside a Vector_ message.
func vectorElementProto(elem string) string {
	if scalar, ok := scalars[elem]; ok {
		return scalar
	}
	return ProtoType(elem)
}

// SyncServiceProto is schema.tl.sync_service.proto: a request message per
// API and teamgram method, the Vector_ messages their vector results are
// read into (in the order the results are first met), and the gRPC
// services of services.txt.
func SyncServiceProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(`
import "google/protobuf/wrappers.proto";
import "schema.tl.core_types.proto";
import "schema.tl.crc32.proto";
import "schema.tl.sync.proto";

option go_package = "github.com/teamgram/proto/mtproto";
`)
	methods := append(append([]*schema.Entity(nil), Methods(m.API)...), Methods(m.Teamgram)...)
	for _, e := range methods {
		b.WriteString("\n")
		m.writeRequest(&b, e)
	}
	b.WriteString("\n" + divider + "// Vector api result type\n")
	seen := map[string]bool{}
	for _, e := range methods {
		elem, ok := vectorElement(e.Versions[0].Result)
		if !ok {
			continue
		}
		name := "Vector_" + vectorElementName(elem)
		if seen[name] {
			continue
		}
		seen[name] = true
		fmt.Fprintf(&b, "message %s {\n  repeated %s datas = 1;\n}\n", name, vectorElementProto(elem))
	}
	b.WriteString("\n" + divider + "// rpc\n")
	for _, svc := range m.Services {
		fmt.Fprintf(&b, "\nservice %s {\n", svc.Name)
		for _, e := range svc.Methods {
			fmt.Fprintf(&b, "  rpc %s(TL_%s) returns (%s) {}\n", m.Names[e], m.Names[e], replyProtoType(e.Versions[0].Result))
		}
		b.WriteString("}\n")
	}
	return b.Bytes()
}

// HandshakeServiceProto is schema.tl.handshake_service.proto: the key
// exchange's requests and the service that answers them, each rpc under its
// TL line.
func HandshakeServiceProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(`
// import "google/protobuf/wrappers.proto";
import "schema.tl.crc32.proto";
import "schema.tl.handshake.proto";

option go_package = "github.com/teamgram/proto/mtproto";
`)
	var methods []*schema.Entity
	for _, e := range Methods(m.Service) {
		if handshakeMethods[e.Name] {
			methods = append(methods, e)
		}
	}
	for _, e := range methods {
		b.WriteString("\n")
		m.writeRequest(&b, e)
	}
	b.WriteString("\n" + divider + "// rpc\nservice RPCAuthKey {\n")
	for i, e := range methods {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "  // %s\n  rpc %s(TL_%s) returns (%s) {}\n", schema.Definition(e.Name, e.Versions[0]), m.Names[e], m.Names[e], replyProtoType(e.Versions[0].Result))
	}
	b.WriteString("}\n")
	return b.Bytes()
}

// TransportServiceProto is schema.tl.transport_service.proto: the
// transport's requests; nothing answers them over gRPC.
func TransportServiceProto(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader)
	b.WriteString(`
// import "google/protobuf/wrappers.proto";
import "schema.tl.crc32.proto";
// import "schema.tl.core_types.proto";

option go_package = "github.com/teamgram/proto/mtproto";
`)
	for _, e := range Methods(m.Service) {
		if !handshakeMethods[e.Name] {
			b.WriteString("\n")
			m.writeRequest(&b, e)
		}
	}
	return b.Bytes()
}
