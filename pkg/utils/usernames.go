package utils

import "fmt"

/*
Take an email input and generate all possible usernames from the email.
Example:
	bob.smith@site.com
Returns:
	[]string{
		bob.smith@site.com,
		bob.smith,
		bob
	}
*/
func GenerateUsernameCombinations(email string) []string {
	invalid := ValidateEmail(email)
	if invalid != nil {
		// still attempt to parse it and pull out a potential username
		fmt.Println(invalid)
	}
	return []string{}
}

// Generate default user/pass combos with all known default user accounts paired with
// every known password input for this domain
func GenerateDefaultCredentials() {

}
