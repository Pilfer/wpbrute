package models

import (
	"time"

	"gorm.io/gorm"
)

type Target struct {
	ID                 int64
	Enabled            bool    // Whether or not we want to attempt to audit this target
	Domain             string  // The root domain name for this target - ex: example.com
	TargetHost         string  // The actionable target hostname for this target - ex: www.example.com or wp.example.com
	Path               string  // The path where Wordpress is installed. Example: "/" or "/blog" or "/website"
	IsHTTPS            *bool   `gorm:"column:is_https"` // Whether or not the target Wordpress instance is accessible behind HTTPS
	XMLRPCEnabled      *bool   // Whether or not the target Wordpress instance has XMLRPC enabled
	XMLRPCLoginEnabled *bool   // Whether or not the target Wordpress instance has XMLRPC logins enabled
	WPLoginEnabled     *bool   // Whether or not the target Wordpress instance has wp-login.php enabled and accessible
	WPLoginCaptcha     *bool   // Whether or not the target Wordpress instance has a Captcha enabled
	WPLoginCaptchaKey  *string // The Captcha key (if recaptcha, and WPLoginCaptcha = true)
	WPLoginSSO         *bool   // Whether or not the target requires authentication with Wordpress.com
	HasWPJson          *bool   // Whether or not the target Wordpress instance /wp-json/ enabled

	IsHTTPSCheckTime *time.Time `gorm:"column:is_https_check_time"` // The time we checked if this site uses HTTPS
	XMLRPCCheckTime  *time.Time // The time we checked if XMLRPC was enabled
	WPLoginCheckTime *time.Time // The time we checked if wp-login.php was enabled and accessible

	gorm.Model
}

type WordpressCredential struct {
	ID             int64
	TargetID       int64  // Parent Target ID
	Enabled        bool   // Whether or not we want to test this credential
	AuthorID       int    // The Wordpress user/author ID
	Username       string // Username or email
	Password       string // Plaintext password
	Role           string // Account role - Admin, Author, User, etc
	XMLRPCSuccess  *bool  // If we successfully logged in with XMLRPC
	WPLoginSuccess *bool  // If we successfully logged in with wp-login.php
	RateLimited    *bool  // If we're rate-limited (any method)

	XMLRPCSuccessTime  *time.Time // When we successfully logged in to this account via XMLRPC
	WPLoginSuccessTime *time.Time // When we successfully logged in to this account via wp-login.php
	XMLRPCFailTime     *time.Time // When we failed to login to this account via XMLRPC
	WPLoginFailTime    *time.Time // When we failed to login to this account via wp-login.php
	RateLimitedTime    *time.Time // When we were rate-limited with this credential

	Target Target // The target for this credential

	gorm.Model
}

type Proxy struct {
	ID       int32
	Enabled  bool   // Whether or not this proxy is enabled
	Protocol string // http, https, etc
	Host     string // Proxy ip or domain
	Port     int    // Port number
	Username string // Proxy auth username
	Password string // Proxy auth password

	gorm.Model
}
