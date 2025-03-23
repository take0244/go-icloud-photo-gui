package util

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

type persistentCookieJar struct {
	jar        *cookiejar.Jar
	mu         sync.Mutex
	allCookies map[string][]*http.Cookie
}

func CastPersistentCookieJar(jar http.CookieJar) *persistentCookieJar {
	return jar.(*persistentCookieJar)
}
func NewPersistentCookieJar() *persistentCookieJar {
	jar, _ := cookiejar.New(nil)
	return &persistentCookieJar{
		jar:        jar,
		allCookies: make(map[string][]*http.Cookie),
	}
}

func (c *persistentCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jar.SetCookies(u, cookies)
	c.allCookies[u.Host] = cookies
}

func (c *persistentCookieJar) Cookies(u *url.URL) []*http.Cookie {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.jar.Cookies(u)
}

func (c *persistentCookieJar) GetAllCookies() map[string][]*http.Cookie {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allCookies
}
