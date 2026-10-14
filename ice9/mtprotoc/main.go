// mtprotoc makes proto/mtproto from proto/ice9/schema, byte for byte what
// teamgram's own mtprotoc made from its scheme.tl (docs/carrying-upstream.md).
//
//	go run ./ice9/mtprotoc generate -schema ice9/schema -out mtproto
//	go run ./ice9/mtprotoc import -schema ice9/schema -tlscheme <N.json>
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
	case "import":
		err = runImport(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mtprotoc generate [-schema dir] [-out dir]\n       mtprotoc import -tlscheme <N.json> [-schema dir] [-layer N]")
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
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	// clang-format takes its style from the .clang-format above the file it
	// is told it formats: the tree's, wherever the output goes.
	schemaDirAbs, err := filepath.Abs(*schemaDir)
	if err != nil {
		return err
	}
	tree := filepath.Join(filepath.Dir(filepath.Dir(schemaDirAbs)), "mtproto")
	for _, file := range emit.Files {
		text, err := file.Formatted(m, filepath.Join(tree, file.Name))
		if err != nil {
			// The unformatted text is left beside, so the line the
			// formatter names can be read.
			raw := filepath.Join(*out, file.Name+".unformatted")
			os.WriteFile(raw, file.Emit(m), 0o644)
			return fmt.Errorf("%s: %w (the text is at %s)", file.Name, err, raw)
		}
		if err := os.WriteFile(filepath.Join(*out, file.Name), text, 0o644); err != nil {
			return err
		}
	}
	return nil
}
