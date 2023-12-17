package utils

import (
	"strings"
)

//goland:noinspection HttpUrlsUsage
func ReplaceAnySchemeWithHttp(endpoint string) string {
	if !strings.Contains(endpoint, "://") { // not contains scheme
		if strings.HasPrefix(endpoint, "//") {
			endpoint = "http:" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	} else if strings.HasPrefix(endpoint, "tcp://") {
		// replace with http scheme
		endpoint = "http" + endpoint[3:]
	} else if strings.HasPrefix(endpoint, "http://") {
		// keep
	} else if strings.HasPrefix(endpoint, "https://") {
		// keep
	} else {
		// replace with http scheme
		endpoint = "http" + endpoint[strings.Index(endpoint, "://"):]
	}
	return endpoint
}
