package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

const schemaDir = "../schema"

var schemaFiles = []string{"layers.tl", "mtproto.tl", "teamgram.tl"}

// Every id written in the schema is the crc32 of its own line, as TL defines
// it, so a definition copied wrong cannot carry a right id. The lines where
// teamgram's id is not that are listed in overrides.txt with the reason, and
// an override that is no longer needed is refused too.
func TestEveryIDIsTheCRC32OfItsLine(t *testing.T) {
	overrides, err := readOverrides(filepath.Join(schemaDir, "overrides.txt"))
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	var paths []string
	for _, name := range schemaFiles {
		paths = append(paths, filepath.Join(schemaDir, name))
	}
	s, err := schema.ParseFiles(paths...)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range s.Entities {
		for _, v := range e.Versions {
			if !v.Explicit {
				continue
			}
			key := fmt.Sprintf("%s %08x", e.Predicate(), v.ID)
			computed := schema.CRC32(e.Name, v)
			if computed == v.ID {
				if overrides[key] != "" {
					t.Errorf("%s: the id is its line's crc32 now; drop the override (%s)", key, overrides[key])
				}
				continue
			}
			if overrides[key] != "" {
				used[key] = true
				continue
			}
			t.Errorf("%s at layer %d: the line's crc32 is %08x: %s", key, v.Layer, computed, schema.Canonical(e.Name, v))
		}
	}
	for key, why := range overrides {
		if !used[key] {
			t.Errorf("override of %s names no line in the schema (%s)", key, why)
		}
	}
}

// readOverrides reads "predicate id reason" lines.
func readOverrides(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	overrides := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.SplitN(line, " ", 3)
		if len(fields) < 3 {
			return nil, fmt.Errorf("%s: an override is predicate, id and reason: %s", path, line)
		}
		id, err := strconv.ParseUint(fields[1], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: bad id in %s", path, line)
		}
		overrides[fmt.Sprintf("%s %08x", fields[0], uint32(id))] = fields[2]
	}
	return overrides, scanner.Err()
}
