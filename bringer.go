package watari

import "github.com/PuerkitoBio/goquery"

// DefaultBringer ...
type DefaultBringer struct{}

// Bring ...
func (b *DefaultBringer) Bring(doc *goquery.Document) (token string) {
	token = ""
	return
}
