package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"haokun-panel/cmd/frppc/bootstrap"
	"haokun-panel/cmd/frppc/facades"
)

func main() {
	clientCommand, err := injectClientEnvironment(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if clientCommand {
		// The client arguments are handled here and must not be parsed again by
		// Goravel's command runner.
		os.Args = os.Args[:1]
	}

	bootstrap.Boot()
	facades.App().Start()
}

func injectClientEnvironment(args []string) (bool, error) {
	if len(args) == 0 || args[0] == "artisan" {
		return false, nil
	}

	flags := flag.NewFlagSet("client", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var secret, id, rpcURL string
	flags.StringVar(&secret, "s", "", "client secret")
	flags.StringVar(&id, "i", "", "client ID")
	flags.StringVar(&rpcURL, "rpc-url", "", "RPC server URL")

	if err := flags.Parse(args); err != nil {
		return true, fmt.Errorf("client: %w", err)
	}
	if flags.NArg() != 0 {
		return true, fmt.Errorf("client: unexpected argument %q", flags.Arg(0))
	}

	id = strings.TrimSpace(id)
	secret = strings.TrimSpace(secret)
	rpcURL = strings.TrimSpace(rpcURL)
	if id == "" {
		return true, fmt.Errorf("client: -i is required")
	}
	if secret == "" {
		return true, fmt.Errorf("client: -s is required")
	}

	if err := os.Setenv("FRPPC_ID", id); err != nil {
		return true, fmt.Errorf("client: set FRPPC_ID: %w", err)
	}
	if err := os.Setenv("FRPPC_SECRET", secret); err != nil {
		return true, fmt.Errorf("client: set FRPPC_SECRET: %w", err)
	}
	if rpcURL != "" {
		if err := os.Setenv("FRPPC_RPC_URL", rpcURL); err != nil {
			return true, fmt.Errorf("client: set FRPPC_RPC_URL: %w", err)
		}
	}

	return true, nil
}
