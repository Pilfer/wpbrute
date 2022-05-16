package report

import (
	"fmt"
	"wpbrute/cmd"

	cli "github.com/urfave/cli/v2"
)

var ReportDisableGov = &cli.Command{
	Name: "disablegov",
	Action: func(c *cli.Context) error {

		err := cmd.DBConnection.Table("targets").Exec(`update targets set enabled = false where domain like '%.gov%';`).Error
		if err != nil {

			fmt.Println(err)
			return err
		}

		return nil
	},
}
