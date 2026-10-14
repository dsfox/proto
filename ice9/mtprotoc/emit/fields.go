package emit

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// Field is one field of a generated message.
type Field struct {
	Name   string // as written: the TL name, with a type suffix when the name has several types
	Type   string // the protobuf type, "repeated X" for a vector
	Number int
	base   string // the TL name
	tag    string // the suffix the type gets: INT32, FLAGSTRING, VECTORPEER
}

// fieldType is how a TL type is written in protobuf, and the suffix it gives
// a field name when the name has several types: an optional scalar is a
// wrapper, a flag bit a bool, a vector repeated, any other type its message.
func fieldType(tl string) (ptype, tag string) {
	flag := false
	if strings.HasPrefix(tl, "flags") {
		if q := strings.Index(tl, "?"); q >= 0 {
			tl = tl[q+1:]
			flag = true
		}
	}
	switch tl {
	case "true":
		ptype, tag = "bool", "BOOLEAN"
	case "int":
		ptype, tag = "int32", "INT32"
		if flag {
			ptype = "google.protobuf.Int32Value"
		}
	case "long":
		ptype, tag = "int64", "INT64"
		if flag {
			ptype = "google.protobuf.Int64Value"
		}
	case "double":
		ptype, tag = "double", "FLOAT64"
		if flag {
			ptype = "google.protobuf.DoubleValue"
		}
	case "string":
		ptype, tag = "string", "STRING"
		if flag {
			ptype = "google.protobuf.StringValue"
		}
	case "bytes", "int128", "int256":
		ptype, tag = "bytes", strings.ToUpper(tl)
	case "!X":
		ptype, tag = "bytes", "X"
	default:
		if elem, ok := vectorElement(tl); ok {
			etype, etag := fieldType(elem)
			if _, scalar := scalars[elem]; !scalar && elem != "" && isLower(elem[0]) && !strings.Contains(elem, ".") {
				// A bare vector of a constructor: vector<future_salt>.
				etype = "TL_" + elem
			}
			ptype, tag = "repeated "+etype, "VECTOR"+etag
		} else {
			ptype, tag = ProtoType(tl), strings.ToUpper(ProtoType(tl))
		}
	}
	if flag {
		tag = "FLAG" + tag
	}
	return ptype, tag
}

// fieldName is the param's name as a field: "size" is "size2", which keeps
// it off the generated Go type's Size method.
func fieldName(param string) string {
	if param == "size" {
		return "size2"
	}
	return param
}

// unionFields is the message a type (or a method) gets: the fields of every
// definition line in order (a line repeating an id with other params counts),
// each name once; a name met again with another protobuf type becomes a
// field of its own, and both carry their type as a suffix. Numbers are given
// in order from `first` and never change.
func unionFields(lines []*schema.Version, first int) []*Field {
	var fields []*Field
	next := first
	for _, v := range lines {
		for _, p := range v.Params {
			if p.Type == "#" {
				continue
			}
			base := fieldName(p.Name)
			ptype, tag := fieldType(p.Type)
			var same []*Field
			merged := false
			for _, f := range fields {
				if f.base != base {
					continue
				}
				if f.Type == ptype {
					merged = true
					break
				}
				same = append(same, f)
			}
			if merged {
				continue
			}
			name := base
			if len(same) > 0 {
				for _, f := range same {
					f.Name = f.base + "_" + f.tag
				}
				name = base + "_" + tag
			}
			fields = append(fields, &Field{Name: name, Type: ptype, Number: next, base: base, tag: tag})
			next++
		}
	}
	return fields
}

// Extra is a field teamgram's generator adds after the schema's: one of a
// name's types under the plain name, for the server code that reads it.
type Extra struct {
	Message string
	Field   Field
}

// loadExtras reads "message type name number" lines.
func loadExtras(path string) (map[string][]Field, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	extras := map[string][]Field{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var msg, ptype, name string
		var number int
		if _, err := fmt.Sscanf(line, "%s %s %s %d", &msg, &ptype, &name, &number); err != nil {
			return nil, fmt.Errorf("%s: an extra field is message, type, name and number: %s", path, line)
		}
		extras[msg] = append(extras[msg], Field{Name: name, Type: ptype, Number: number})
	}
	return extras, scanner.Err()
}

// distinctIDs keeps the first line of each id: a definition repeated at a
// lower layer with the same id adds nothing.
func distinctIDs(lines []*schema.Version) []*schema.Version {
	seen := map[uint32]bool{}
	var out []*schema.Version
	for _, v := range lines {
		if !seen[v.ID] {
			seen[v.ID] = true
			out = append(out, v)
		}
	}
	return out
}
