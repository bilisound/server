package utils

import "strings"

var allowedRedirectDomains = []string{"b23.tv"}
var allowedImageDomains = []string{"hdslb.com"}

func IsAllowedRedirectDomain(hostname string) bool {
	for _, allowedDomain := range allowedRedirectDomains {
		if strings.HasSuffix(hostname, allowedDomain) {
			return true
		}
	}
	return false
}

func IsAllowedImageDomain(hostname string) bool {
	for _, allowedDomain := range allowedImageDomains {
		if strings.HasSuffix(hostname, allowedDomain) {
			return true
		}
	}
	return false
}
