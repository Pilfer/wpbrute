package report

import (
	"fmt"
	"wpbrute/cmd"

	cli "github.com/urfave/cli/v2"
)

var ReportClean = &cli.Command{
	Name: "clean",
	Action: func(c *cli.Context) error {

		err := cmd.DBConnection.Table("targets").Exec("update targets set domain = lower(domain), target_host = lower(target_host);").Error
		if err != nil {

			fmt.Println(err)
			return err
		}

		return nil
	},
}
