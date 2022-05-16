package main

import (
	"fmt"
	"os"
	"wpbrute/cmd"
	"wpbrute/cmd/brute"
	"wpbrute/cmd/load"
	"wpbrute/cmd/recon"
	"wpbrute/cmd/report"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Before = cmd.BeforeCommand
	app.After = cmd.AfterCommand
	app.Name = "wpbrute"
	app.Usage = "A bespoke security tool for redteamers to test wordpress credentials on a variety of targets. See the 'help' command for usage instructions."

	app.Commands = []*cli.Command{
		{
			Name: "load",
			Subcommands: []*cli.Command{
				load.LoadTargetsCommand,
				load.LoadProxiesCommand,
			},
		},
		{
			Name: "report",
			Subcommands: []*cli.Command{
				report.ReportDumpGoodies,
				report.ReportClean,
				report.ReportStats,
				report.ReportDisableEdu, // Disable .edu domains - you better not actually be using this for evil.
				report.ReportDisableGov, // Disable .gov domains - you better not actually be using this for evil.
			},
		},
		{
			Name: "recon",
			Subcommands: []*cli.Command{
				recon.ReconCheckHTTPS,
				recon.ReconCheckHTTPS2,
				recon.ReconCheckWPLogin,
				recon.ReconCheckXMLRPC,
			},
		},
		{
			Name: "brute",
			Subcommands: []*cli.Command{
				brute.BruteWPLogin,
				brute.BruteXMLRPCLogin,
			},
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Value: "config.yml",
			Usage: "Config file",
		},
		&cli.BoolFlag{
			Name:  "migrate",
			Value: false,
			Usage: "Whether or not to migrate the database",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
