package wordpress

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"wpbrute/pkg/models"
)

type WPLoginCheckResult struct {
	Enabled    bool
	WPSSO      bool
	Captcha    bool
	CaptchaKey string
}

func (wp *Wordpress) CheckWPLogin(target *models.Target) (*WPLoginCheckResult, error) {
	/*
		Check the target for the existence of wplogin
	*/
	protocol := "http"
	if target.IsHTTPS != nil && *target.IsHTTPS {
		protocol = "https"
	}
	targetHost := target.Domain
	if len(target.TargetHost) != 0 {
		targetHost = target.TargetHost
	}
	path := ""
	if len(target.Path) > 0 {
		path = target.Path
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	path = fmt.Sprintf("%swp-login.php", path)

	tUrl, err := url.Parse(fmt.Sprintf("%s://%s%s", protocol, targetHost, path))
	if err != nil {
		fmt.Println("Unable to construct url for ", target.Domain, " error of: ", err)
		return &WPLoginCheckResult{}, err
	}

	req := &http.Request{
		Method: "GET",
		URL:    tUrl,
		Header: http.Header{
			"User-Agent":                []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"},
			"Accept":                    []string{"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			"accept-language":           []string{"en-US,en;q=0.9,de;q=0.8"},
			"cache-control":             []string{"no-cache"},
			"pragma":                    []string{"no-cache"},
			"upgrade-insecure-requests": []string{"1"},
			"cookie":                    []string{"wordpress_test_cookie=WP%20Cookie%20check"},
		},
	}

	fmt.Println(req.URL)

	resp, err := wp.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &WPLoginCheckResult{}, nil
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	result := &WPLoginCheckResult{}
	if strings.Contains(bodyString, `name="log"`) && strings.Contains(bodyString, `id="user_login"`) && strings.Contains(bodyString, `name="pwd"`) && strings.Contains(bodyString, `id="user_pass"`) {
		result.Enabled = true
	}

	return result, nil
}
