package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rmdashrf/go-misc/cookiejar2"
	"github.com/rmdashrf/go-misc/editthiscookie"
)

func mostSignificantDomain(e []*editthiscookie.Entry) (*url.URL, error) {
	if len(e) == 0 {
		return nil, errors.New("no cookie entries supplied")
	}

	domainCounts := make(map[string]int)
	for _, entry := range e {
		domain := entry.Domain
		if strings.HasPrefix(domain, ".") {
			domain = domain[1:]
		}

		domainCounts[domain] = domainCounts[domain] + 1
	}

	maxCount := -1
	maxDomain := ""

	for d, c := range domainCounts {
		if c > maxCount {
			maxCount = c
			maxDomain = d
		}
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", maxDomain))
	if err != nil {
		return nil, err
	}

	return u, nil

}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s cookiefile", os.Args[0])
		os.Exit(1)
	}

	cookieFile := os.Args[1]
	var r io.Reader
	if cookieFile == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(cookieFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file: %v", err)
			os.Exit(1)
		}

		r = f
		defer f.Close()
	}

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading: %v", err)
		os.Exit(1)
	}

	var etcentries []*editthiscookie.Entry
	if err := json.Unmarshal(contents, &etcentries); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read json: %v", err)
		os.Exit(1)
	}

	msd, err := mostSignificantDomain(etcentries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find most significant domain: %v", err)
		os.Exit(1)
	}

	var gocookies []*http.Cookie
	for _, e := range etcentries {
		gocookies = append(gocookies, e.GoCookie())
	}

	cj := cookiejar2.New(nil)
	cj.SetCookies(msd, gocookies)
	cjEntries := cj.Entries()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	if err := enc.Encode(cjEntries); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode: %v", err)
		os.Exit(1)
	}
}
