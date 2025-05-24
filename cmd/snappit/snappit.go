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
		fmt.Fprintf(os.Stderr, "Usage: %s [options] create <name> [<baseline>]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Usage: %s [options] restore <name>\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Usage: %s [options] list\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(1)
	}

	prune := flag.Bool("prune", false, "prune empty directories on restore")
	config_file := flag.String("config", "~/.config/snappit/config.yaml", "override the default config file")
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
	config.Prune = *prune

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
		err = CreateSnapshot(config, name, baseline)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create snapshot: %v\n", err)
			os.Exit(1)
		}

	} else if command == "restore" {
		if flag.NArg() != 2 {
			fmt.Fprintf(os.Stderr, "Error: must specify a snapshot name\n")
			flag.Usage()
		}
		err = RestoreSnapshot(config, flag.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to restore snapshot: %v\n", err)
			os.Exit(1)
		}

	} else if command == "list" {
		if flag.NArg() != 1 {
			fmt.Fprintf(os.Stderr, "Error: too many arguments for list\n")
			flag.Usage()
		}
		err = ListSnapshots(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to list snapshots: %v\n", err)
			os.Exit(1)
		}

	} else {
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
		os.Exit(1)
	}
}
