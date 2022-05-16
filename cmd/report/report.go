package report

import (
	"fmt"
	"wpbrute/cmd"
	"wpbrute/pkg/models"

	cli "github.com/urfave/cli/v2"
)

var ReportDumpGoodies = &cli.Command{
	Name: "goodies",
	Action: func(c *cli.Context) error {

		var creds []*models.WordpressCredential
		err := cmd.DBConnection.Where("xml_rpc_success = true or wp_login_success = true").Find(&creds).Error
		if err != nil {

			fmt.Println(err)
		}
		if len(creds) == 0 {
			fmt.Println("No new hits..")
		}
		for _, cred := range creds {
			var target *models.Target
			err := cmd.DBConnection.Where("id = ?", cred.TargetID).Find(&target).Error
			if err != nil {
				fmt.Printf("Error fetching target id %d: %s\n", target.ID, err)
			}
			fmt.Printf("%s\t%s\t%s\n", target.Domain, cred.Username, cred.Password)
		}

		return nil
	},
}
