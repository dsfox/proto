package emit

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// A codecField is one param of a definition line as the codec writes it:
// its field in the message (name with any suffix), its place among the
// message's params (the number of the mN/listN/cN/lN/vN variables), and
// which flag bit, if any, says whether it is there.
type codecField struct {
	param  schema.Param
	name   string // the field's name in the message
	goName string
	index  int
	flags  string // "flags" or "flags2" for an optional param
	bit    int
	inner  string // the TL type without its flags prefix
	ptype  string // the protobuf type
}

func (f *codecField) optional() bool { return f.flags != "" }
func (f *codecField) isFlags() bool  { return f.param.Type == "#" }
func (f *codecField) isBool() bool   { return f.inner == "true" }
func (f *codecField) isQuery() bool  { return f.inner == "!X" }

// isScalar is a TL type written by one EncodeBuf call.
func (f *codecField) isScalar() bool {
	switch f.inner {
	case "int", "long", "double", "string", "bytes", "int128", "int256":
		return true
	}
	return false
}

func (f *codecField) isVector() bool {
	_, ok := vectorElement(f.inner)
	return ok
}

// bare is a vector<...> with a lower-case v: written without the vector
// constructor in front.
func (f *codecField) bare() bool { return strings.HasPrefix(f.inner, "vector<") }

func (f *codecField) element() string {
	elem, _ := vectorElement(f.inner)
	return elem
}

// elementScalar is the EncodeBuf method name for a vector of scalars, or ""
// for a vector of objects.
func (f *codecField) elementScalar() string {
	switch f.element() {
	case "int":
		return "VectorInt"
	case "long":
		return "VectorLong"
	case "string":
		return "VectorString"
	case "bytes":
		return "VectorBytes"
	}
	return ""
}

// elementGo is the Go type of a vector's element object.
func (f *codecField) elementGo() string {
	elem := f.element()
	if isLower(elem[0]) && !strings.Contains(elem, ".") {
		return GoName("TL_" + elem)
	}
	return GoName(ProtoType(elem))
}

// scalarCall is the EncodeBuf / DecodeBuf method for a scalar.
func (f *codecField) scalarCall() string {
	switch f.inner {
	case "int":
		return "Int"
	case "long":
		return "Long"
	case "double":
		return "Double"
	case "string":
		return "String"
	case "bytes":
		return "StringBytes"
	case "int128", "int256":
		return "Bytes"
	}
	return ""
}

// wrapper is the wrapperspb type of an optional scalar.
func (f *codecField) wrapper() string {
	switch f.inner {
	case "int":
		return "Int32Value"
	case "long":
		return "Int64Value"
	case "double":
		return "DoubleValue"
	case "string":
		return "StringValue"
	}
	return ""
}

// goType is the Go type of the field, as the setters take it.
func goType(ptype string) string {
	if strings.HasPrefix(ptype, "repeated ") {
		elem := goType(strings.TrimPrefix(ptype, "repeated "))
		return "[]" + elem
	}
	switch ptype {
	case "int32", "int64", "string", "bool":
		return ptype
	case "double":
		return "float64"
	case "bytes":
		return "[]byte"
	}
	if strings.HasPrefix(ptype, "google.protobuf.") {
		return "*wrapperspb." + strings.TrimPrefix(ptype, "google.protobuf.")
	}
	return "*" + GoName(ptype)
}

// slots is the union of params over lines by name, each name once, "flags"
// included: the index of a param is its place here, whatever its type.
type slots struct {
	fields map[string]*Field // by base name and protobuf type
	order  []string          // names in order
	index  map[string]int
}

func slotKey(p schema.Param) string { return p.Name }

func newSlots(lines []*schema.Version, first int) *slots {
	s := &slots{fields: map[string]*Field{}, index: map[string]int{}}
	for _, f := range unionFields(lines, 3) {
		s.fields[f.base+"|"+f.Type] = f
	}
	next := first
	for _, v := range lines {
		for _, p := range v.Params {
			k := slotKey(p)
			if _, seen := s.index[k]; seen {
				continue
			}
			s.index[k] = next
			s.order = append(s.order, k)
			next++
		}
	}
	return s
}

// fieldsOf is the version's params as codec fields, named by these slots
// and numbered by `numbering` (the same slots for a constructor; a method
// numbers each version's params on their own).
func (s *slots) fieldsOf(v *schema.Version, numbering *slots) []*codecField {
	var out []*codecField
	for _, p := range v.Params {
		f := &codecField{param: p, index: numbering.index[slotKey(p)], inner: p.Type}
		if p.Type == "#" {
			f.name = p.Name
		} else {
			ptype, _ := fieldType(p.Type)
			f.ptype = ptype
			f.name = s.fields[fieldName(p.Name)+"|"+ptype].Name
			if strings.HasPrefix(p.Type, "flags") {
				if q := strings.Index(p.Type, "?"); q >= 0 {
					f.flags = p.Type[:strings.Index(p.Type, ".")]
					fmt.Sscanf(p.Type[strings.Index(p.Type, ".")+1:q], "%d", &f.bit)
					f.inner = p.Type[q+1:]
				}
			}
		}
		f.goName = GoName(f.name)
		out = append(out, f)
	}
	return out
}

// codecWriter writes lines at an indent.
type codecWriter struct {
	b      *bytes.Buffer
	indent int
}

func (w *codecWriter) line(format string, args ...any) {
	if format == "" {
		w.b.WriteString("\n")
		return
	}
	w.b.WriteString(strings.Repeat("\t", w.indent))
	fmt.Fprintf(w.b, format, args...)
	w.b.WriteString("\n")
}

func (w *codecWriter) in()  { w.indent++ }
func (w *codecWriter) out() { w.indent-- }

// flagChecks writes the "flags |= 1 << N" checks of one flags field, in
// param order; a param that is not that field's gives a blank line. A string
// sharing its bit with a later vector of objects fills the vector in, so the
// bit means both.
func (w *codecWriter) flagChecks(fields []*codecField, which string, getter func(*codecField) string, sibling bool) {
	for i, f := range fields {
		if f.flags != which {
			w.line("")
			continue
		}
		if f.isBool() {
			w.line("if %s == true {", getter(f))
			w.line("\tflags |= 1 << %d", f.bit)
			w.line("}")
			continue
		}
		w.line("if %s != nil {", getter(f))
		w.line("\tflags |= 1 << %d", f.bit)
		if sibling && f.inner == "string" {
			for _, g := range fields[i+1:] {
				if g.flags == which && g.bit == f.bit && g.isVector() && g.elementScalar() == "" {
					w.line("\tif %s == nil {", getter(g))
					w.line("\t\tm.Set%s([]*%s{})", g.goName, g.elementGo())
					w.line("\t}")
				}
			}
		}
		w.line("}")
	}
}

// encodeField writes what puts one field on the wire.
func (w *codecWriter) encodeField(f *codecField, getter string) {
	switch {
	case f.isFlags():
		w.line("// set flags")
		w.line("var %s = get%s()", f.name, strings.ToUpper(f.name[:1])+f.name[1:])
		w.line("x.UInt(%s)", f.name)
	case f.isBool():
	case f.isQuery():
		w.line("x.Bytes(%s)", getter)
	case f.isVector():
		if scalar := f.elementScalar(); scalar != "" {
			if f.optional() {
				w.line("if %s != nil {", getter)
				w.line("\tx.%s(%s)", scalar, getter)
				w.line("}")
			} else {
				w.line("")
				w.line("x.%s(%s)", scalar, getter)
				w.line("")
			}
			return
		}
		if f.bare() {
			w.line("// x.Int(int32(CRC32_vector))")
			w.line("list%d := %s", f.index, getter)
			w.line("x.Int(int32(len(list%d)))", f.index)
			w.line("for _, v := range list%d {", f.index)
			w.line("\tv.Encode(x, layer)")
			w.line("}")
			w.line("")
			return
		}
		if f.optional() {
			w.line("if %s != nil {", getter)
			w.in()
		} else {
			w.line("")
		}
		w.line("x.Int(int32(CRC32_vector))")
		w.line("{")
		w.line("\tvar (")
		w.line("\t\tsz     = int32(0)")
		w.line("\t\toffset = x.GetOffset()")
		w.line("\t)")
		w.line("\tlist%d := %s", f.index, getter)
		w.line("\tx.Int(int32(len(list%d)))", f.index)
		w.line("\tfor _, v := range list%d {", f.index)
		w.line("\t\terr := v.Encode(x, layer)")
		w.line("\t\tif err == nil {")
		w.line("\t\t\tsz += 1")
		w.line("\t\t}")
		w.line("\t}")
		w.line("\tx.IntOffset(offset, sz)")
		w.line("}")
		if f.optional() {
			w.out()
			w.line("}")
		} else {
			w.line("")
		}
	case f.isScalar():
		if !f.optional() {
			w.line("x.%s(%s)", f.scalarCall(), getter)
			return
		}
		w.line("if %s != nil {", getter)
		if f.wrapper() == "" {
			w.line("\tx.%s(%s)", f.scalarCall(), getter)
		} else {
			w.line("\tx.%s(%s.Value)", f.scalarCall(), getter)
		}
		w.line("}")
		w.line("")
	default: // an object
		if !f.optional() {
			w.line("%s.Encode(x, layer)", getter)
			return
		}
		w.line("if %s != nil {", getter)
		w.line("\t%s.Encode(x, layer)", getter)
		w.line("}")
		w.line("")
	}
}

// decodeField writes what reads one field; set is how a value is stored
// ("m.SetX(%s)" for a constructor, "m.X = %s" for a method), and the
// vector length error is written the method's way when `method`.
func (w *codecWriter) decodeField(f *codecField, set string, method bool) {
	store := func(value string) string { return fmt.Sprintf(set, value) }
	switch {
	case f.isFlags():
		if method {
			w.line("%s := dBuf.UInt()", f.name)
		} else {
			w.line("var %s = dBuf.UInt()", f.name)
		}
		w.line("_ = %s", f.name)
	case f.isBool():
		w.line("if (%s & (1 << %d)) != 0 {", f.flags, f.bit)
		w.line("\t" + store("true"))
		w.line("}")
	case f.isQuery():
		w.line(store("dBuf.Bytes(dBuf.GetSize() - dBuf.GetOffset())"))
	case f.isVector():
		if scalar := f.elementScalar(); scalar != "" {
			if f.optional() {
				w.line("if (%s & (1 << %d)) != 0 {", f.flags, f.bit)
				w.line("\t" + store("dBuf."+scalar+"()"))
				w.line("}")
			} else {
				w.line("")
				w.line(store("dBuf." + scalar + "()"))
				w.line("")
			}
			return
		}
		n := f.index
		if f.optional() {
			w.line("if (%s & (1 << %d)) != 0 {", f.flags, f.bit)
			w.in()
		}
		if f.bare() {
			w.line("// c%d := dBuf.Int()", n)
			w.line("// if c%d != int32(CRC32_vector) {", n)
			w.line("//     dBuf.err = fmt.Errorf(\"invalid CRC32_vector, c%%d: %%d\", %d, c%d)", n, n)
			w.line("//     return dBuf.err")
			w.line("// }")
		} else {
			w.line("c%d := dBuf.Int()", n)
			w.line("if c%d != int32(CRC32_vector) {", n)
			w.line("\t// dBuf.err = fmt.Errorf(\"invalid CRC32_vector, c%%d: %%d\", %d, c%d)", n, n)
			w.line("\treturn dBuf.GetError()")
			w.line("}")
		}
		w.line("l%d := dBuf.Int()", n)
		w.line("if l%d < 0 || l%d > maxVectorLen {", n, n)
		if method {
			w.line("\tdBuf.err = fmt.Errorf(\"invalid vector length: %%d\", l%d)", n)
			w.line("\treturn dBuf.err")
		} else {
			w.line("\treturn fmt.Errorf(\"invalid vector length: %%d\", l%d)", n)
		}
		w.line("}")
		w.line("v%d := make([]*%s, l%d)", n, f.elementGo(), n)
		w.line("for i := int32(0); i < l%d; i++ {", n)
		w.line("\tv%d[i] = &%s{}", n, f.elementGo())
		w.line("\tv%d[i].Decode(dBuf)", n)
		w.line("}")
		w.line(store(fmt.Sprintf("v%d", n)))
		if f.optional() {
			w.out()
			w.line("}")
		} else {
			w.line("")
		}
	case f.isScalar():
		read := "dBuf." + f.scalarCall() + "()"
		switch f.inner {
		case "int128":
			read = "dBuf.Bytes(16)"
		case "int256":
			read = "dBuf.Bytes(32)"
		}
		if !f.optional() {
			w.line(store(read))
			return
		}
		w.line("if (%s & (1 << %d)) != 0 {", f.flags, f.bit)
		if f.wrapper() == "" {
			w.line("\t" + store(read))
		} else {
			w.line("\t" + store(fmt.Sprintf("&wrapperspb.%s{Value: %s}", f.wrapper(), read)))
		}
		w.line("}")
		w.line("")
	default: // an object
		n := f.index
		if f.optional() {
			w.line("if (%s & (1 << %d)) != 0 {", f.flags, f.bit)
			w.in()
		} else {
			w.line("")
		}
		w.line("m%d := &%s{}", n, GoName(f.ptype))
		w.line("m%d.Decode(dBuf)", n)
		w.line(store(fmt.Sprintf("m%d", n)))
		if f.optional() {
			w.out()
			w.line("}")
		} else {
			w.line("")
		}
	}
}

// hasFlags says whether a version has a given flags field.
func hasFlags(fields []*codecField, name string) bool {
	for _, f := range fields {
		if f.isFlags() && f.name == name {
			return true
		}
	}
	return false
}

// writeType writes one type's block: the type's Encode and Decode, a To_X
// per constructor, then each constructor's MakeTLX, To_type, setters,
// GetPredicateName, Encode, CalcByteSize and Decode.
func (m *Model) writeType(w *codecWriter, t *protoType) {
	typeName := GoName(ProtoType(t.name))
	sl := newSlots(t.lines, 0)
	w.line("///////////////////////////////////////////////////////////////////////////////")
	w.line("// %s <--", typeName)
	for _, e := range t.constructors {
		w.line("//  + TL_%s", GoName(m.Names[e]))
	}
	w.line("//")
	w.line("")
	w.line("func (m *%s) Encode(x *EncodeBuf, layer int32) error {", typeName)
	w.line("\tpredicateName := m.PredicateName")
	w.line("\tif predicateName == \"\" {")
	w.line("\t\tif n, ok := clazzIdNameRegisters2[int32(m.Constructor)]; ok {")
	w.line("\t\t\tpredicateName = n")
	w.line("\t\t}")
	w.line("\t}")
	w.line("")
	w.b.WriteString(codecEncodePrologue[t.name])
	w.line("\tswitch predicateName {")
	for _, e := range t.constructors {
		w.line("\tcase Predicate_%s:", m.Names[e])
		if body := codecCaseOverride[t.name][m.Names[e]]; body != "" {
			w.b.WriteString(body)
			continue
		}
		w.line("\t\tt := m.To_%s()", GoName(m.Names[e]))
		w.line("\t\terr := t.Encode(x, layer)")
		w.line("\t\tif err != nil {")
		w.line("\t\t\treturn err")
		w.line("\t\t}")
	}
	if extra := ice9EncodeCase[t.name]; extra != "" {
		w.b.WriteString(extra)
	}
	w.line("")
	w.line("\tdefault:")
	w.line("\t\treturn fmt.Errorf(\"invalid predicate error: %%s\", m.PredicateName)")
	w.line("\t}")
	w.line("")
	w.b.WriteString(codecEncodeEpilogue[t.name])
	w.line("\treturn nil")
	w.line("}")
	w.line("")
	w.line("func (m *%s) CalcByteSize(layer int32) int {", typeName)
	w.line("\treturn 0")
	w.line("}")
	w.line("")
	w.line("func (m *%s) Decode(dBuf *DecodeBuf) error {", typeName)
	w.line("\tm.Constructor = TLConstructor(dBuf.Int())")
	w.line("\tswitch uint32(m.Constructor) {")
	for _, e := range t.constructors {
		for _, id := range e.IDs() {
			w.line("\tcase 0x%x:", id)
			w.line("\t\tm2 := MakeTL%s(m)", GoName(m.Names[e]))
			w.line("\t\tm2.Decode(dBuf)")
			if extra := ice9DecodeCase[t.name][id]; extra != "" {
				w.b.WriteString(extra)
			}
		}
	}
	w.line("")
	w.line("\tdefault:")
	w.line("\t\treturn fmt.Errorf(\"invalid constructorId: 0x%%x\", uint32(m.Constructor))")
	w.line("\t}")
	w.line("\treturn dBuf.GetError()")
	w.line("}")
	for _, e := range t.constructors {
		name := GoName(m.Names[e])
		w.line("")
		w.line("// To_%s", name)
		w.line("func (m *%s) To_%s() *TL%s {", typeName, name, name)
		w.line("\tm.PredicateName = Predicate_%s", m.Names[e])
		w.line("\treturn &TL%s{", name)
		w.line("\t\tData2: m,")
		w.line("\t}")
		w.line("}")
	}
	for _, e := range t.constructors {
		m.writeConstructor(w, t, typeName, e, sl)
	}
}

// writeConstructor writes one constructor's functions.
func (m *Model) writeConstructor(w *codecWriter, t *protoType, typeName string, e *schema.Entity, sl *slots) {
	name := GoName(m.Names[e])
	pred := m.Names[e]
	w.line("")
	w.line("// MakeTL%s", name)
	w.line("func MakeTL%s(data2 *%s) *TL%s {", name, typeName, name)
	w.line("\tif data2 == nil {")
	w.line("\t\treturn &TL%s{Data2: &%s{", name, typeName)
	w.line("\t\t\tPredicateName: Predicate_%s,", pred)
	w.line("\t\t}}")
	w.line("\t} else {")
	w.line("\t\tdata2.PredicateName = Predicate_%s", pred)
	w.line("\t\treturn &TL%s{Data2: data2}", name)
	w.line("\t}")
	w.line("}")
	w.line("")
	w.line("func (m *TL%s) To_%s() *%s {", name, typeName, typeName)
	w.line("\tm.Data2.PredicateName = Predicate_%s", pred)
	w.line("\treturn m.Data2")
	w.line("}")
	// Setters and getters: every param of every version, each once; the
	// "flags" param is a comment above the first setter after it.
	seen := map[string]bool{}
	gap := "\n"
	for _, v := range e.Versions {
		for _, f := range sl.fieldsOf(v, sl) {
			if f.isFlags() {
				if !seen["#"+f.name] {
					seen["#"+f.name] = true
					w.b.WriteString(gap + "// // flags\n")
					gap = ""
				}
				continue
			}
			if seen[f.name] {
				continue
			}
			seen[f.name] = true
			typ := goType(f.ptype)
			set := fmt.Sprintf("func (m *TL%s) Set%s(v %s) { m.Data2.%s = v }", name, f.goName, typ, f.goName)
			get := fmt.Sprintf("func (m *TL%s) Get%s() %s { return m.Data2.%s }", name, f.goName, typ, f.goName)
			w.b.WriteString(gap)
			gap = "\n"
			w.b.WriteString(splitLong(set, fmt.Sprintf("m.Data2.%s = v", f.goName)))
			w.b.WriteString(splitLong(get, fmt.Sprintf("return m.Data2.%s", f.goName)))
		}
	}
	w.line("")
	w.line("func (m *TL%s) GetPredicateName() string {", name)
	w.line("\treturn Predicate_%s", pred)
	w.line("}")
	w.line("")
	w.line("func (m *TL%s) Encode(x *EncodeBuf, layer int32) error {", name)
	w.line("\tclazzId := GetClazzID(Predicate_%s, int(layer))", pred)
	w.line("\tswitch uint32(clazzId) {")
	getter := func(f *codecField) string { return "m.Get" + f.goName + "()" }
	for _, id := range e.IDs() {
		v := e.Version(id)
		fields := sl.fieldsOf(v, sl)
		w.line("\tcase 0x%x:", id)
		w.line("\t\tx.UInt(0x%x)", id)
		w.line("")
		w.indent = 2
		// The closures of the flags fields, then a blank line before the
		// fields.
		var closures []string
		for _, which := range []string{"flags", "flags2"} {
			if hasFlags(fields, which) {
				closures = append(closures, which)
			}
		}
		for i, which := range closures {
			w.line("// set %s", which)
			w.line("var get%s = func() uint32 {", strings.ToUpper(which[:1])+which[1:])
			w.line("\tvar flags uint32 = 0")
			w.line("")
			w.in()
			w.flagChecks(fields, which, getter, true)
			w.out()
			w.line("")
			w.line("\treturn flags")
			w.line("}")
			if i == len(closures)-1 {
				w.line("")
			}
		}
		for _, f := range fields {
			w.encodeField(f, getter(f))
		}
		w.line("")
		w.line("return nil")
		w.indent = 0
	}
	w.line("")
	w.line("\tdefault:")
	w.line("\t\treturn fmt.Errorf(\"not found clazzId by (%%s, %%d)\", Predicate_%s, layer)", pred)
	w.line("\t}")
	w.line("}")
	w.line("")
	w.line("func (m *TL%s) CalcByteSize(layer int32) int {", name)
	w.line("\treturn 0")
	w.line("}")
	w.line("")
	w.line("func (m *TL%s) Decode(dBuf *DecodeBuf) error {", name)
	w.line("\tswitch uint32(m.Data2.Constructor) {")
	for _, id := range e.IDs() {
		v := e.Version(id)
		w.line("\tcase 0x%x:", id)
		w.indent = 2
		for _, f := range sl.fieldsOf(v, sl) {
			w.decodeField(f, "m.Set"+f.goName+"(%s)", false)
		}
		w.line("return dBuf.GetError()")
		w.indent = 0
	}
	w.line("")
	w.line("\tdefault:")
	w.line("\t\treturn fmt.Errorf(\"invalid constructor: %%x\", uint32(m.Data2.Constructor))")
	w.line("\t}")
	w.line("}")
}

// splitLong writes a one-line function, or its three-line form when the
// line would be longer than 104 characters.
func splitLong(oneLine, body string) string {
	if len(oneLine) <= 104 {
		return oneLine + "\n"
	}
	head := oneLine[:strings.Index(oneLine, " { ")]
	return head + " {\n\t" + body + "\n}\n"
}

// writeMethod writes a request type's Encode, CalcByteSize and Decode.
func (m *Model) writeMethod(w *codecWriter, e *schema.Entity) {
	name := GoName("TL_" + m.Names[e])
	sl := newSlots(e.Versions, 1)
	w.line("// %s", name)
	w.line("// /////////////////////////////////////////////////////////////////////////////")
	w.line("func (m *%s) Encode(x *EncodeBuf, layer int32) error {", name)
	w.line("\tswitch uint32(m.Constructor) {")
	getter := func(f *codecField) string { return "m.Get" + f.goName + "()" }
	for _, id := range e.IDs() {
		v := e.Version(id)
		fields := sl.fieldsOf(v, newSlots([]*schema.Version{v}, 1))
		w.line("\tcase 0x%x:", id)
		w.line("\t\tx.UInt(0x%x)", id)
		w.line("")
		w.indent = 2
		if hasFlags(fields, "flags") {
			w.line("// set flags")
			w.line("var flags uint32 = 0")
			w.line("")
			w.flagChecks(fields, "flags", getter, false)
			w.line("")
			w.line("x.UInt(flags)")
			w.line("")
			w.line("// flags Debug by @benqi")
		} else {
			w.line("// no flags")
			w.line("")
		}
		for _, f := range fields {
			if !f.isFlags() {
				w.encodeField(f, getter(f))
			}
		}
		w.line("")
		w.indent = 0
	}
	w.line("\tdefault:")
	w.line("\t\t// log.Errorf(\"\")")
	w.line("\t}")
	w.line("")
	w.line("\treturn nil")
	w.line("}")
	w.line("")
	w.line("func (m *%s) CalcByteSize(layer int32) int {", name)
	w.line("\treturn 0")
	w.line("}")
	w.line("")
	w.line("func (m *%s) Decode(dBuf *DecodeBuf) error {", name)
	w.line("\tswitch uint32(m.Constructor) {")
	for _, id := range e.IDs() {
		v := e.Version(id)
		fields := sl.fieldsOf(v, newSlots([]*schema.Version{v}, 1))
		w.line("\tcase 0x%x:", id)
		w.line("")
		w.indent = 2
		if hasFlags(fields, "flags") {
			w.line("flags := dBuf.UInt()")
			w.line("_ = flags")
			w.line("")
			w.line("// flags Debug by @benqi")
		} else {
			w.line("// not has flags")
			w.line("")
		}
		for _, f := range fields {
			if f.isFlags() {
				continue
			}
			w.decodeField(f, "m."+f.goName+" = %s", true)
		}
		w.line("return dBuf.GetError()")
		w.indent = 0
	}
	w.line("")
	w.line("\tdefault:")
	w.line("\t\t// log.Errorf(\"\")")
	w.line("\t}")
	w.line("\treturn dBuf.GetError()")
	w.line("}")
}

// writeVector writes a Vector_ reply type's codec.
func (m *Model) writeVector(w *codecWriter, name, elem string) {
	w.line("// %s", name)
	w.line("// /////////////////////////////////////////////////////////////////////////////")
	w.line("func (m *%s) Encode(x *EncodeBuf, layer int32) error {", name)
	f := &codecField{inner: "Vector<" + elem + ">"}
	if scalar := f.elementScalar(); scalar != "" {
		w.line("\tx.%s(m.Datas)", scalar)
		w.line("")
		w.line("\treturn nil")
		w.line("}")
		w.line("")
		w.line("func (m *%s) Decode(dBuf *DecodeBuf) error {", name)
		w.line("\tm.Datas = dBuf.%s()", scalar)
		w.line("")
		w.line("\treturn dBuf.GetError()")
		w.line("}")
	} else {
		w.line("\tx.Int(int32(CRC32_vector))")
		w.line("\t{")
		w.line("\t\tvar (")
		w.line("\t\t\tsz     = int32(0)")
		w.line("\t\t\toffset = x.GetOffset()")
		w.line("\t\t)")
		w.line("\t\tx.Int(int32(len(m.Datas)))")
		w.line("\t\tfor _, v := range m.Datas {")
		w.line("\t\t\terr := v.Encode(x, layer)")
		w.line("\t\t\tif err == nil {")
		w.line("\t\t\t\tsz += 1")
		w.line("\t\t\t}")
		w.line("\t\t}")
		w.line("\t\tx.IntOffset(offset, sz)")
		w.line("\t}")
		w.line("")
		w.line("\treturn nil")
		w.line("}")
		w.line("")
		w.line("func (m *%s) Decode(dBuf *DecodeBuf) error {", name)
		w.line("\tdBuf.Int() // TODO(@benqi): Check crc32 invalid")
		w.line("\tl1 := dBuf.Int()")
		w.line("\tif l1 < 0 || l1 > maxVectorLen {")
		w.line("\t\treturn fmt.Errorf(\"invalid vector length: %%d\", l1)")
		w.line("\t}")
		w.line("\tm.Datas = make([]*%s, l1)", f.elementGo())
		w.line("\tfor i := int32(0); i < l1; i++ {")
		w.line("\t\tm.Datas[i] = new(%s)", f.elementGo())
		w.line("\t\t(*m.Datas[i]).Decode(dBuf)")
		w.line("\t}")
		w.line("")
		w.line("\treturn dBuf.GetError()")
		w.line("}")
	}
	w.line("")
	w.line("func (m *%s) CalcByteSize(layer int32) int {", name)
	w.line("\treturn 0")
	w.line("}")
}

// Codec is codec_schema.tl.pb.go before gofmt: how every type, constructor
// and request goes on the wire and comes off it. The types come as the id
// registers order them (the handshake's, the transport's, every other, by
// Go name), FutureSalts and FutureSalt first as teamgram writes them; then
// the requests in schema order, the service's first; then the Vector_ reply
// types in the order their methods come.
func Codec(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(protoHeader[:strings.Index(protoHeader, "\nsyntax")])
	b.WriteString(`
// ConstructorList
// RequestList

package mtproto

import (
	"fmt"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

const maxVectorLen = 1000000

`)
	b.WriteString(futureSaltsCodec)
	w := &codecWriter{b: &b}
	var handshake, transport, rest []*protoType
	for _, t := range typesIn(m.types(), Constructors(m.Service)) {
		switch {
		case t.name == "FutureSalts" || t.name == "FutureSalt":
		case handshakeTypes[t.name]:
			handshake = append(handshake, t)
		default:
			transport = append(transport, t)
		}
	}
	rest = typesIn(m.types(), append(append([]*schema.Entity(nil), Constructors(m.API)...), Constructors(m.Teamgram)...))
	byGoName := func(name string) string { return GoName(ProtoType(name)) }
	for _, part := range [][]*protoType{handshake, transport, rest} {
		for _, t := range sortTypes(part, byGoName) {
			w.line("")
			m.writeType(w, t)
		}
	}
	w.line("")
	w.line("// ----------------------------------------------------------------------------------------------------------------")
	methods := append(append(append([]*schema.Entity(nil), Methods(m.Service)...), Methods(m.API)...), Methods(m.Teamgram)...)
	for i, e := range methods {
		if i > 0 {
			w.line("")
		}
		m.writeMethod(w, e)
	}
	w.line("")
	w.line("// ----------------------------------------------------------------------------------------------------------------")
	seen := map[string]bool{}
	first := true
	for _, e := range append(append([]*schema.Entity(nil), Methods(m.API)...), Methods(m.Teamgram)...) {
		elem, ok := vectorElement(e.Versions[0].Result)
		if !ok {
			continue
		}
		name := "Vector_" + vectorElementName(elem)
		if seen[name] {
			continue
		}
		seen[name] = true
		if !first {
			w.line("")
		}
		first = false
		m.writeVector(w, name, elem)
	}
	return b.Bytes()
}
