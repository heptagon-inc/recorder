package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heptagon-inc/recorder/command"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{

	{
		Name:   "self",
		Usage:  "Snapshotting for own volume.",
		Action: command.CmdSelf,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "lifecycle, l",
				Value: 5,
				Usage: "Set the number of life cycle for snapshot.",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
