package recon

import (
	"fmt"
	"net/url"
	"time"
	"wpbrute/cmd"
	"wpbrute/pkg/models"
	"wpbrute/pkg/wordpress"

	"github.com/gammazero/workerpool"
	cli "github.com/urfave/cli/v2"
)

// Code to test HTTPS, xmlrpc, wp-login etc here

var ReconCheckWPLogin = &cli.Command{
	Name: "wplogin",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:    "number, n",
			Aliases: []string{"n"},
			Usage:   "the number of targets you'd like to use for this",
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
		var numSites int64
		if c.IsSet("number") {
			numSites = c.Int64("number")
		}

		var useProxies bool
		if c.IsSet("useproxy") {
			useProxies = c.Bool("useproxy")
		}

		fmt.Printf("Checking WPLogin status for a maximum of %d sites - using proxies: %t\n", numSites, useProxies)

		var targets []*models.Target
		err := cmd.DBConnection.Where("enabled = ? and is_https_check_time is not null and (wp_login_check_time is null or wp_login_enabled is null)", true).Order("random()").Limit(int(numSites)).Find(&targets).Error
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			fmt.Println("No targets meet the criteria for this check (is_https_check_time = null or is_https = null).")
		} else {
			fmt.Printf("A total number of %d site(s) meet the criteria. Checking now...\n", len(targets))
		}

		wp := workerpool.New(cmd.Config.WPBruteConfig.WorkerCount)
		for _, target := range targets {
			t := target
			wp.Submit(func() {
				checkWPLoginfunc(t, useProxies)
			})
		}

		fmt.Println("Waiting now...")
		wp.StopWait()
		return nil
	},
}

func checkWPLoginfunc(t *models.Target, useProxies bool) {

	fmt.Printf("Checking for wplogin.php for Target: %s\n", t.Domain)
	proxyString := ""
	if useProxies {
		var proxy *models.Proxy
		err := cmd.DBConnection.Where("enabled = ?", true).Order("random()").Limit(1).Find(&proxy).Error
		if err != nil {
			fmt.Printf("Error fetching proxy from database - skipping for target %s. Error was: %s\n", t.Domain, err)
			return
		}

		purl := url.URL{
			Scheme: proxy.Protocol,
			Host:   fmt.Sprintf("%s:%d", proxy.Host, proxy.Port),
			User:   url.UserPassword(proxy.Username, proxy.Password),
		}

		proxyString = purl.String()

	}
	wp, err := wordpress.New(proxyString, nil)
	if err != nil {
		fmt.Printf("An error occurred while instantiating the Wordpress instance for target %s with a proxy of %s. Skipping now! Error was: %s\n", t.Domain, proxyString, err)
		return
	}

	result, err := wp.CheckWPLogin(t)
	if err != nil {
		fmt.Printf("An error occurred while checking wp-login target %s with a proxy of %s. Skipping now! Error was: %s\n", t.Domain, proxyString, err)
		return
	}

	now := time.Now()
	t.WPLoginEnabled = &result.Enabled
	t.WPLoginCaptcha = &result.Captcha
	t.WPLoginCaptchaKey = &result.CaptchaKey
	t.WPLoginSSO = &result.WPSSO
	t.WPLoginCheckTime = &now

	if result.Enabled {
		fmt.Printf("wp-login on %s is enabled\n", t.Domain)
	} else {
		fmt.Printf("wp-login on %s is disabled\n", t.Domain)
	}

	err = cmd.DBConnection.Save(t).Error
	if err != nil {
		fmt.Printf("An error occurred while updating the target %s in the database. Error was: %s\n", t.Domain, err)
		return
	}
}
