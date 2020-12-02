package helpers

import (
	"bytes"
	"encoding/json"
	"github.com/yuin/goldmark"
	"golang.org/x/time/rate"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"sync"
)

type pageData struct {
	Title  string    `json:"Title"`
	Text   string    `json:"text"`
	Author [4]string `json:"author"`
	Image  []string  `json:"image"`
	Links  []string  `json:"links"`
}

func GetPage(term string) pageData {
	path := "pages/" + term + ".json"
	if FileExists(path) {
		jsonFile, err := os.Open(path)

		if err != nil {
			return pageData{
				Title: "500ERROR",
				Text:  err.Error(),
				Image: nil,
				Links: nil,
			}
		}

		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var data pageData

		_ = json.Unmarshal(byteValue, &data)

		return data

	} else {
		println("Tried accessing a page that doesn't exist")
		return pageData{
			Title: "404ERROR",
			Text:  "",
			Image: nil,
			Links: nil,
		}
	}
}

func MakePage(title string, text string, author [4]string, image []string, links []string) {
	page := pageData{
		Title:  title,
		Text:   text,
		Author: author,
		Image:  image,
		Links:  links,
	}
	file, _ := json.MarshalIndent(page, "", " ")

	_ = ioutil.WriteFile("pages/"+url.QueryEscape(title)+".json", file, 0644)
}

func MarkDownReady(text string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(text), &buf); err != nil {
		panic(err)
	}
	finalTemp := template.HTML(buf.String())
	return finalTemp
}

func MarkDownResponse(text string) string {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(text), &buf); err != nil {
		panic(err)
	}
	return buf.String()
}

// IPRateLimiter .
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter .
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	return i
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}
