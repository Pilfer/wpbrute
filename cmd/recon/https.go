package recon

import (
	"fmt"
	"net/url"
	"sync"
	"time"
	"wpbrute/cmd"
	"wpbrute/pkg/models"
	"wpbrute/pkg/wordpress"

	cli "github.com/urfave/cli/v2"
)

// Code to test HTTPS, xmlrpc, wp-login etc here

var ReconCheckHTTPS = &cli.Command{
	Name: "https",
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

		fmt.Printf("Checking HTTPS status for a maximum of %d sites - using proxies: %t\n", numSites, useProxies)

		var targets []*models.Target
		err := cmd.DBConnection.Where("enabled = ? and (is_https is null or is_https_check_time is null)", true).Order("random()").Limit(int(numSites)).Find(&targets).Error
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			fmt.Println("No targets meet the criteria for this check (is_https_check_time = null or is_https = null).")
		} else {
			fmt.Printf("A total number of %d site(s) meet the criteria. Checking now...\n", len(targets))
		}

		var wg sync.WaitGroup
		for _, target := range targets {
			wg.Add(1)
			go checkHTTPS(target, useProxies, &wg)
		}

		fmt.Println("Waiting for results now...")
		wg.Wait()

		return nil
	},
}

func checkHTTPS(t *models.Target, useProxies bool, wg *sync.WaitGroup) {
	fmt.Printf("Checking for HTTPS/Redirects for Target: %s\n", t.Domain)
	proxyString := ""
	if useProxies {
		var proxy *models.Proxy
		err := cmd.DBConnection.Where("enabled = ?", true).Order("random()").Limit(1).Find(&proxy).Error
		if err != nil {
			fmt.Printf("Error fetching proxy from database - skipping for target %s. Error was: %s\n", t.Domain, err)
			wg.Done()
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
		wg.Done()
		return
	}

	_, _, dest, err := wp.IsHTTPS(t)
	if err != nil {
		fmt.Printf("An error occurred while checking HTTPS for target %s with a proxy of %s. Skipping now. Error was %s.\n", t.Domain, proxyString, err)
		wg.Done()
		return
	}

	now := time.Now()
	t.IsHTTPSCheckTime = &now

	err = wp.UpdateRedirectMeta(t, dest)
	if err != nil {
		fmt.Println("An error occurred while updating the target's redirect metadata locally: ", err)
		wg.Done()
		return
	}

	err = cmd.DBConnection.Save(t).Error
	if err != nil {
		fmt.Println("An error occurred while updating the target in the database: ", err)
		wg.Done()
		return
	}

	fmt.Printf("Saved result for %s in the database\n", t.Domain)
	wg.Done()
}
