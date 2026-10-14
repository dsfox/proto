// Package schema is the multi-layer TL schema the generator reads: every
// constructor and method teamgram's scheme.tl holds, laid out as their
// input is, so that everything generated comes out in their order.
//
// The text is TL in sections, one per layer, newest first:
//
//	// LAYER 228
//	---types---
//	boolFalse#bc799737 = Bool;
//	...
//	---functions---
//	invokeAfterMsg#cb9f372d {X:Type} msg_id:long query:!X = X;
//	// LAYER 227 from 228
//	---types---
//	user#5d99adee flags:# ... = User;
//	- someConstructorAddedAt228
//	...
//	// LAYER 71
//	---functions---
//	account.registerDevice#637ea878 token_type:int token:string = Bool;
//
// A section lists what exists at its layer. One written "from" the section
// above lists only what differs there: a definition that replaces the one
// above, or "- name" for an entity the layer above has and this one has not.
// A section without "from" lists exactly its entities (the top one is the
// whole schema; the old layers 71 to 134 are a handful of methods each;
// layer 0, in mtproto.tl and teamgram.tl, is what is served when a client's
// layer is not listed). Within a section, lines are in the order teamgram's
// input has them at that layer, which is the order of everything generated:
// the order in which entities are first seen, newest layer first, is the
// order of the predicates; the order within a section is the order of the
// ids that layer retires.
//
// A line without "#id" has the id TL defines: the crc32 of the line.
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

// Version is one definition line of an entity.
type Version struct {
	Layer    int // the section's layer: the highest layer this definition holds at
	ID       uint32
	Explicit bool     // the id was written in the text; otherwise it is the crc32 of the line
	Generics []string // type parameters, "{X:Type}" in the text: the X of query:!X
	Params   []Param
	Result   string
}

// Entity is a constructor or a method: one name, its definitions over the
// layers. A method whose result type changed is one entity per id (see
// PredicateNames).
type Entity struct {
	Name     string // the TL name: "help.getConfig", "storage.fileUnknown", "req_pq_multi"
	Method   bool
	File     string     // the source the entity is first seen in
	Versions []*Version // one per definition line, in file order: highest layer first
	listing  map[int]uint32
}

// Line is one line of a section.
type Line struct {
	Entity  *Entity
	Version *Version // nil for a removal
}

// Section is one layer's listing.
type Section struct {
	File  string
	Layer int
	From  *Section // the section this one lists the changes to; nil when it lists everything
	Lines []*Line
}

// Schema is every section in file order, and every entity in the order it
// is first seen.
type Schema struct {
	Top      int
	Sections []*Section
	Entities []*Entity
}

// Predicate is the entity's name in the generated code: the TL name with the
// namespace dot as an underscore ("storage_fileUnknown", "help_getConfig").
func (e *Entity) Predicate() string { return strings.ReplaceAll(e.Name, ".", "_") }

// Listing is the id the entity has at each layer it is listed at: exactly
// class_name_registers.go's map for the entity.
func (e *Entity) Listing() map[int]uint32 { return e.listing }

// IDs is every id the entity ever had, newest first: the id listed at the
// highest layer, then the next one down. That is the order the generated
// files list a predicate's ids in.
func (e *Entity) IDs() []uint32 {
	newest := map[uint32]int{}
	for layer, id := range e.listing {
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

// Version is the newest definition of the entity that has this id.
func (e *Entity) Version(id uint32) *Version {
	for _, v := range e.Versions {
		if v.ID == id {
			return v
		}
	}
	return nil
}

// PredicateNames is the name of every entity in the generated code. A
// constructor keeps its name through every version; a method with several
// ids is a request type per id, each named with the id's uppercase hex,
// leading zeros dropped: account_verifyEmail32DA4CF, account_verifyEmailECBA39DB.
func PredicateNames(s *Schema) map[*Entity]string {
	entities := map[string]int{}
	for _, e := range s.Entities {
		if e.Method {
			entities[e.Name]++
		}
	}
	names := make(map[*Entity]string, len(s.Entities))
	for _, e := range s.Entities {
		name := e.Predicate()
		if e.Method && entities[e.Name] > 1 {
			name += strings.ToUpper(strconv.FormatUint(uint64(e.Versions[0].ID), 16))
		}
		names[e] = name
	}
	return names
}

// Source is one file of the schema.
type Source struct {
	Name   string
	Reader io.Reader
}

// ParseFiles reads the schema from its files, in this order.
func ParseFiles(paths ...string) (*Schema, error) {
	var sources []Source
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		sources = append(sources, Source{Name: path, Reader: f})
	}
	return Parse(sources)
}

// rawLine is a line as read, before entities exist.
type rawLine struct {
	source  string
	number  int
	name    string
	method  bool
	version *Version
}

type rawSection struct {
	file  string
	layer int
	from  int
	lines []rawLine
}

// Parse reads the sources in order.
//
// A constructor is one entity however many ids it had. A method is one entity
// too, unless its result type ever changed: then every id is an entity of its
// own, a request type per id in the generated code (account_verifyEmail
// answered Bool until layer 144 and account.EmailVerified from 145, so there
// are two; messages_sendMessage had nine ids and always answered Updates, so
// there is one). That is teamgram's rule, measured on every method they
// generated: the 23 names with several predicates all changed their result,
// none of the 139 others did.
func Parse(sources []Source) (*Schema, error) {
	var raws []*rawSection
	for _, src := range sources {
		sections, err := scan(src)
		if err != nil {
			return nil, err
		}
		raws = append(raws, sections...)
	}
	results := map[string]map[string]bool{}
	for _, sec := range raws {
		for _, l := range sec.lines {
			if l.version == nil || !l.method {
				continue
			}
			if results[l.name] == nil {
				results[l.name] = map[string]bool{}
			}
			results[l.name][l.version.Result] = true
		}
	}
	s := &Schema{}
	byKey := map[string]*Entity{}
	byName := map[string][]*Entity{}
	entityOf := func(l rawLine) (*Entity, error) {
		k := key(l.name, l.method)
		if l.version == nil {
			// A removal names whatever entity of that name is current; the
			// state walk below resolves it, so any entity of the name will do.
			if es := byName[k]; len(es) > 0 {
				return es[0], nil
			}
			return nil, fmt.Errorf("%s:%d: removal of %s, which was never defined", l.source, l.number, l.name)
		}
		if l.method && len(results[l.name]) > 1 {
			k += "#" + strconv.FormatUint(uint64(l.version.ID), 16)
		}
		e := byKey[k]
		if e == nil {
			e = &Entity{Name: l.name, Method: l.method, File: l.source, listing: map[int]uint32{}}
			s.Entities = append(s.Entities, e)
			byKey[k] = e
			byName[key(l.name, l.method)] = append(byName[key(l.name, l.method)], e)
		}
		e.Versions = append(e.Versions, l.version)
		return e, nil
	}
	type held struct {
		entity  *Entity
		version *Version
	}
	var states []map[string]held
	for i, raw := range raws {
		sec := &Section{File: raw.file, Layer: raw.layer}
		state := map[string]held{}
		if raw.from != 0 {
			base := -1
			for j := i - 1; j >= 0; j-- {
				if raws[j].file == raw.file && raws[j].layer == raw.from {
					base = j
					break
				}
			}
			if base < 0 {
				return nil, fmt.Errorf("%s: LAYER %d from %d: no such section above it", raw.file, raw.layer, raw.from)
			}
			sec.From = s.Sections[base]
			for k, h := range states[base] {
				state[k] = h
			}
		}
		for _, l := range raw.lines {
			e, err := entityOf(l)
			if err != nil {
				return nil, err
			}
			k := key(l.name, l.method)
			if l.version == nil {
				h, ok := state[k]
				if !ok {
					return nil, fmt.Errorf("%s:%d: %s is not in layer %d to be removed from layer %d", l.source, l.number, l.name, raw.from, raw.layer)
				}
				sec.Lines = append(sec.Lines, &Line{Entity: h.entity})
				delete(state, k)
				continue
			}
			// The same id with other params is a line too: Telegram has
			// dropped flag bits without a new id, and the union of fields
			// has them.
			if h, ok := state[k]; ok && sec.From != nil && Definition(l.name, h.version) == Definition(l.name, l.version) {
				return nil, fmt.Errorf("%s:%d: %s is the same at layer %d; the line changes nothing", l.source, l.number, l.name, raw.from)
			}
			state[k] = held{entity: e, version: l.version}
			sec.Lines = append(sec.Lines, &Line{Entity: e, Version: l.version})
		}
		for _, h := range state {
			h.entity.listing[raw.layer] = h.version.ID
		}
		if raw.layer > s.Top {
			s.Top = raw.layer
		}
		s.Sections = append(s.Sections, sec)
		states = append(states, state)
	}
	return s, nil
}

func key(name string, method bool) string {
	if method {
		return "f:" + name
	}
	return "t:" + name
}

// scan reads one source's sections.
func scan(src Source) ([]*rawSection, error) {
	var sections []*rawSection
	method := false
	scanner := bufio.NewScanner(src.Reader)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	number := 0
	for scanner.Scan() {
		number++
		text := strings.TrimSpace(scanner.Text())
		switch {
		case text == "":
			continue
		case strings.HasPrefix(text, "// LAYER "):
			fields := strings.Fields(strings.TrimPrefix(text, "// LAYER "))
			sec := &rawSection{file: src.Name}
			var err error
			switch {
			case len(fields) == 1:
				sec.layer, err = strconv.Atoi(fields[0])
			case len(fields) == 3 && fields[1] == "from":
				sec.layer, err = strconv.Atoi(fields[0])
				if err == nil {
					sec.from, err = strconv.Atoi(fields[2])
				}
			default:
				err = fmt.Errorf("not a layer")
			}
			if err != nil {
				return nil, fmt.Errorf("%s:%d: bad section header: %s", src.Name, number, text)
			}
			sections = append(sections, sec)
			method = false
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
		if len(sections) == 0 {
			return nil, fmt.Errorf("%s:%d: a line before the first // LAYER", src.Name, number)
		}
		sec := sections[len(sections)-1]
		if strings.HasPrefix(text, "- ") {
			sec.lines = append(sec.lines, rawLine{source: src.Name, number: number, name: strings.TrimSpace(text[2:]), method: method})
			continue
		}
		v, name, err := parseDefinition(text)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", src.Name, number, err)
		}
		v.Layer = sec.layer
		sec.lines = append(sec.lines, rawLine{source: src.Name, number: number, name: name, method: method, version: v})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sections, nil
}

// parseDefinition reads "name#id {X:Type} params = Result;". Without "#id"
// the id is computed, as TL defines it.
func parseDefinition(body string) (*Version, string, error) {
	v := &Version{}
	body = strings.TrimSuffix(strings.TrimSpace(body), ";")
	eq := strings.LastIndex(body, " = ")
	if eq < 0 {
		return nil, "", fmt.Errorf("no result type: %s", body)
	}
	v.Result = strings.TrimSpace(body[eq+3:])
	head := strings.Fields(body[:eq])
	if len(head) == 0 {
		return nil, "", fmt.Errorf("empty definition: %s", body)
	}
	name := head[0]
	if nameID := strings.SplitN(head[0], "#", 2); len(nameID) == 2 {
		id, err := strconv.ParseUint(nameID[1], 16, 32)
		if err != nil {
			return nil, "", fmt.Errorf("bad id: %s", head[0])
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
			return nil, "", fmt.Errorf("param without a type: %s in %s", p, body)
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
func Canonical(name string, v *Version) string {
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
func CRC32(name string, v *Version) uint32 { return crc32.ChecksumIEEE([]byte(Canonical(name, v))) }

// Write prints one file's sections in the text format.
func (s *Schema) Write(w io.Writer, file string) error {
	bw := bufio.NewWriter(w)
	for _, sec := range s.Sections {
		if sec.File != file {
			continue
		}
		if sec.From != nil {
			fmt.Fprintf(bw, "// LAYER %d from %d\n", sec.Layer, sec.From.Layer)
		} else {
			fmt.Fprintf(bw, "// LAYER %d\n", sec.Layer)
		}
		section := -1
		for _, l := range sec.Lines {
			want := 0
			if l.Entity.Method {
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
			if l.Version == nil {
				fmt.Fprintf(bw, "- %s\n", l.Entity.Name)
			} else {
				fmt.Fprintln(bw, Definition(l.Entity.Name, l.Version))
			}
		}
	}
	return bw.Flush()
}

// Definition is the TL line of one version.
func Definition(name string, v *Version) string {
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
