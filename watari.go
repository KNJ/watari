package watari

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/KNJ/persistent-cookiejar"
	"github.com/PuerkitoBio/goquery"
)

// ErrRedirectAttempted ...
var ErrRedirectAttempted = errors.New("redirect")

// FormData ...
type FormData interface {
	Get(*goquery.Document, map[string]string)
}

// Profile ...
type Profile struct {
	Destination string
	Login       string
	Username    string
	Password    string
	FormData    FormData
	Credentials *Credentials
}

// NewClient provides a client wrapping http.Client and CookieJar.
// If your session file already exists in your local disk, the
// CookieJar will resume the session.
func NewClient(filePath string) *Client {
	jar, err := cookiejar.New(&cookiejar.Options{Filename: filePath})
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	return &Client{
		HTTP:      client,
		CookieJar: jar,
	}
}

// NewCredentials ...
func NewCredentials(filePath string) *Credentials {
	cred := &Credentials{}
	err := cred.Load(filePath)
	if err != nil {
		log.Fatal("Error@NewCredentials:", err)
	}

	return cred
}

// Source fetches the document as type of string from dest URL.
func Source(client *Client, profile *Profile) (s string, err error) {
	resp, err := Access(client, profile)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	s = string(b)

	return
}

// Scrape ...
func Scrape(client *Client, profile *Profile, fn func(*goquery.Document) interface{}) (result interface{}, err error) {
	s, err := Source(client, profile)
	if err != nil {
		return
	}

	reader := strings.NewReader(s)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		fmt.Println(err)
	}
	result = fn(doc)
	return
}

// Access ...
func Access(client *Client, profile *Profile) (resp *http.Response, err error) {
	resp, auth, err := Attempt(client, profile)
	if auth == true && err != nil {
		fmt.Println(err)
	}

	if auth == false {
		err = errors.New("Failed to be authorized")
	}

	return
}

// Attempt ...
func Attempt(client *Client, profile *Profile) (resp *http.Response, auth bool, err error) {
	auth = true
	client.HTTP.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return ErrRedirectAttempted
	}
	resp, err = client.Get(profile.Destination)

	// If redirected, the client tries to get authorized.
	if urlError, ok := err.(*url.Error); ok && urlError.Err == ErrRedirectAttempted {
		auth = false

		resp, err = client.Get(profile.Login)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		// If status code is not 200
		if resp.StatusCode != 200 {
			fmt.Printf("StatusCode=%d\n", resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		reader := strings.NewReader(string(b))

		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			fmt.Println(err)
		}

		values := url.Values{}
		values.Add(profile.Username, profile.Credentials.Username)
		values.Add(profile.Password, profile.Credentials.Password)
		data := map[string]string{}
		profile.FormData.Get(doc, data)
		for k, v := range data {
			values.Add(k, v)
		}

		// Attempt to sign in.
		req, _ := http.NewRequest("POST", profile.Login, bytes.NewBufferString(values.Encode()))
		req.Header.Add("Referer", profile.Login)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err = client.Do(req)
		defer resp.Body.Close()
		if urlError, ok := err.(*url.Error); ok && urlError.Err == ErrRedirectAttempted {
			// save session
			err = client.CookieJar.Save()
			if err != nil {
				fmt.Println(err)
			}
			auth = true
			resp, err = client.Get(profile.Destination)
		}
	} else {
		// save session
		err = client.CookieJar.Save()
		if err != nil {
			fmt.Println(err)
		}
	}

	return
}
