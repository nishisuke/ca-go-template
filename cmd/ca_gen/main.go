package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/nishisuke/ca-go-template/internal/gen"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&gen.GenAPICmd{}, "")
	subcommands.Register(&gen.GenDBCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
