package utils

import (
	"errors"
	"regexp"
)

var domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)

// Email validator
func ValidateEmail(email string) error {
	return errors.New("not implemented yet")
}

// Domain validator
func ValidateDomain(domain string) bool {
	valid := domainRegexp.MatchString(domain)
	return valid
}
