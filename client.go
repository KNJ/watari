package watari

import (
	"net/http"
	"time"

	"github.com/KNJ/persistent-cookiejar"
)

// Client ...
type Client struct {
	HTTP      *http.Client
	CookieJar *cookiejar.Jar
	UserAgent string
}

// SetTimeout ...
func (c *Client) SetTimeout(d time.Duration) {
	c.HTTP.Timeout = d
}

// SetUserAgent ...
func (c *Client) SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
	return
}

// Get ...
func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", c.UserAgent)
	resp, err = c.HTTP.Do(req)
	return
}

// Do ...
func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	req.Header.Add("User-Agent", c.UserAgent)
	resp, err = c.HTTP.Do(req)
	return
}
