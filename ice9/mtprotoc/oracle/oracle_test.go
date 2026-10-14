package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

const tree = "../../../mtproto"

var schemaFiles = []string{"layers.tl", "mtproto.tl", "teamgram.tl"}

// The schema files, read in order, are what the registers hold: the same
// predicates in the same order, each listed at the same layers with the same
// ids. The text is also what Write prints, so an edit by hand cannot drift
// into a form the writer would not make.
func TestTheSchemaIsTheRegisters(t *testing.T) {
	registers, err := ReadRegisters(filepath.Join(tree, "class_name_registers.go"))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, name := range schemaFiles {
		paths = append(paths, filepath.Join("..", "..", "schema", name))
	}
	s, err := schema.ParseFiles(paths...)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var printed bytes.Buffer
		if err := s.Write(&printed, path); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(printed.Bytes(), text) {
			t.Errorf("%s: not what Write prints; the first differing line: %s", path, firstDifference(string(text), printed.String()))
		}
	}
	if len(s.Entities) != len(registers.Order) {
		t.Fatalf("schema has %d entities, the registers %d predicates", len(s.Entities), len(registers.Order))
	}
	names := schema.PredicateNames(s)
	for i, e := range s.Entities {
		want := registers.Order[i]
		if names[e] != want {
			t.Fatalf("entity %d is %s, the registers have %s", i, names[e], want)
		}
		listed := e.Listing()
		theirs := registers.Layers[want]
		if len(listed) != len(theirs) {
			t.Errorf("%s: listed at %d layers, the registers at %d", want, len(listed), len(theirs))
			continue
		}
		for layer, id := range theirs {
			if listed[layer] != id {
				t.Errorf("%s at layer %d: %08x, the registers have %08x", want, layer, listed[layer], id)
				break
			}
		}
	}
}

func firstDifference(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range al {
		if i >= len(bl) {
			return "line " + strconv.Itoa(i+1) + ": " + al[i] + " (missing in the printed text)"
		}
		if al[i] != bl[i] {
			return "line " + strconv.Itoa(i+1) + ":\n  file:    " + al[i] + "\n  printed: " + bl[i]
		}
	}
	if len(bl) > len(al) {
		return "printed has extra lines from " + strconv.Itoa(len(al)+1)
	}
	return "none"
}
