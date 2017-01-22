package watari

import (
	"net/http"

	"github.com/juju/persistent-cookiejar"
)

// Client ...
type Client struct {
	HTTP      *http.Client
	CookieJar *cookiejar.Jar
	UserAgent string
}

// SetUserAgent ...
func (c *Client) SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
	return
}

// Get ...
func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	resp, err = c.HTTP.Do(req)
	return
}
