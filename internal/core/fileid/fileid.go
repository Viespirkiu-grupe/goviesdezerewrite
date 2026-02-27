package fileid

import "regexp"

var (
	numericRe = regexp.MustCompile(`^[0-9]+$`)
	md5Re     = regexp.MustCompile(`(?i)^[a-f0-9]{32}$`)
)

func IsNumericOrMD5(id string) bool {
	return numericRe.MatchString(id) || md5Re.MatchString(id)
}
