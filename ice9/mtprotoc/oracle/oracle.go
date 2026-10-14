// Package oracle reads what teamgram's mtprotoc generated and writes the
// schema our generator reads from it.
//
// teamgram's scheme.tl is not published, and it is not Telegram's history as
// Telegram ships it: the per-layer JSON files in the Android client carry
// one-layer blips (layer 161 holds other layers' ids for user, channel and
// message), typed ids (messageActionChatDeletePhoto is 95e3f807 there for 82
// layers and 95e3fbef in reality) and boundaries off by one. So the history of
// ids by layer, and the order of everything, come from the generated
// class_name_registers.go, and Telegram's JSON only says what a given id's
// definition is. Every place the two disagree is printed, so a change in
// either is seen rather than quietly absorbed.
package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// Registers is class_name_registers.go read back: the predicates in their
// order, and the id each one has at each layer.
type Registers struct {
	Order  []string
	Layers map[string]map[int]uint32
}

var (
	constantLine = regexp.MustCompile(`^\t(Predicate_\w+)\s+= "([^"]+)"`)
	entryLine    = regexp.MustCompile(`^\t(\S+): \{$`)
	layerLine    = regexp.MustCompile(`^\t\t(\d+):\s+(-?\d+),`)
)

// ReadRegisters parses class_name_registers.go.
func ReadRegisters(path string) (*Registers, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(data)
	constants := map[string]string{}
	r := &Registers{Layers: map[string]map[int]uint32{}}
	for _, line := range strings.Split(text, "\n") {
		if m := constantLine.FindStringSubmatch(line); m != nil {
			constants[m[1]] = m[2]
			r.Order = append(r.Order, m[2])
		}
	}
	start := strings.Index(text, "var clazzNameRegisters2")
	if start < 0 {
		return nil, fmt.Errorf("%s: no clazzNameRegisters2", path)
	}
	var current string
	for _, line := range strings.Split(text[start:], "\n") {
		if m := entryLine.FindStringSubmatch(line); m != nil {
			name, ok := constants[m[1]]
			if !ok {
				name = strings.Trim(m[1], `"`)
			}
			current = name
			r.Layers[current] = map[int]uint32{}
			continue
		}
		if m := layerLine.FindStringSubmatch(line); m != nil && current != "" {
			layer, _ := strconv.Atoi(m[1])
			id, _ := strconv.ParseInt(m[2], 10, 64)
			r.Layers[current][layer] = uint32(id)
			continue
		}
		if line == "}" {
			break
		}
	}
	for _, name := range r.Order {
		if _, ok := r.Layers[name]; !ok {
			return nil, fmt.Errorf("%s: predicate %s has no entry in the registers", path, name)
		}
	}
	return r, nil
}

// Definition is one line of Telegram's schema at one layer.
type Definition struct {
	Name   string // underscore form
	Method bool
	ID     uint32
	Params []schema.Param
	Result string
	Layer  int
}

// Telegram is Telegram's per-layer schema files, indexed.
type Telegram struct {
	ByID   map[uint32]Definition         // the definition at the highest layer that has the id
	ByName map[string]map[int]Definition // name -> layer -> definition
	Layers []int
}

type jsonSchema struct {
	Constructors []jsonEntry `json:"constructors"`
	Methods      []jsonEntry `json:"methods"`
}

type jsonEntry struct {
	ID        string `json:"id"`
	Predicate string `json:"predicate"`
	Method    string `json:"method"`
	Params    []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"params"`
	Type string `json:"type"`
}

// ReadTelegram reads every <layer>.json in the directory.
func ReadTelegram(dir string) (*Telegram, error) {
	t := &Telegram{ByID: map[uint32]Definition{}, ByName: map[string]map[int]Definition{}}
	names, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range names {
		base := strings.TrimSuffix(filepath.Base(path), ".json")
		layer, err := strconv.Atoi(base)
		if err != nil {
			continue
		}
		t.Layers = append(t.Layers, layer)
	}
	sort.Ints(t.Layers)
	for _, layer := range t.Layers {
		data, err := os.ReadFile(filepath.Join(dir, strconv.Itoa(layer)+".json"))
		if err != nil {
			return nil, err
		}
		var js jsonSchema
		if err := json.Unmarshal(data, &js); err != nil {
			return nil, fmt.Errorf("%d.json: %w", layer, err)
		}
		add := func(entry jsonEntry, method bool) error {
			name := entry.Predicate
			if method {
				name = entry.Method
			}
			signed, err := strconv.ParseInt(entry.ID, 10, 64)
			if err != nil {
				return fmt.Errorf("%d.json: %s: bad id %q", layer, name, entry.ID)
			}
			d := Definition{Name: strings.ReplaceAll(name, ".", "_"), Method: method, ID: uint32(signed), Result: entry.Type, Layer: layer}
			for _, p := range entry.Params {
				d.Params = append(d.Params, schema.Param{Name: p.Name, Type: p.Type})
			}
			// Later layers add flag bits under the same id; the newest
			// definition of an id holds every field it ever had.
			t.ByID[d.ID] = d
			if t.ByName[d.Name] == nil {
				t.ByName[d.Name] = map[int]Definition{}
			}
			t.ByName[d.Name][layer] = d
			return nil
		}
		for _, c := range js.Constructors {
			if err := add(c, false); err != nil {
				return nil, err
			}
		}
		for _, m := range js.Methods {
			if err := add(m, true); err != nil {
				return nil, err
			}
		}
	}
	return t, nil
}

// Report is what the oracle noticed while building the schema.
type Report struct {
	Lines []string
}

func (r *Report) Say(format string, args ...any) {
	r.Lines = append(r.Lines, fmt.Sprintf(format, args...))
}

// Build makes the schema from the registers and Telegram's definitions: the
// API entities up to the first MTProto service type, in the registers' order.
// Service and teamgram entities are returned separately, with no params,
// for mtproto.tl and teamgram.tl to be completed from other sources.
func Build(r *Registers, tg *Telegram, top int, firstService, firstTeamgram string) (api, service, teamgram *schema.Schema, report *Report) {
	report = &Report{}
	api = &schema.Schema{Layer: top}
	service = &schema.Schema{Layer: top}
	teamgram = &schema.Schema{Layer: top}
	into := api
	for _, name := range r.Order {
		switch name {
		case firstService:
			into = service
		case firstTeamgram:
			into = teamgram
		}
		layers := r.Layers[name]
		e := buildEntity(name, layers, top, tg, report, into != api)
		into.Entities = append(into.Entities, e)
	}
	report.Say("api %d, service %d, teamgram %d entities", len(api.Entities), len(service.Entities), len(teamgram.Entities))
	return api, service, teamgram, report
}

var hexSuffix = regexp.MustCompile(`^(.*?)([0-9A-F]{1,8})$`)

// buildEntity turns one predicate's layer map into versions.
func buildEntity(predicate string, layers map[int]uint32, top int, tg *Telegram, report *Report, noParams bool) *schema.Entity {
	name := predicate
	// A method with several ids is listed once per id, the id's uppercase
	// hex (without leading zeros) as a suffix; the TL name has no suffix.
	if m := hexSuffix.FindStringSubmatch(predicate); m != nil {
		for _, id := range layers {
			if strings.EqualFold(m[2], strconv.FormatUint(uint64(id), 16)) {
				name = m[1]
				break
			}
		}
	}
	sorted := make([]int, 0, len(layers))
	for l := range layers {
		sorted = append(sorted, l)
	}
	sort.Ints(sorted)
	e := &schema.Entity{Name: name}
	if d, ok := anyDefinition(tg, name, layers); ok {
		e.Method = d.Method
	}
	previous := -2
	for _, layer := range sorted {
		id := layers[layer]
		newVersion := layer == 0 || previous == -2 || layer != previous+1 || e.Versions[len(e.Versions)-1].ID != id
		if layer != 0 && previous != -2 && previous != 0 && layer != previous+1 {
			// Gone for a while and back: a removal, then a version again.
			e.Removals = append(e.Removals, previous+1)
		}
		if newVersion {
			v := schema.Version{Layer: layer, ID: id, Explicit: true}
			if !noParams {
				v.Params, v.Result = definitionFor(tg, name, id, layer, report)
			}
			e.Versions = append(e.Versions, v)
		}
		previous = layer
	}
	if previous != 0 && previous < top {
		e.Removals = append(e.Removals, previous+1)
	}
	return e
}

// anyDefinition is any of Telegram's definitions of the name, to learn
// whether it is a method.
func anyDefinition(tg *Telegram, name string, layers map[int]uint32) (Definition, bool) {
	for _, id := range layers {
		if d, ok := tg.ByID[id]; ok {
			return d, true
		}
	}
	for _, d := range tg.ByName[name] {
		return d, true
	}
	return Definition{}, false
}

// definitionFor is the params and result for an id: Telegram's definition of
// that id, else its definition of the name at that layer (the id is one
// Telegram's files got wrong), else nothing, said out loud.
func definitionFor(tg *Telegram, name string, id uint32, layer int, report *Report) ([]schema.Param, string) {
	if d, ok := tg.ByID[id]; ok {
		if d.Name != name {
			report.Say("%s: id %08x belongs to %s in Telegram's files", name, id, d.Name)
		}
		return d.Params, d.Result
	}
	byLayer := tg.ByName[name]
	if d, ok := byLayer[layer]; ok {
		report.Say("%s: id %08x at layer %d is not in Telegram's files; its definition there has id %08x, used", name, id, layer, d.ID)
		return d.Params, d.Result
	}
	// The nearest layer that has the name at all.
	best, found := Definition{}, false
	for l, d := range byLayer {
		if !found || abs(l-layer) < abs(best.Layer-layer) {
			best, found = d, true
			_ = l
		}
	}
	if found {
		report.Say("%s: id %08x at layer %d: Telegram has the name only at layer %d (id %08x), used", name, id, layer, best.Layer, best.ID)
		return best.Params, best.Result
	}
	report.Say("%s: id %08x at layer %d: no definition anywhere in Telegram's files", name, id, layer)
	return nil, ""
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Services is schema.tl.sync_service.proto read back: the services in order,
// each with its rpcs in order.
type Services struct {
	Order []string
	RPCs  map[string][]string
}

var (
	serviceLine = regexp.MustCompile(`^service (\w+) \{`)
	rpcLine     = regexp.MustCompile(`^  rpc (\w+)\(`)
)

// ReadServices parses schema.tl.sync_service.proto.
func ReadServices(path string) (*Services, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := &Services{RPCs: map[string][]string{}}
	var current string
	for _, line := range strings.Split(string(data), "\n") {
		if m := serviceLine.FindStringSubmatch(line); m != nil {
			current = m[1]
			s.Order = append(s.Order, current)
			continue
		}
		if m := rpcLine.FindStringSubmatch(line); m != nil && current != "" {
			s.RPCs[current] = append(s.RPCs[current], m[1])
		}
	}
	return s, nil
}

// WriteServices prints "service rpc" lines in file order.
func WriteServices(path string, s *Services) error {
	var b strings.Builder
	b.WriteString("# The RPC service each method is answered from, in the order of schema.tl.sync_service.proto.\n")
	b.WriteString("# Telegram's schema has no such grouping; it is teamgram's, read from their output once.\n")
	for _, svc := range s.Order {
		for _, rpc := range s.RPCs[svc] {
			fmt.Fprintf(&b, "%s %s\n", svc, rpc)
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
