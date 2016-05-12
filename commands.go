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
			cli.BoolFlag{
				Name:  "json, j",
				Usage: "Log Format json.",
			},
		},
	},
	{
		Name:   "run",
		Usage:  "Snapshotting for specific instance's volume.",
		Action: command.CmdRun,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "profile, p",
				Usage: "Set AWS-Credentials profile name.",
			},
			cli.StringFlag{
				Name:  "instance-id, i",
				Usage: "Set InstanceId.",
			},
			cli.StringFlag{
				Name:  "region, r",
				Value: "ap-northeast-1",
				Usage: "Set Region.",
			},
			cli.IntFlag{
				Name:  "lifecycle, l",
				Value: 5,
				Usage: "Set the number of life cycle for snapshot.",
			},
			cli.BoolFlag{
				Name:  "json, j",
				Usage: "Log Format json.",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
