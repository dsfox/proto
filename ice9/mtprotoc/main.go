// mtprotoc makes proto/mtproto from proto/ice9/schema, byte for byte what
// teamgram's own mtprotoc made from its scheme.tl (docs/carrying-upstream.md).
//
//	go run ./ice9/mtprotoc generate -schema ice9/schema -out mtproto
//
// The .proto files come out formatted (clang-format, the tree's style) and the
// .go files gofmt-ed; the .pb.go files are protoc's, see build.sh.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamgram/proto/ice9/mtprotoc/emit"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "generate":
		err = runGenerate(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mtprotoc generate [-schema dir] [-out dir]")
	os.Exit(2)
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	schemaDir := fs.String("schema", "ice9/schema", "the schema files")
	out := fs.String("out", "mtproto", "where the generated files go")
	fs.Parse(args)
	m, err := emit.Load(*schemaDir)
	if err != nil {
		return err
	}
	for _, file := range emit.Files {
		text, err := file.Formatted(m, filepath.Join(*out, file.Name))
		if err != nil {
			return fmt.Errorf("%s: %w", file.Name, err)
		}
		if err := os.WriteFile(filepath.Join(*out, file.Name), text, 0o644); err != nil {
			return err
		}
	}
	return nil
}
