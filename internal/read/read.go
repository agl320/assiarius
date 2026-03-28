package read

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/go-shiori/go-readability"
)

func ReadNewsFromLink(link string) error {
	text, err := ReadNewsTextFromLink(link)
	if err != nil {
		return err
	}

	fmt.Println(text)
	return nil
}

// ReadNewsTextFromLink fetches the URL and extracts the primary readable text content.
// This is the reusable core used by both CLI commands and other packages.
func ReadNewsTextFromLink(link string) (string, error) {
	// Methods of extraction:
	// Streaming: no buffering, cannot be re-read, single-pass
	// In-memory buffering (RAM): what we do, fast
	// Disc buffering (Disc): more storage but slower + requires cleanup
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	// cookies
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}

	// headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) "+
		"Chrome/121.0 Safari/537.36")

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	req.Header.Set("Referer", parsedURL.Scheme+"://"+parsedURL.Host)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read bytes response
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", err
	}

	article, err := readability.FromReader(bytes.NewReader(buf.Bytes()), parsedURL)
	if err != nil {
		return "", err
	}

	return article.TextContent, nil
}