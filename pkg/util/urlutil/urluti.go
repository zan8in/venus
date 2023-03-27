package urlutil

import "github.com/bobesa/go-domain-util/domainutil"

// Get the top level domain from url
func TopDomain(url string) string {
	return domainutil.Domain(url)
}
