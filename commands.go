package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heptagon-inc/recorder/command"
)

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "Set LogLevel Debug.",
	},
}

var Commands = []cli.Command{

	{
		Name:   "ebs",
		Usage:  "Snapshotting for specific instance's volume.",
		Action: command.CmdEbs,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "self, s",
				Usage: "Snapshotting for own volume.",
			},
			cli.StringFlag{
				Name:  "instance-id, i",
				Usage: "Set InstanceId. If '--self' option is set, this is ignored.",
			},
			cli.StringFlag{
				Name:  "region, r",
				Value: "ap-northeast-1",
				Usage: "Set Region. If '--self' option is set, this is ignored.",
			},
			cli.IntFlag{
				Name:  "lifecycle, l",
				Value: 5,
				Usage: "Set the number of life cycle for snapshot.",
			},
		},
	},
	{
		Name:   "ami",
		Usage:  "Creating Image for specific instance",
		Action: command.CmdAmi,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "self, s",
				Usage: "Creating Image for own.",
			},
			cli.StringFlag{
				Name:  "instance-id, i",
				Value: "i-xxxxxxx",
				Usage: "Set InstanceId. If '--self' option is set, this is ignored.",
			},
			cli.StringFlag{
				Name:  "region, r",
				Value: "ap-northeast-1",
				Usage: "Set Region. If '--self' option is set, this is ignored.",
			},
			cli.BoolFlag{
				Name:  "reboot",
				Usage: "Reboot instance when create image. (Default-value: false, NOT-Reboot.)",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
