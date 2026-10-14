package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// A tlEntry is one line of Telegram's per-layer schema file, as the Android
// client ships it (TMessagesProj_AppTests/tlscheme/N.json).
type tlEntry struct {
	ID        string `json:"id"`
	Predicate string `json:"predicate"`
	Method    string `json:"method"`
	Params    []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"params"`
	Type string `json:"type"`
}

type tlFile struct {
	Constructors []tlEntry `json:"constructors"`
	Methods      []tlEntry `json:"methods"`
}

// imported is one entity of the new layer.
type imported struct {
	name    string
	method  bool
	version *schema.Version
}

func (i *imported) key() string {
	if i.method {
		return "f:" + i.name
	}
	return "t:" + i.name
}

// readTLScheme reads a layer's file: its constructors, then its methods, in
// its order. "vector" is not a constructor of ours.
func readTLScheme(path string) ([]*imported, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file tlFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var out []*imported
	add := func(e tlEntry, method bool) error {
		name := e.Predicate
		if method {
			name = e.Method
		}
		if name == "vector" {
			return nil
		}
		signed, err := strconv.ParseInt(e.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("%s: %s: bad id %q", path, name, e.ID)
		}
		v := &schema.Version{ID: uint32(signed), Explicit: true, Result: e.Type}
		generic := e.Type == "X"
		for _, p := range e.Params {
			v.Params = append(v.Params, schema.Param{Name: p.Name, Type: p.Type})
			if strings.Contains(p.Type, "!X") {
				generic = true
			}
		}
		if generic {
			v.Generics = []string{"X"}
		}
		out = append(out, &imported{name: name, method: method, version: v})
		return nil
	}
	for _, c := range file.Constructors {
		if err := add(c, false); err != nil {
			return nil, err
		}
	}
	for _, m := range file.Methods {
		if err := add(m, true); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// runImport puts a new layer on top of layers.tl: the whole layer in the order
// the schema has, what the previous layer had and this one changes in a
// "from" section under it, and every new method into services.txt after the
// method before it.
func runImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	schemaDir := fs.String("schema", "ice9/schema", "the schema files")
	tlscheme := fs.String("tlscheme", "", "Telegram's schema file of the new layer, <N>.json")
	layerFlag := fs.Int("layer", 0, "the new layer; by default the file's name")
	fs.Parse(args)
	if *tlscheme == "" {
		return fmt.Errorf("-tlscheme is required")
	}
	layer := *layerFlag
	if layer == 0 {
		n, err := strconv.Atoi(strings.TrimSuffix(filepath.Base(*tlscheme), ".json"))
		if err != nil {
			return fmt.Errorf("the layer is not in the file's name; say -layer")
		}
		layer = n
	}
	layersPath := filepath.Join(*schemaDir, "layers.tl")
	s, err := schema.ParseFiles(layersPath, filepath.Join(*schemaDir, "mtproto.tl"), filepath.Join(*schemaDir, "teamgram.tl"))
	if err != nil {
		return err
	}
	top := s.Sections[0]
	if top.File != layersPath || top.From != nil {
		return fmt.Errorf("%s does not begin with a whole layer", layersPath)
	}
	if top.Layer != layer-1 {
		return fmt.Errorf("%s is at layer %d; the next layer is %d, not %d", layersPath, top.Layer, top.Layer+1, layer)
	}
	entries, err := readTLScheme(*tlscheme)
	if err != nil {
		return err
	}
	byKey := map[string]*imported{}
	byID := map[uint32]*imported{}
	for _, e := range entries {
		if _, dup := byKey[e.key()]; dup {
			return fmt.Errorf("%s: %s twice", *tlscheme, e.name)
		}
		byKey[e.key()] = e
		byID[e.version.ID] = e
	}

	// The new top: the old order, each line kept, replaced or dropped; the
	// old line goes under it when the new layer changes it. An entity is
	// the same by name, or by id under a name of the file's own: teamgram
	// shortened two predicates, and the server code names them so.
	var newTop, changed []*imported
	placed := map[string]bool{}
	for _, l := range top.Lines {
		old := &imported{name: l.Entity.Name, method: l.Entity.Method, version: l.Version}
		e := byKey[old.key()]
		if e == nil {
			if same := byID[old.version.ID]; same != nil && same.method == old.method && !placed[same.key()] {
				placed[same.key()] = true
				e = &imported{name: old.name, method: old.method, version: same.version}
			}
		}
		if e == nil {
			changed = append(changed, old)
			continue
		}
		placed[old.key()] = true
		placed[e.key()] = true
		if schema.Definition(e.name, e.version) == schema.Definition(old.name, old.version) {
			newTop = append(newTop, old)
			continue
		}
		newTop = append(newTop, e)
		changed = append(changed, old)
	}
	// New entities go after the entry before them in the schema file.
	var added []*imported
	for i, e := range entries {
		if placed[e.key()] {
			continue
		}
		at := -1
		for j := i - 1; j >= 0; j-- {
			if idx := indexOf(newTop, entries[j].key()); idx >= 0 {
				at = idx
				break
			}
		}
		newTop = insertAfter(newTop, at, e)
		placed[e.key()] = true
		added = append(added, e)
	}
	overridden, err := overriddenIDs(filepath.Join(*schemaDir, "overrides.txt"))
	if err != nil {
		return err
	}
	for _, e := range newTop {
		if computed := schema.CRC32(e.name, e.version); computed != e.version.ID && !overridden[e.version.ID] {
			fmt.Printf("%s#%08x: the crc32 of the line is %08x (a typo in the schema file, or an override to record)\n", e.name, e.version.ID, computed)
		}
	}

	// The text: the two new sections, then the file from its second section on.
	text, err := os.ReadFile(layersPath)
	if err != nil {
		return err
	}
	rest := ""
	if at := strings.Index(string(text), "\n// LAYER "); at >= 0 {
		rest = string(text)[at+1:]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "// LAYER %d\n", layer)
	writeLines(&b, newTop, nil)
	fmt.Fprintf(&b, "// LAYER %d from %d\n", layer-1, layer)
	writeLines(&b, changed, added)
	b.WriteString(rest)
	if err := os.WriteFile(layersPath, []byte(b.String()), 0o644); err != nil {
		return err
	}
	if _, err := schema.ParseFiles(layersPath); err != nil {
		return fmt.Errorf("the new %s does not read back: %w", layersPath, err)
	}

	// New methods, and methods whose result changed, into services.txt.
	if err := placeMethods(filepath.Join(*schemaDir, "services.txt"), newTop, added, s); err != nil {
		return err
	}
	fmt.Printf("layer %d: %d entities, %d changed or gone at %d, %d new\n", layer, len(newTop), len(changed), layer-1, len(added))
	return nil
}

func indexOf(list []*imported, key string) int {
	for i, e := range list {
		if e.key() == key {
			return i
		}
	}
	return -1
}

func insertAfter(list []*imported, at int, e *imported) []*imported {
	if at < 0 {
		// Nothing of its kind before it: first of the constructors, or first
		// of the methods.
		at = -1
		if e.method {
			for i, x := range list {
				if x.method {
					at = i - 1
					break
				}
			}
			if at == -1 && len(list) > 0 && !list[len(list)-1].method {
				at = len(list) - 1
			}
		}
	}
	out := make([]*imported, 0, len(list)+1)
	out = append(out, list[:at+1]...)
	out = append(out, e)
	return append(out, list[at+1:]...)
}

// writeLines writes definitions, then removals, with the section markers.
func writeLines(b *strings.Builder, definitions, removals []*imported) {
	section := -1
	marker := func(method bool) {
		want := 0
		if method {
			want = 1
		}
		if want != section {
			if want == 0 {
				b.WriteString("---types---\n")
			} else {
				b.WriteString("---functions---\n")
			}
			section = want
		}
	}
	for _, e := range definitions {
		marker(e.method)
		b.WriteString(schema.Definition(e.name, e.version) + "\n")
	}
	for _, e := range removals {
		marker(e.method)
		b.WriteString("- " + e.name + "\n")
	}
}

// placeMethods puts each new method into services.txt after the method
// before it in the schema, and renames the predicates of a method whose
// result type changed (each id is a request type of its own then).
func placeMethods(path string, newTop, added []*imported, old *schema.Schema) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	find := func(predicate string) int {
		for i, l := range lines {
			if fields := strings.Fields(l); len(fields) == 2 && fields[1] == predicate {
				return i
			}
		}
		return -1
	}
	hex := func(v *schema.Version) string { return strings.ToUpper(strconv.FormatUint(uint64(v.ID), 16)) }
	oldNames := schema.PredicateNames(old)
	// A method whose result changed: the old predicate is renamed, the new
	// one listed after it.
	for _, e := range newTop {
		if !e.method {
			continue
		}
		for _, oe := range old.Entities {
			if oe.Name != e.name || !oe.Method || oe.Listing()[old.Top] == 0 {
				continue
			}
			oldVersion := oe.Version(oe.Listing()[old.Top])
			if oldVersion.Result == e.version.Result || oldVersion.ID == e.version.ID {
				continue
			}
			wasPlain := oldNames[oe] == oe.Predicate()
			oldName := oe.Predicate() + hex(oldVersion)
			newName := oe.Predicate() + hex(e.version)
			if i := find(oldNames[oe]); i >= 0 {
				fields := strings.Fields(lines[i])
				if wasPlain {
					lines[i] = fields[0] + " " + oldName
				}
				lines = append(lines[:i+1], append([]string{fields[0] + " " + newName}, lines[i+1:]...)...)
				fmt.Printf("services.txt: %s answered %s, now %s: %s and %s under %s\n", oe.Predicate(), oldVersion.Result, e.version.Result, oldName, newName, fields[0])
			}
		}
	}
	// A new method goes after the nearest served method of its namespace
	// before it, else after the nearest served method at all.
	namespace := func(name string) string {
		if dot := strings.Index(name, "."); dot >= 0 {
			return name[:dot]
		}
		return ""
	}
	for _, e := range added {
		if !e.method {
			continue
		}
		at := indexOf(newTop, e.key())
		served := func(j int) int {
			if j < 0 || j >= len(newTop) || !newTop[j].method {
				return -1
			}
			name := strings.ReplaceAll(newTop[j].name, ".", "_")
			if i := find(name); i >= 0 {
				return i
			}
			return find(name + hex(newTop[j].version))
		}
		placedAt := -1
		// The nearest served method of the namespace, before or after; else
		// the nearest served method before.
		for d := 1; d < len(newTop) && placedAt < 0; d++ {
			for _, j := range []int{at - d, at + d} {
				if j >= 0 && j < len(newTop) && namespace(newTop[j].name) == namespace(e.name) && newTop[j].method {
					if i := served(j); i >= 0 {
						placedAt = i
						break
					}
				}
			}
		}
		for j := at - 1; j >= 0 && placedAt < 0; j-- {
			placedAt = served(j)
		}
		predicate := strings.ReplaceAll(e.name, ".", "_")
		if placedAt < 0 {
			fmt.Printf("services.txt: %s has no method before it that is served; add it by hand\n", predicate)
			continue
		}
		service := strings.Fields(lines[placedAt])[0]
		lines = append(lines[:placedAt+1], append([]string{service + " " + predicate}, lines[placedAt+1:]...)...)
		fmt.Printf("services.txt: %s -> %s (after %s)\n", predicate, service, strings.Fields(lines[placedAt])[1])
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// overriddenIDs reads the ids overrides.txt accounts for.
func overriddenIDs(path string) (map[uint32]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ids := map[uint32]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.HasPrefix(line, "#") {
			continue
		}
		id, err := strconv.ParseUint(fields[1], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: bad id in %q", path, line)
		}
		ids[uint32(id)] = true
	}
	return ids, nil
}
