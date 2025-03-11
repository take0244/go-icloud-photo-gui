package util

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sync"
)

type PersistentCookieJar struct {
	jar        *cookiejar.Jar
	mu         sync.Mutex
	allCookies map[string][]*http.Cookie
}

func NewPersistentCookieJar() *PersistentCookieJar {
	jar, _ := cookiejar.New(nil)
	return &PersistentCookieJar{
		jar:        jar,
		allCookies: make(map[string][]*http.Cookie),
	}
}

func (c *PersistentCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jar.SetCookies(u, cookies)
	c.allCookies[u.Host] = cookies
}

func (c *PersistentCookieJar) Cookies(u *url.URL) []*http.Cookie {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.jar.Cookies(u)
}

func (c *PersistentCookieJar) GetAllCookies() map[string][]*http.Cookie {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allCookies
}

func SaveCookies(cookieFile string, jar http.CookieJar) error {
	data, err := json.Marshal(jar.(*PersistentCookieJar).GetAllCookies())
	if err != nil {
		return err
	}
	return os.WriteFile(cookieFile, data, 0644)
}

func LoadCookies(cookieFile string, jar http.CookieJar) error {
	data, err := os.ReadFile(cookieFile)
	if err != nil {
		return err
	}

	var cookies map[string][]*http.Cookie
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return err
	}

	for domain, ck := range cookies {
		u, _ := url.Parse("https://" + domain)
		jar.SetCookies(u, ck)
	}
	return nil
}
