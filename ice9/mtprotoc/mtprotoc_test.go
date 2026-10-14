package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/teamgram/proto/ice9/mtprotoc/emit"
)

const tree = "../../mtproto"

// Every generated file is the tree's, byte for byte.
func TestTheGeneratorReproducesTheTree(t *testing.T) {
	m, err := emit.Load(schemaDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range emit.Files {
		t.Run(file.Name, func(t *testing.T) {
			path := filepath.Join(tree, file.Name)
			got, err := file.Formatted(m, path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				out := filepath.Join(t.TempDir(), file.Name)
				os.WriteFile(out, got, 0o644)
				t.Errorf("%s differs from the tree; %s\nours is at %s", file.Name, firstDifference(want, got), out)
			}
		})
	}
}

func firstDifference(want, got []byte) string {
	wl, gl := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	for i := range wl {
		if i >= len(gl) {
			return "ours ends at line " + itoa(i) + "; the tree has " + itoa(len(wl)) + " lines"
		}
		if wl[i] != gl[i] {
			return "line " + itoa(i+1) + ":\n  tree: " + wl[i] + "\n  ours: " + gl[i]
		}
	}
	if len(gl) > len(wl) {
		return "ours has " + itoa(len(gl)-len(wl)) + " extra lines from line " + itoa(len(wl)+1)
	}
	return "no differing line (a trailing byte?)"
}

func itoa(n int) string { return strconv.Itoa(n) }
