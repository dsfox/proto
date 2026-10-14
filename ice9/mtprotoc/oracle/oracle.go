// Package oracle reads what teamgram's mtprotoc generated, for the tests
// that hold the schema to it. The schema itself was read back from the tree
// once (ice9/schema/README.md says how) and is edited by hand since.
package oracle

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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
