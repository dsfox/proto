// mtprotoc makes proto/mtproto from proto/ice9/schema, byte for byte what
// teamgram's own mtprotoc made from its scheme.tl (docs/carrying-upstream.md).
//
//	go run ./ice9/mtprotoc oracle -tree mtproto -tlscheme <dir of N.json> -out ice9/schema
//	go run ./ice9/mtprotoc generate -schema ice9/schema -out mtproto
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamgram/proto/ice9/mtprotoc/oracle"
	"github.com/teamgram/proto/ice9/mtprotoc/schema"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "oracle":
		err = runOracle(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mtprotoc oracle|generate [flags]")
	os.Exit(2)
}

func runOracle(args []string) error {
	fs := flag.NewFlagSet("oracle", flag.ExitOnError)
	tree := fs.String("tree", "mtproto", "the generated files as they are")
	tlscheme := fs.String("tlscheme", "", "Telegram's per-layer schema files, <layer>.json")
	out := fs.String("out", "ice9/schema", "where the schema files go")
	top := fs.Int("layer", 228, "the top layer the tree knows")
	fs.Parse(args)
	if *tlscheme == "" {
		return fmt.Errorf("-tlscheme is required")
	}
	registers, err := oracle.ReadRegisters(filepath.Join(*tree, "class_name_registers.go"))
	if err != nil {
		return err
	}
	telegram, err := oracle.ReadTelegram(*tlscheme)
	if err != nil {
		return err
	}
	api, service, teamgram, report := oracle.Build(registers, telegram, *top, "resPQ", "int32")
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	for name, s := range map[string]*schema.Schema{"layers.tl": api, "mtproto.tl": service, "teamgram.tl": teamgram} {
		f, err := os.Create(filepath.Join(*out, name))
		if err != nil {
			return err
		}
		if err := schema.Write(f, s); err != nil {
			return err
		}
		f.Close()
	}
	services, err := oracle.ReadServices(filepath.Join(*tree, "schema.tl.sync_service.proto"))
	if err != nil {
		return err
	}
	if err := oracle.WriteServices(filepath.Join(*out, "services.txt"), services); err != nil {
		return err
	}
	report.Say("services %d, rpcs %d", len(services.Order), countRPCs(services))
	for _, line := range report.Lines {
		fmt.Println(line)
	}
	return nil
}

func countRPCs(s *oracle.Services) int {
	n := 0
	for _, rpcs := range s.RPCs {
		n += len(rpcs)
	}
	return n
}
