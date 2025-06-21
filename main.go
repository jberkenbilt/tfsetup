package main

import (
	"errors"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/jberkenbilt/tfsetup/tfsetup"
	"github.com/jberkenbilt/tfsetup/util"
	"os"
	"path/filepath"
)

const version = "1.1.0"

type Args struct {
	Generate bool `help:"regenerate outdated files"`
	Check    bool `help:"check all generated files against expected output"`
	Render   bool `help:"render standard input with current context"`
	Version  bool
}

func run() error {
	var args Args
	arg.MustParse(&args)
	if args.Version {
		fmt.Printf("%s %s\n", filepath.Base(os.Args[0]), version)
		return nil
	}
	if util.CountTrue(args.Generate, args.Check, args.Render) != 1 {
		return fmt.Errorf("exactly one of --generate, --check, or --render must be specified")
	}
	if args.Render {
		return tfsetup.Render()
	}
	ok, err := tfsetup.Run(args.Generate)
	if err != nil {
		return err
	}
	if args.Check {
		if !ok {
			return errors.New("some files are out of date; rerun tfsetup --generate and terraform init")
		}
		fmt.Println(`{"message": "all files are current"}`)
	} else if ok {
		fmt.Println("all files are already current")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}
