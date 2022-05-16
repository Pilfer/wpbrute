package report

import (
	"fmt"
	"wpbrute/cmd"

	cli "github.com/urfave/cli/v2"
)

var EnumerateAuthors = &cli.Command{
	Name: "authors",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "site, s",
			Aliases: []string{"s"},
			Usage:   "the site name you'd like to target",
		},
		&cli.Int64Flag{
			Name:    "max, m",
			Aliases: []string{"m"},
			Usage:   "the number of authors to attempt to iterate (defaults to 1)",
			Value:   1,
		},
		&cli.BoolFlag{
			Name:    "useproxy, p",
			Aliases: []string{"p"},
			Usage:   "whether or not you want to perform this task with proxies from the database",
			Value:   false,
		},
	},
	Action: func(c *cli.Context) error {
		var site string
		if c.IsSet("site") {
			site = c.String("site")
		}

		var maxIterations int64
		if c.IsSet("max") {
			maxIterations = c.Int64("max")
		}

		var useProxies bool
		if c.IsSet("useproxy") {
			useProxies = c.Bool("useproxy")
		}

		fmt.Printf("Enumerating authors (max: %d iterations) on site %s - Using proxies: %t\n", maxIterations, site, useProxies)

		err := cmd.DBConnection.Table("targets").Exec("update targets set domain = lower(domain), target_host = lower(target_host);").Error
		if err != nil {

			fmt.Println(err)
			return err
		}

		return nil
	},
}
