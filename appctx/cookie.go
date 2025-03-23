package appctx

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/take0244/go-icloud-photo-gui/util"
)

const (
	cookieKey      contextKey = "cookie"
	cookieFilename            = "cookie.json"
)

type (
	cookies   map[string][]*http.Cookie
	cookieMap map[string]cookies
)

var (
	dir              = ""
	cache *cookieMap = nil
)

func saveCookie(cm cookieMap) {
	cookieFilePath := filepath.Join(dir, cookieFilename)
	cache = &cm
	err := os.WriteFile(cookieFilePath, util.MustMarshal(cm), 0777)
	if err != nil {
		panic(err)
	}

}

func loadCookie() (cookieMap, bool) {
	cookieFilePath := filepath.Join(dir, cookieFilename)
	if cache != nil {
		return *cache, true
	}

	byts, err := os.ReadFile(cookieFilePath)
	if err != nil {
		return cookieMap{}, false
	}

	localCookies, err := util.Unmarshal[cookieMap](byts)
	if err != nil {
		return cookieMap{}, false
	}

	return *localCookies, true
}

func InitCookies(_dir string) {
	dir = _dir
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}

	_, err := os.OpenFile(filepath.Join(dir, cookieFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
}

func WithCookies(ctx context.Context, key string, c cookies) context.Context {
	if cm, ok := Cookies(ctx); ok {
		cm[key] = c
		saveCookie(cm)
		return context.WithValue(ctx, cookieKey, cm)
	}

	cm := cookieMap{key: c}
	saveCookie(cm)
	return context.WithValue(ctx, cookieKey, cm)
}

func WithCacheCookies(ctx context.Context) context.Context {
	if c, ok := loadCookie(); ok {
		return context.WithValue(ctx, cookieKey, c)
	}

	return context.WithValue(ctx, cookieKey, cookieMap{})
}

func Cookies(ctx context.Context) (cookieMap, bool) {
	_cookies, ok := ctx.Value(cookieKey).(cookieMap)
	if ok {
		return _cookies, ok
	}

	return loadCookie()
}

func ClearCookiesCache(id string) {
	cm, ok := Cookies(context.TODO())
	if !ok {
		return
	}
	_, ok = cm[id]
	if !ok {
		return
	}
	cm[id] = nil
	saveCookie(cm)
}

func ApplyCookieJar(ctx context.Context, jar http.CookieJar, key string) bool {
	cm, ok := Cookies(ctx)
	if !ok {
		return false
	}

	m, ok := cm[key]
	if !ok {
		return false
	}

	for domain, ck := range m {
		u, _ := url.Parse("https://" + domain)
		jar.SetCookies(u, ck)
	}

	return true
}

func CacheCookies(id string, c cookies) {
	cm, _ := Cookies(context.TODO())
	cm[id] = c
	saveCookie(cm)
}
