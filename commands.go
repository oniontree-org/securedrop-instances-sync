package main

import (
	"github.com/urfave/cli/v2"
	"time"
)

func (a *Application) commands() {
	a.app = &cli.App{
		Name:    "securedrop-instances",
		Version: Version,
		Usage:   "Interact with SecureDrop's instance API",
		Commands: cli.Commands{
			&cli.Command{
				Name:      "sync",
				Usage:     "Download SecureDrop instances and update an OnionTree repository",
				ArgsUsage: " ",
				Before:    a.handleOnionTreeOpen(),
				Action:    a.handleSyncCommand(),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "url",
						Usage: "SecureDrop API URL",
						Value: "http://sdolvtfhatvsysc6l34d65ymdwxcujausv7k5jk4cy5ttzhjoi6fzvyd.onion/api/v1/directory/",
					},
					&cli.StringSliceFlag{
						Name:  "tag",
						Usage: "attach tags",
						Value: cli.NewStringSlice("securedrop"),
					},
					&cli.DurationFlag{
						Name:  "timeout",
						Usage: "request timeout",
						Value: 15 * time.Second,
					},
				},
			},
		},
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "C",
				Value: ".",
				Usage: "change directory to",
			},
		},
	}
}
