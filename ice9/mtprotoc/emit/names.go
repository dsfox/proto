package emit

import "strings"

// GoName is the Go identifier protoc-gen-go makes of a protobuf name: an
// interior underscore followed by a lower-case letter is dropped and the
// letter capitalised (TL_help_getConfig is TLHelpGetConfig; TL_client_DH_inner_data
// is TLClient_DHInnerData). The generated Go code names types that way, and
// the id registers sort by it.
func GoName(name string) string {
	var b strings.Builder
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '.' && i+1 < len(name) && isLower(name[i+1]):
		case c == '.':
			b.WriteByte('_')
		case c == '_' && (i == 0 || name[i-1] == '.'):
			b.WriteByte('X')
		case c == '_' && i+1 < len(name) && isLower(name[i+1]):
		case isDigit(c):
			b.WriteByte(c)
		default:
			if isLower(c) {
				c -= 'a' - 'A'
			}
			b.WriteByte(c)
			for ; i+1 < len(name) && isLower(name[i+1]); i++ {
				b.WriteByte(name[i+1])
			}
		}
	}
	return b.String()
}

func isLower(c byte) bool { return 'a' <= c && c <= 'z' }
func isDigit(c byte) bool { return '0' <= c && c <= '9' }

// ProtoType is the protobuf message of a TL type: the namespace dot as an
// underscore (help.AppConfig is help_AppConfig).
func ProtoType(tl string) string { return strings.ReplaceAll(tl, ".", "_") }

// scalars are the TL scalars as protobuf writes them.
var scalars = map[string]string{
	"int":    "int32",
	"long":   "int64",
	"double": "double",
	"string": "string",
	"bytes":  "bytes",
	"int128": "bytes",
	"int256": "bytes",
}

// vectorElement is the type inside "Vector<...>" or "vector<...>", and
// whether the type was a vector at all.
func vectorElement(tl string) (string, bool) {
	for _, prefix := range []string{"Vector<", "vector<"} {
		if strings.HasPrefix(tl, prefix) && strings.HasSuffix(tl, ">") {
			return tl[len(prefix) : len(tl)-1], true
		}
	}
	return "", false
}

// ReplyType is the Go type a method's result is read into: the type's Go
// name, or Vector_<element> for a vector, with scalars as Int, Long, String.
func ReplyType(result string) string {
	if elem, ok := vectorElement(result); ok {
		return "Vector_" + vectorElementName(elem)
	}
	return GoName(ProtoType(result))
}

// vectorElementName is the element as the Vector_ message names it.
func vectorElementName(elem string) string {
	switch elem {
	case "int":
		return "Int"
	case "long":
		return "Long"
	case "string":
		return "String"
	}
	return GoName(ProtoType(elem))
}
