package load

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"wpbrute/cmd"
	"wpbrute/pkg/models"
	"wpbrute/pkg/utils"

	cli "github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Command to load the proxy list
var LoadTargetsCommand = &cli.Command{
	Name: "targets",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file, f",
			Aliases: []string{"f"},
			Value:   "targets.txt",
		},
	},
	Action: func(c *cli.Context) error {
		var filename string
		if c.IsSet("file") {
			filename = c.String("file")
		} else {
			filename = "targets.txt"
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

		// essentially: { "domain.com": [{user, pass}, {user, pass}]}
		targetCreds := make(map[string][]models.WordpressCredential)

		for _, line := range lines {
			if len(line) != 0 {
				segments := strings.Split(line, ":")
				if len(segments) == 2 {
					var email, password, domain string

					email = segments[0]
					password = segments[1]
					innerseg := strings.Split(email, "@")
					domain = strings.ToLower(innerseg[1])

					// Validate the domain name
					if !utils.ValidateDomain(domain) {
						fmt.Printf("❌ Domain %s is invalid - skipping it! ❌\n", domain)
					} else {
						// Append to the domain slice
						targetCreds[domain] = append(targetCreds[domain], models.WordpressCredential{
							Username: email,
							Password: password,
						})
					}
				}
			}
		}

		listContains := func(username, password string, list []models.WordpressCredential) bool {
			for _, l := range list {
				if l.Username == username && l.Password == password {
					return true
				}
			}
			return false
		}

		for domain, list := range targetCreds {
			pwList := []string{}
			for _, cred := range list {
				pwList = append(pwList, cred.Password)
			}

			// Set admins here.
			for _, pw := range pwList {
				if !listContains("admin", pw, list) {
					list = append(list, models.WordpressCredential{
						Username: "admin",
						Password: pw,
					})
				}
			}

			// If there's an email, add <this-username-here>@<domain> to the credentials list
			// and "."-split prefix to the credential list for this domain
			for _, cred := range list {
				if strings.Contains(cred.Username, "@") {
					user := strings.Split(cred.Username, "@")[0]
					if len(user) > 0 {
						list = append(list, models.WordpressCredential{
							Username: user,
							Password: cred.Password,
						})

						if strings.Contains(user, ".") {
							prefix := strings.Split(user, ".")[0]
							if len(prefix) > 0 {

								list = append(list, models.WordpressCredential{
									Username: prefix,
									Password: cred.Password,
								})
							}
						}
					}
				}
			}

			// Assign the newly created list here
			targetCreds[domain] = list
		}

		fmt.Println("All done setting up the list.. iterating now.")

		// Insert all unique targets here
		uniqueDomains := []string{}
		for domain := range targetCreds {
			if !contains(uniqueDomains, domain) {
				uniqueDomains = append(uniqueDomains, domain)
			}
		}

		var targetList []*models.Target

		for _, domain := range uniqueDomains {
			t := &models.Target{}

			// Check if exists
			err := cmd.DBConnection.First(t, "domain = ?", domain).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					fmt.Printf("Error finding domain %s: %s\n", domain, err)
					continue
				}
			}

			if t.ID == 0 {
				t.Domain = domain
				t.Enabled = true
				t.Path = "/"
				err = cmd.DBConnection.Create(t).Error
				if err != nil {
					fmt.Println("error creating target record", err)
					continue
				}
			}

			targetList = append(targetList, t)
		}

		// Iterate over all of these. Add targets and credentials ezpz.
		for domain, list := range targetCreds {
			var t *models.Target
			for _, predefinedTarget := range targetList {
				if predefinedTarget.Domain == domain {
					t = predefinedTarget
					break
				}
			}

			// break loop if nil target
			if t == nil {
				break
			}

			// start woop
			var creds []*models.WordpressCredential
			for _, cred := range list {
				wc := &models.WordpressCredential{
					TargetID: t.ID,
				}
				wc.Username = cred.Username
				wc.Password = cred.Password
				cred.TargetID = t.ID
				cred.XMLRPCSuccessTime = nil
				cred.WPLoginSuccessTime = nil
				cred.XMLRPCFailTime = nil
				cred.WPLoginFailTime = nil
				cred.RateLimitedTime = nil
				creds = append(creds, wc)
			}

			err := cmd.DBConnection.CreateInBatches(creds, 1000).Error
			if err != nil {
				fmt.Println("Error inserting items", err)
			}
			// end woop
		}

		/*
			1. Load and sanitize credentials
			2. Explode credentials list into targets and place in database
			3. Permutate potential defaults (admin, etc)
			3. Store credentials in database
		*/

		var targetCount int64
		err = cmd.DBConnection.Model(&models.Target{}).Count(&targetCount).Error
		if err != nil {
			return err
		}

		var credsCount int64
		err = cmd.DBConnection.Model(&models.WordpressCredential{}).Count(&credsCount).Error
		if err != nil {
			return err
		}

		fmt.Printf("All done! We have a total of %d target domains and %d credentials in the database.\n", targetCount, credsCount)
		return nil
	},
}
