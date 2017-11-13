package ircmisc

import "regexp"

var (
	// https://stackoverflow.com/questions/970545/how-to-strip-color-codes-used-by-mirc-users
	regexpStrip = regexp.MustCompile(`\x03(?:\d{1,2}(?:,\d{1,2})?)?`)
)

func StripIrcColors(line string) string {
	return regexpStrip.ReplaceAllLiteralString(line, "")
}
