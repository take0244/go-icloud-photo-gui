package infraicloud

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	ifICloud struct {
		*authService
		*photoService
	}
)

func NewICloud() usecase.ICloudService {
	instance := &ifICloud{
		authService:  &authService{},
		photoService: &photoService{},
	}

	return instance
}

func (i *ifICloud) Login(ctx context.Context, appleId, password string) (bool, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	_, _, _, user := MetaData(ctx)
	session := sessionManager.getSessionData(user.ID)

	if i.loadMeta(ctx) {
		return false, nil
	}

	signinResp, err := i.signin(ctx, session.clientId, appleId, password)
	if err != nil {
		return false, err
	}

	accountResp, err := i.accountLogin(ctx, signinResp, nil)
	if err != nil {
		return false, err
	}

	sessionManager.setSigninResponse(user.ID, signinResp)

	appctx.PeekConfig(ctx, func(conf *appctx.ConfigFile) {
		conf.AppleInfo[user.ID] = appctx.AppleInfo{
			AppleId:                    appleId,
			WebServiceSckdatabasewsUrl: accountResp.WebServiceSckdatabasewsUrl,
		}
	})

	if !accountResp.Required2fa {
		appctx.CacheCookies(user.ID, util.CastPersistentCookieJar(session.client.Jar).GetAllCookies())
	}

	return accountResp.Required2fa, nil
}

func (i *ifICloud) Code2fa(ctx context.Context, code string) error {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	_, _, _, user := MetaData(ctx)
	session := sessionManager.getSessionData(user.ID)

	if session.signinResponse == nil {
		return errors.New("no signin")
	}

	if err := i.login2fa(ctx, session.clientId, code, session.signinResponse); err != nil {
		return err
	}

	trustResponse, err := i.trustSession(ctx, session.clientId, session.signinResponse)
	if err != nil {
		return err
	}

	if _, err := i.accountLogin(ctx, session.signinResponse, trustResponse); err != nil {
		return err
	}
	appctx.CacheCookies(user.ID, util.CastPersistentCookieJar(session.client.Jar).GetAllCookies())

	return nil
}

func (i *ifICloud) loadMeta(ctx context.Context) (okCookie bool) {
	_, _, _, user := MetaData(ctx)
	session := sessionManager.getSessionData(user.ID)
	defer func() {
		if !okCookie {
			appctx.ClearCookiesCache(user.ID)
		}
	}()

	if !appctx.ApplyCookieJar(ctx, session.client.Jar, user.ID) {
		return false
	}

	if err := i.validateCookie(ctx); err != nil {
		slog.WarnContext(ctx, err.Error())
		return false
	}

	_, _, appleInfo, _ := MetaData(ctx)
	if appleInfo.WebServiceSckdatabasewsUrl != "" {
		return true
	}

	return false
}

func MetaData(ctx context.Context) (*http.Client, appctx.ConfigFile, appctx.AppleInfo, *appctx.ContextUser) {
	user := appctx.User(ctx)
	config := appctx.Config(ctx)
	appInfo := config.AppleInfo[user.ID]
	httpClient := sessionManager.getSessionData(user.ID).client

	return httpClient, config, appInfo, user
}
