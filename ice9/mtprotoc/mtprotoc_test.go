package main

import (
	"bytes"
	"os"
	"os/exec"
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
	for _, file := range []struct {
		name  string
		emit  func(*emit.Model) []byte
		proto bool
	}{
		{"schema.tl.crc32.proto", emit.CRC32Proto, true},
	} {
		t.Run(file.name, func(t *testing.T) {
			got := file.emit(m)
			if file.proto {
				got = clangFormat(t, got, filepath.Join(tree, file.name))
			}
			want, err := os.ReadFile(filepath.Join(tree, file.name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				out := filepath.Join(t.TempDir(), file.name)
				os.WriteFile(out, got, 0o644)
				t.Errorf("%s differs from the tree; %s\nours is at %s", file.name, firstDifference(want, got), out)
			}
		})
	}
}

// clangFormat formats a .proto as the tree's .clang-format says, as the
// build does.
func clangFormat(t *testing.T, text []byte, as string) []byte {
	t.Helper()
	cmd := exec.Command("clang-format", "-assume-filename="+as)
	cmd.Stdin = bytes.NewReader(text)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("clang-format: %v: %s", err, stderr.String())
	}
	return out
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
