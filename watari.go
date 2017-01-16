package watari

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CSRFTokenBringer ...
type CSRFTokenBringer interface {
	Bring(*goquery.Document) string
}

// Profile ...
type Profile struct {
	Destination      string
	Login            string
	Username         string
	Password         string
	CSRFToken        string
	CSRFTokenBringer CSRFTokenBringer
	Credentials      *Credentials
}

// Credentials ...
type Credentials struct {
	Username string
	Password string
}

// ErrRedirectAttempted ...
var ErrRedirectAttempted = errors.New("redirect")

// Access ...
func Access(client *http.Client, profile *Profile) (resp *http.Response, err error) {
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
func Attempt(client *http.Client, profile *Profile) (resp *http.Response, auth bool, err error) {
	auth = true
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return ErrRedirectAttempted
	}
	resp, err = client.Get(profile.Destination)

	// If redirected, the client tries to get authorized.
	if urlError, ok := err.(*url.Error); ok && urlError.Err == ErrRedirectAttempted {
		auth = false
		if profile.CSRFTokenBringer == nil {
			return
		}

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

		token := profile.CSRFTokenBringer.Bring(doc)

		values := url.Values{}
		values.Add(profile.Username, profile.Credentials.Username)
		values.Add(profile.Password, profile.Credentials.Password)
		values.Add(profile.CSRFToken, token)

		// Attempt to sign in.
		req, _ := http.NewRequest("POST", profile.Login, bytes.NewBufferString(values.Encode()))
		req.Header.Add("Referer", profile.Login)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err = client.Do(req)
		defer resp.Body.Close()
		if urlError, ok := err.(*url.Error); ok && urlError.Err == ErrRedirectAttempted {
			auth = true
			resp, err = client.Get(profile.Destination)
		}
	}

	return
}
