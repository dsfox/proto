package emit

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

// goHeader opens a generated .go file; the year is the file's own.
func goHeader(years string) string {
	return fmt.Sprintf(`/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright (c) %s,  Teamgram Authors.
 *  All rights reserved.
 *
 * Author: Benqi (wubenqi@gmail.com)
 */
`, years)
}

// hexComment is the id as the generated files comment it: hex, with "0x"
// when the schema line had no id and the generator computed it.
func hexComment(e *schema.Entity, id uint32) string {
	if v := e.Version(id); v != nil && !v.Explicit {
		return fmt.Sprintf("0x%x", id)
	}
	return fmt.Sprintf("%x", id)
}

// ClassNameRegisters is class_name_registers.go before gofmt: a constant per
// predicate, the id of each predicate at each layer it is listed at (newest
// layer first, layer 0 last), the name of each id, and GetClazzID - with
// ice9's fallback for a client ahead of the schema.
func ClassNameRegisters(m *Model) []byte {
	var b bytes.Buffer
	b.WriteString(goHeader("2025-present"))
	b.WriteString("\npackage mtproto\n\nconst (\n")
	for _, e := range m.All.Entities {
		fmt.Fprintf(&b, "\tPredicate_%s = %q\n", m.Names[e], m.Names[e])
	}
	b.WriteString(")\n\nvar clazzNameRegisters2 = map[string]map[int]int32{\n")
	for _, e := range m.All.Entities {
		fmt.Fprintf(&b, "\tPredicate_%s: {\n", m.Names[e])
		listed := e.Listing()
		layers := make([]int, 0, len(listed))
		for layer := range listed {
			layers = append(layers, layer)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(layers)))
		for _, layer := range layers {
			fmt.Fprintf(&b, "\t\t%d: %d, // %s\n", layer, int32(listed[layer]), hexComment(e, listed[layer]))
		}
		b.WriteString("\n\t},\n")
	}
	// Each id names its predicate, in the order the ids are met reading the
	// schema: the whole top layer, then each lower layer's lines.
	b.WriteString("}\n\nvar clazzIdNameRegisters2 = map[int32]string{\n")
	named := map[uint32]bool{}
	for _, sec := range m.All.Sections {
		for _, l := range sec.Lines {
			if l.Version == nil || named[l.Version.ID] {
				continue
			}
			named[l.Version.ID] = true
			fmt.Fprintf(&b, "\t%d: Predicate_%s, // %s\n", int32(l.Version.ID), m.Names[l.Entity], hexComment(l.Entity, l.Version.ID))
		}
	}
	b.WriteString(`
}

func GetClazzID(clazzName string, layer int) int32 {
	if m, ok := clazzNameRegisters2[clazzName]; ok {
		m2, ok2 := m[layer]
		if ok2 {
			return m2
		}
		m2, ok2 = m[0]
		if ok2 {
			return m2
		}
		// ice9: a client ahead of this schema is served what we have (clazz_id_ahead.go)
		if top, ok3 := clazzNameNewest[clazzName]; ok3 && layer > top.layer {
			return top.id
		}
	}
	return 0
}
`)
	return b.Bytes()
}
