package handler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

const (
	feedURL = "https://vercel.com/atom"
	author  = "Rob Herley"
)

type entry struct {
	ID    string `xml:"id" json:"id"`
	Title string `xml:"title" json:"title"`
	Link  struct {
		Href string `xml:"href,attr" json:"href"`
	} `xml:"link" json:"link"`
	Updated string   `xml:"updated" json:"updated"`
	Authors []string `xml:"author>name" json:"authors"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(feedURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetching feed: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	var feed struct {
		Entries []entry `xml:"entry"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		http.Error(w, fmt.Sprintf("parsing feed: %v", err), http.StatusBadGateway)
		return
	}

	entries := []entry{}
	for _, e := range feed.Entries {
		if slices.ContainsFunc(e.Authors, func(name string) bool {
			return strings.Contains(name, author)
		}) {
			entries = append(entries, e)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "s-maxage=3600, stale-while-revalidate=86400")
	json.NewEncoder(w).Encode(map[string]any{
		"count":   len(entries),
		"entries": entries,
	})
}
