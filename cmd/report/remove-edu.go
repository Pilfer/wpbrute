package report

import (
	"fmt"
	"wpbrute/cmd"

	cli "github.com/urfave/cli/v2"
)

var ReportDisableEdu = &cli.Command{
	Name: "disableedu",
	Action: func(c *cli.Context) error {

		err := cmd.DBConnection.Table("targets").Exec(`update targets set enabled = false where domain like '%.edu%';`).Error
		if err != nil {

			fmt.Println(err)
			return err
		}

		return nil
	},
}
