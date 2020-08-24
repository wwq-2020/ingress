package server

import (
	"net/url"
	"regexp"
)

type backend struct {
	path *regexp.Regexp
	url  *url.URL
}

func (b *backend) match(path string) bool {
	return b.path.MatchString(path)
}
