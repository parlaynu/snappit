package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// process the command line
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-c config.yml] [-t] create <name> [<baseline>]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Usage: %s [-c config,yml] [-t] reset <name>\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Usage: %s [-c config.yml] list\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(1)
	}

	verbose := flag.Bool("t", false, "run in test only mode")
	config_file := flag.String("c", "~/.config/snappit/config.yaml", "override the default config file")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: must specify a command\n")
		flag.Usage()
	}
	command := strings.ToLower(flag.Arg(0))

	// load the configuration file
	fmt.Printf("Loading config from %s\n", *config_file)
	config, err := LoadConfig(*config_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to %v\n", err)
		os.Exit(1)
	}

	// and... run the command
	if command == "create" {
		if flag.NArg() < 2 || flag.NArg() > 3 {
			fmt.Fprintf(os.Stderr, "Error: must specify a snapshot name\n")
			flag.Usage()
		}
		name := flag.Arg(1)
		baseline := ""
		if flag.NArg() == 3 {
			baseline = flag.Arg(2)
		}
		err = CreateSnapshot(config, name, baseline, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create snapshot: %v\n", err)
			os.Exit(1)
		}

	} else if command == "reset" {
		if flag.NArg() != 2 {
			fmt.Fprintf(os.Stderr, "Error: must specify a snapshot name\n")
			flag.Usage()
		}
		err = ResetSnapshot(config, flag.Arg(1), *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to reset snapshot: %v\n", err)
			os.Exit(1)
		}

	} else if command == "list" {
		if flag.NArg() != 1 {
			fmt.Fprintf(os.Stderr, "Error: too many arguments for list\n")
			flag.Usage()
		}
		err = ListSnapshots(config, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to list snapshots: %v\n", err)
			os.Exit(1)
		}

	} else {
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
		os.Exit(1)
	}
}
