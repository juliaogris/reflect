package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var version = "v0.0.0"

type config struct {
	Address   string           `short:"a" env:"GRPC_ADDRESS" help:"gRPC server address"`
	Plaintext bool             `short:"p" help:"Use plain-text; no TLS" env:"GRPC_PLAINTEXT"`
	Version   kong.VersionFlag `short:"V" help:"Print version information" group:"Other:"`
}

func main() {
	cfg := &config{}
	_ = kong.Parse(cfg, kong.Vars{"version": version})
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cfg *config) error {

	return nil
}
