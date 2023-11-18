package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/jberkenbilt/tfsetup/tfsetup"
	"os"
)

type Args struct {
	Generate bool
	Check    bool
}

func run() error {
	var args Args
	arg.MustParse(&args)
	if (args.Generate && args.Check) || !(args.Generate || args.Check) {
		return fmt.Errorf("exactly one of --generate and --check must be specified")
	}
	ok, err := tfsetup.Run(args.Generate)
	if err != nil {
		return err
	}
	if args.Check {
		if !ok {
			return fmt.Errorf("%s is out of date; rerun tfsetup --generate and terraform init", tfsetup.SetupFile)
		}
		fmt.Printf(`{"message": "%s is current"}`+"\n", tfsetup.SetupFile)
	} else if ok {
		fmt.Printf("%s is already current\n", tfsetup.SetupFile)
	} else {
		fmt.Printf("updated %s\n", tfsetup.SetupFile)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}
