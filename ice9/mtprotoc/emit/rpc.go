package emit

import (
	"bytes"
	"fmt"
)

// RPCClientRegisters is rpc_client_registers.go before gofmt: for every
// method the server answers, the gRPC route of its service and the type its
// reply is read into, in the services' order (services.txt).
func RPCClientRegisters(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(goHeader("2024-present"))
	b.WriteString(`
package mtproto

import (
	"reflect"
)

type newRPCReplyFunc func() interface{}

type RPCContextTuple struct {
	Method       string
	NewReplyFunc newRPCReplyFunc
}

var rpcContextRegisters = map[string]RPCContextTuple{
`)
	for _, svc := range m.Services {
		for _, e := range svc.Methods {
			name := m.Names[e]
			fmt.Fprintf(&b, "\t%q: RPCContextTuple{\"/mtproto.%s/%s\", func() interface{} { return new(%s) }},\n",
				GoName("TL_"+name), svc.Name, name, ReplyType(e.Versions[0].Result))
		}
	}
	b.WriteString(`}

func FindRPCContextTuple(t interface{}) *RPCContextTuple {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	m, ok := rpcContextRegisters[rt.Name()]
	if !ok {
		// log.Errorf("Can't find name: %s", rt.Name())
		return nil
	}
	return &m
}

func GetRPCContextRegisters() map[string]RPCContextTuple {
	return rpcContextRegisters
}
`)
	return b.Bytes()
}
