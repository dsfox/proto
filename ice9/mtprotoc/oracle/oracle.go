// Package oracle reads the registers teamgram's mtprotoc generated, for the
// test that holds the schema to them. The schema itself was read back from
// the tree once (ice9/schema/README.md says how) and is edited by hand since.
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
