package utils

import "regexp"

var (
	emailRe     = regexp.MustCompile("^([a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+)@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	cellphoneRe = regexp.MustCompile("^1[0-9]{10}$")
	telphoneRe  = regexp.MustCompile("^[0-9]{7,8}$")
)

func IsEmail(emails ...string) bool {
	// TODO: add email max length 320 check:{64}@{255}
	return checkRegexp(emailRe, emails)
}

func IsCellphone(cellphones ...string) bool {
	return checkRegexp(cellphoneRe, cellphones)
}

func IsTelephone(telephones ...string) bool {
	return checkRegexp(telphoneRe, telephones)
}

func checkRegexp(re *regexp.Regexp, data []string) bool {
	for i := range data {
		if !re.MatchString(data[i]) {
			return false
		}
	}
	return true
}
