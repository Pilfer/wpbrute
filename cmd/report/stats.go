package report

import (
	"fmt"
	"wpbrute/cmd"

	cli "github.com/urfave/cli/v2"
)

var ReportStats = &cli.Command{
	Name: "stats",
	Action: func(c *cli.Context) error {
		var proxyCount int64
		err := cmd.DBConnection.Table("proxies").Where("enabled = ?", true).Count(&proxyCount).Error
		if err != nil {
			return err
		}

		var targetCount int64
		err = cmd.DBConnection.Table("targets").Where("enabled = ?", true).Count(&targetCount).Error
		if err != nil {
			return err
		}

		var xmlrpcEnabledTargetCount int64
		err = cmd.DBConnection.Table("targets").Where("enabled = ? and xml_rpc_enabled = ?", true, true).Count(&xmlrpcEnabledTargetCount).Error
		if err != nil {
			return err
		}

		var xmlrpcDisabledTargetCount int64
		err = cmd.DBConnection.Table("targets").Where("enabled = ? and xml_rpc_enabled = ?", true, false).Count(&xmlrpcDisabledTargetCount).Error
		if err != nil {
			return err
		}

		var wploginEnabledTargetCount int64
		err = cmd.DBConnection.Table("targets").Where("enabled = ? and wp_login_enabled = ?", true, true).Count(&wploginEnabledTargetCount).Error
		if err != nil {
			return err
		}
		var wploginDisabledTargetCount int64
		err = cmd.DBConnection.Table("targets").Where("enabled = ? and wp_login_enabled = ?", true, false).Count(&wploginDisabledTargetCount).Error
		if err != nil {
			return err
		}

		var credCount int64
		err = cmd.DBConnection.Table("wordpress_credentials").Count(&credCount).Error
		if err != nil {
			return err
		}

		var goodCredCount int64
		err = cmd.DBConnection.Table("wordpress_credentials").Where("xml_rpc_success = true or wp_login_success = true").Count(&goodCredCount).Error
		if err != nil {
			return err
		}

		var badCredCount int64
		err = cmd.DBConnection.Table("wordpress_credentials").Where("xml_rpc_success = false or wp_login_success = false").Count(&badCredCount).Error
		if err != nil {
			return err
		}

		fmt.Println("------------------------------------")
		fmt.Println("Proxy count:", proxyCount)
		fmt.Println("Target count:", targetCount)
		fmt.Println("XMLRPC-enabled target count", xmlrpcEnabledTargetCount)
		fmt.Println("XMLRPC-disabled target count", xmlrpcDisabledTargetCount)
		fmt.Println("wp-login-enabled target count", wploginEnabledTargetCount)
		fmt.Println("wp-login-disabled target count", wploginEnabledTargetCount)
		fmt.Println("Credentials count:", credCount)
		fmt.Println("------------------------------------")
		fmt.Println("Tried and failed credentials:", badCredCount)
		fmt.Println("Found credentials:", goodCredCount)
		return nil
	},
}
