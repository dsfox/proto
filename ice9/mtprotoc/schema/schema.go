// Package schema is the multi-layer TL schema the generator reads: every
// constructor and method teamgram's scheme.tl accumulated, in its order, with
// the layer each version holds from.
//
// The text format is TL, one line per version, the layer in a trailing comment:
//
//	// LAYER 228
//	---types---
//	user#3ff6ecb0 flags:# self:flags.10?true ... = User; // 139
//	user#5d99adee flags:# ... = User; // 145
//	- messageUserVote // 170
//	---functions---
//	help.getConfig#c4f9186b = Config; // 139
//
// A version holds from its layer until the next version of the same entity or
// a removal line. Layer 0 is the entry served when a client's layer is not
// listed (the MTProto service types, the six core predicates). The order of
// first lines is the order of everything generated: it is teamgram's order,
// not Telegram's, and it cannot be derived from anything but their output.
package schema

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Param is one field of a TL definition, its type as TL writes it:
// "int", "flags.3?true", "Vector<User>", "!X", "#".
type Param struct {
	Name string
	Type string
}

// Version is one definition of an entity, holding from Layer on.
type Version struct {
	Layer    int
	ID       uint32
	Explicit bool     // the id was written in the text; otherwise it is the crc32 of the line, as TL defines it
	Generics []string // type parameters, "{X:Type}" in the text: the X of query:!X
	Params   []Param
	Result   string
}

// Entity is a constructor or a method with its history.
type Entity struct {
	Name     string // the TL name: "help.getConfig", "storage.fileUnknown", "req_pq_multi"
	Method   bool
	Versions []Version // ascending by Layer
	Removals []int     // layers from which the entity no longer exists, until a later version brings it back
}

// Predicate is the entity's name in the generated code: the TL name with the
// namespace dot as an underscore ("storage_fileUnknown", "help_getConfig").
func (e *Entity) Predicate() string { return strings.ReplaceAll(e.Name, ".", "_") }

// Schema is every entity in order, and the top layer the schema knows.
type Schema struct {
	Layer    int
	Entities []*Entity
}

// PredicateNames is the name of every entity in the generated code. A
// constructor keeps its name through every version; a method with several
// ids is a request type per id, each named with the id's uppercase hex,
// leading zeros dropped: account_verifyEmail32DA4CF, account_verifyEmailECBA39DB.
func PredicateNames(s *Schema) map[*Entity]string {
	versions := map[string]int{}
	for _, e := range s.Entities {
		if e.Method {
			versions[e.Name]++
		}
	}
	names := make(map[*Entity]string, len(s.Entities))
	for _, e := range s.Entities {
		name := e.Predicate()
		if e.Method && versions[e.Name] > 1 {
			name += strings.ToUpper(strconv.FormatUint(uint64(e.Last().ID), 16))
		}
		names[e] = name
	}
	return names
}

// Last is the newest version.
func (e *Entity) Last() *Version { return &e.Versions[len(e.Versions)-1] }

// Listing is the entity as the registers list it: every layer from each
// version's layer to the next event (a newer version or a removal) or the
// schema's top layer, with its id; layer 0 when the entity has a layer-0
// version. This is exactly class_name_registers.go's map for the entity.
func (e *Entity) Listing(top int) map[int]uint32 {
	listed := map[int]uint32{}
	events := make([]event, 0, len(e.Versions)+len(e.Removals))
	for _, v := range e.Versions {
		if v.Layer == 0 {
			listed[0] = v.ID
			continue
		}
		events = append(events, event{layer: v.Layer, id: v.ID, version: true})
	}
	for _, r := range e.Removals {
		events = append(events, event{layer: r})
	}
	sort.Slice(events, func(i, j int) bool { return events[i].layer < events[j].layer })
	for i, ev := range events {
		if !ev.version {
			continue
		}
		end := top
		if i+1 < len(events) {
			end = events[i+1].layer - 1
		}
		for l := ev.layer; l <= end; l++ {
			listed[l] = ev.id
		}
	}
	return listed
}

type event struct {
	layer   int
	id      uint32
	version bool
}

// IDAt is the constructor id a client at this layer is written, or 0 when the
// entity does not exist at that layer.
func (e *Entity) IDAt(layer int, top int) uint32 { return e.Listing(top)[layer] }

// Find is the entity with this TL name; a method and a constructor may share
// one, and a method with several ids appears once per id.
func (s *Schema) Find(name string, method bool) []*Entity {
	var found []*Entity
	for _, e := range s.Entities {
		if e.Name == name && e.Method == method {
			found = append(found, e)
		}
	}
	return found
}

// ParseFile reads the schema text.
func ParseFile(path string) (*Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// item is one line of the text: a definition, or a removal (no Version).
type item struct {
	line    int
	name    string
	method  bool
	layer   int
	version *Version
}

// Parse reads the schema text; several files parsed in a row are joined by
// Append, in that order.
//
// A constructor is one entity however many ids it had. A method is one entity
// too, unless its result type ever changed: then every version is an entity of
// its own, a request type per id in the generated code (account_verifyEmail
// answered Bool until layer 144 and account.EmailVerified from 145, so there
// are two; messages_sendMessage had nine ids and always answered Updates, so
// there is one). That is teamgram's rule, measured on every method they
// generated: the 23 names with several predicates all changed their result,
// none of the 139 others did.
func Parse(r io.Reader) (*Schema, error) {
	items, layer, err := scan(r)
	if err != nil {
		return nil, err
	}
	s := &Schema{Layer: layer}
	results := map[string]map[string]bool{}
	for _, it := range items {
		if it.version == nil || !it.method {
			continue
		}
		if results[it.name] == nil {
			results[it.name] = map[string]bool{}
		}
		results[it.name][it.version.Result] = true
	}
	latest := map[string]*Entity{}
	for _, it := range items {
		k := key(it.name, it.method)
		if it.version == nil {
			e := latest[k]
			if e == nil {
				return nil, fmt.Errorf("line %d: removal of %s, which was never defined", it.line, it.name)
			}
			e.Removals = append(e.Removals, it.layer)
			continue
		}
		e := latest[k]
		if e == nil || (it.method && len(results[it.name]) > 1) {
			e = &Entity{Name: it.name, Method: it.method}
			s.Entities = append(s.Entities, e)
			latest[k] = e
		}
		e.Versions = append(e.Versions, *it.version)
	}
	for _, e := range s.Entities {
		sort.SliceStable(e.Versions, func(i, j int) bool { return e.Versions[i].Layer < e.Versions[j].Layer })
		sort.Ints(e.Removals)
	}
	return s, nil
}

// scan reads the lines; the schema's top layer comes back with them.
func scan(r io.Reader) ([]item, int, error) {
	var items []item
	top := 0
	method := false
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		switch {
		case text == "":
			continue
		case strings.HasPrefix(text, "// LAYER "):
			n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(text, "// LAYER ")))
			if err != nil {
				return nil, 0, fmt.Errorf("line %d: bad layer: %s", line, text)
			}
			top = n
			continue
		case strings.HasPrefix(text, "//"):
			continue
		case text == "---types---":
			method = false
			continue
		case text == "---functions---":
			method = true
			continue
		}
		body, layer, err := splitLayer(text)
		if err != nil {
			return nil, 0, fmt.Errorf("line %d: %w", line, err)
		}
		if strings.HasPrefix(body, "- ") {
			items = append(items, item{line: line, name: strings.TrimSpace(body[2:]), method: method, layer: layer})
			continue
		}
		v, name, err := parseDefinition(body)
		if err != nil {
			return nil, 0, fmt.Errorf("line %d: %w", line, err)
		}
		v.Layer = layer
		items = append(items, item{line: line, name: name, method: method, layer: layer, version: &v})
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}
	return items, top, nil
}

func key(name string, method bool) string {
	if method {
		return "f:" + name
	}
	return "t:" + name
}

// splitLayer takes the "// 139" off the end of a line.
func splitLayer(text string) (string, int, error) {
	at := strings.LastIndex(text, "//")
	if at < 0 {
		return "", 0, fmt.Errorf("no layer comment: %s", text)
	}
	layer, err := strconv.Atoi(strings.TrimSpace(text[at+2:]))
	if err != nil {
		return "", 0, fmt.Errorf("bad layer comment: %s", text)
	}
	return strings.TrimSpace(text[:at]), layer, nil
}

// parseDefinition reads "name#id {X:Type} params = Result;". Without "#id"
// the id is computed, as TL defines it.
func parseDefinition(body string) (Version, string, error) {
	var v Version
	body = strings.TrimSuffix(strings.TrimSpace(body), ";")
	eq := strings.LastIndex(body, " = ")
	if eq < 0 {
		return v, "", fmt.Errorf("no result type: %s", body)
	}
	v.Result = strings.TrimSpace(body[eq+3:])
	head := strings.Fields(body[:eq])
	if len(head) == 0 {
		return v, "", fmt.Errorf("empty definition: %s", body)
	}
	name := head[0]
	if nameID := strings.SplitN(head[0], "#", 2); len(nameID) == 2 {
		id, err := strconv.ParseUint(nameID[1], 16, 32)
		if err != nil {
			return v, "", fmt.Errorf("bad id: %s", head[0])
		}
		name, v.ID, v.Explicit = nameID[0], uint32(id), true
	}
	for _, p := range head[1:] {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, ":Type}") {
			v.Generics = append(v.Generics, p[1:len(p)-len(":Type}")])
			continue
		}
		colon := strings.Index(p, ":")
		if colon < 0 {
			return v, "", fmt.Errorf("param without a type: %s in %s", p, body)
		}
		v.Params = append(v.Params, Param{Name: p[:colon], Type: p[colon+1:]})
	}
	if !v.Explicit {
		v.ID = CRC32(name, v)
	}
	return v, name, nil
}

// Canonical is the line TL's constructor id is the crc32 of: the definition
// without the id and the semicolon, flag-only ("flags.N?true") params left
// out, angle brackets as spaces, "bytes" as "string", generics unbraced.
func Canonical(name string, v Version) string {
	var b strings.Builder
	b.WriteString(name)
	for _, g := range v.Generics {
		b.WriteString(" " + g + ":Type")
	}
	for _, p := range v.Params {
		if strings.HasSuffix(p.Type, "?true") {
			continue
		}
		t := p.Type
		if t == "bytes" || strings.HasSuffix(t, "?bytes") {
			t = strings.TrimSuffix(t, "bytes") + "string"
		}
		b.WriteString(" " + p.Name + ":" + t)
	}
	b.WriteString(" = " + v.Result)
	line := strings.NewReplacer("<", " ", ">", " ").Replace(b.String())
	for strings.Contains(line, "  ") {
		line = strings.ReplaceAll(line, "  ", " ")
	}
	return strings.TrimSpace(line)
}

// CRC32 is the constructor id TL gives the definition.
func CRC32(name string, v Version) uint32 { return crc32.ChecksumIEEE([]byte(Canonical(name, v))) }

// Write prints the schema in the text format.
func Write(w io.Writer, s *Schema) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "// LAYER %d\n", s.Layer)
	section := -1
	for _, e := range s.Entities {
		want := 0
		if e.Method {
			want = 1
		}
		if want != section {
			if want == 0 {
				fmt.Fprintln(bw, "---types---")
			} else {
				fmt.Fprintln(bw, "---functions---")
			}
			section = want
		}
		// Versions and removals in layer order, so the history reads top down.
		removals := append([]int(nil), e.Removals...)
		for _, v := range e.Versions {
			for len(removals) > 0 && removals[0] < v.Layer {
				fmt.Fprintf(bw, "- %s // %d\n", e.Name, removals[0])
				removals = removals[1:]
			}
			fmt.Fprintf(bw, "%s // %d\n", Definition(e.Name, v), v.Layer)
		}
		for _, r := range removals {
			fmt.Fprintf(bw, "- %s // %d\n", e.Name, r)
		}
	}
	return bw.Flush()
}

// Definition is the TL line of one version, without the layer.
func Definition(name string, v Version) string {
	var b strings.Builder
	b.WriteString(name)
	if v.Explicit {
		fmt.Fprintf(&b, "#%x", v.ID)
	}
	for _, g := range v.Generics {
		fmt.Fprintf(&b, " {%s:Type}", g)
	}
	for _, p := range v.Params {
		fmt.Fprintf(&b, " %s:%s", p.Name, p.Type)
	}
	fmt.Fprintf(&b, " = %s;", v.Result)
	return b.String()
}

// Append joins another schema's entities after this one's, in their order.
func (s *Schema) Append(other *Schema) {
	s.Entities = append(s.Entities, other.Entities...)
	if other.Layer > s.Layer {
		s.Layer = other.Layer
	}
}
