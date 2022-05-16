package load

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"wpbrute/cmd"
	"wpbrute/pkg/models"

	cli "github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

// Command to load the proxy list
var LoadProxiesCommand = &cli.Command{
	Name: "proxies",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file, f",
			Aliases: []string{"f"},
			Value:   "proxies.txt",
		},
	},
	Action: func(c *cli.Context) error {
		var filename string
		if c.IsSet("file") {
			filename = c.String("file")
		} else {
			filename = "proxies.txt"
		}

		fi, err := os.Stat(filename)
		if err != nil {
			return err
		}

		if fi.Size() == 0 {
			return fmt.Errorf("it looks like %s is empty", filename)
		}

		f, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		fileStr := string(f)
		lines := strings.Split(fileStr, "\n")
		var proxies []models.Proxy

		for _, line := range lines {
			if len(line) != 0 {
				pUrl, err := url.Parse(line)

				if err != nil {
					fmt.Printf("Proxy parse error: %s", err)
					continue
				}

				tmpProxy := models.Proxy{
					Enabled:  true,
					Protocol: pUrl.Scheme,
					Host:     pUrl.Hostname(),
				}

				if len(pUrl.User.Username()) > 0 {
					tmpProxy.Username = pUrl.User.Username()
				}

				pass, hasPass := pUrl.User.Password()
				if hasPass && len(pass) > 0 {
					tmpProxy.Password = pass
				}

				if len(pUrl.Port()) >= 1 {
					port, err := strconv.Atoi(pUrl.Port())
					if err != nil {
						fmt.Println("An error occurred while parsing the port for proxy host: ", pUrl.Hostname(), " : ", err)
					} else {
						tmpProxy.Port = port
					}
				}
				proxies = append(proxies, tmpProxy)
			}
		}

		for _, proxy := range proxies {
			dbProxy := &models.Proxy{}
			err := cmd.DBConnection.First(&dbProxy, "host = ? and port = ?", proxy.Host, proxy.Port).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					fmt.Printf("Error checking to see if proxy exists in db already: %s\n", err)
					continue
				}
			}

			if dbProxy.ID != 0 {
				if (dbProxy.Username != proxy.Username) || (dbProxy.Password != proxy.Password) || (dbProxy.Port != proxy.Port) || (dbProxy.Protocol != proxy.Protocol) {
					dbProxy.Enabled = true
					dbProxy.Username = proxy.Username
					dbProxy.Password = proxy.Password
					dbProxy.Host = proxy.Host
					dbProxy.Port = proxy.Port
					dbProxy.Protocol = proxy.Protocol
					err := cmd.DBConnection.Save(dbProxy).Error
					if err != nil {
						fmt.Println(err)
						continue
					}
				}
			} else {
				dbProxy.Enabled = true
				dbProxy.Username = proxy.Username
				dbProxy.Password = proxy.Password
				dbProxy.Host = proxy.Host
				dbProxy.Port = proxy.Port
				dbProxy.Protocol = proxy.Protocol
				err = cmd.DBConnection.Create(dbProxy).Error
				if err != nil {
					fmt.Println("error creating proxy record", proxy.Host, err)
					continue
				}
			}
		}

		var proxyCount int64
		err = cmd.DBConnection.Table("proxies").Where("enabled = ?", true).Count(&proxyCount).Error
		if err != nil {
			return err
		}

		fmt.Printf("All done! We have a total of %d enabled proxies in the database.\n", proxyCount)
		return nil
	},
}
