package infraicloud

import (
	"net/http"
	"sync"

	"github.com/take0244/go-icloud-photo-gui/aop"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	SessionData struct {
		client         *http.Client
		clientId       string
		signinResponse *SigninResponse
	}
	SessionManager struct {
		mu          sync.Mutex
		sessionData map[string]*SessionData
	}
)

var sessionManager = &SessionManager{sessionData: make(map[string]*SessionData)}

func (sm *SessionManager) setSigninResponse(userID string, res *SigninResponse) {
	if len(userID) == 0 {
		panic("userID is Empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessionData[userID].signinResponse = res
}

func (sm *SessionManager) getSessionData(userID string) *SessionData {
	if len(userID) == 0 {
		panic("userID is Empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, ok := sm.sessionData[userID]
	if !ok {
		data = &SessionData{}
		sm.sessionData[userID] = data
	}

	if data.clientId == "" {
		data.clientId = "auth-" + util.MustUUID()
	}

	if data.client == nil {
		if aop.IsDebug() {
			data.client = &http.Client{
				Transport: util.NewLoggingTransport(
					util.NewCacheTransport(
						http.DefaultTransport, "./.cache",
					),
				),
				Jar: util.NewPersistentCookieJar(),
			}
		} else {
			data.client = &http.Client{
				Transport: util.NewLoggingTransport(http.DefaultTransport),
				Jar:       util.NewPersistentCookieJar(),
			}
		}
	}

	sm.sessionData[userID] = data

	return data
}
