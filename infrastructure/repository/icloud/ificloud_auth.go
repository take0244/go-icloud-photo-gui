package infraicloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	authService    struct{}
	SigninResponse struct {
		SessionId, SessionToken, Scnt, AccountCountry, TrustToken string
	}
	AccountLoginResponse struct {
		Required2fa                bool
		WebServiceSckdatabasewsUrl string
	}
	TrustResponse struct {
		TrustToken, SessionToken string
	}
)

func (a *authService) signin(ctx context.Context, clientId, username, password string) (*SigninResponse, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, config, _, _ := MetaData(ctx)

	req := util.MustRequest(
		ctx,
		http.MethodPost,
		"https://idmsa.apple.com/appleauth/auth/signin",
		bytes.NewBuffer(util.MustMarshal(map[string]any{
			"accountName": username,
			"password":    password,
			"rememberMe":  true,
			"trustTokens": []string{},
		})),
		map[string]string{
			"Accept":                           "*/*",
			"Content-Type":                     "application/json",
			"X-Apple-OAuth-Client-Id":          config.OauthClientId,
			"X-Apple-OAuth-Client-Type":        "firstPartyAuth",
			"X-Apple-OAuth-Redirect-URI":       "https://www.icloud.com",
			"X-Apple-OAuth-Require-Grant-Code": "true",
			"X-Apple-OAuth-Response-Mode":      "web_message",
			"X-Apple-OAuth-Response-Type":      "code",
			"X-Apple-OAuth-State":              clientId,
			"X-Apple-Widget-Key":               config.OauthClientId,
			"User-Agent":                       util.UserAgent,
		},
	)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to signin request: %w", err)
	}
	if !util.HttpCheck2XX(resp) {
		return nil, errors.New("invalid request")
	}

	return &SigninResponse{
		SessionId:      resp.Header.Get("X-Apple-Id-Session-Id"),
		SessionToken:   resp.Header.Get("X-Apple-Session-Token"),
		Scnt:           resp.Header.Get("Scnt"),
		AccountCountry: resp.Header.Get("X-Apple-Id-Account-Country"),
	}, nil
}

func (a *authService) accountLogin(ctx context.Context, signinResponse *SigninResponse, trustResponse *TrustResponse) (*AccountLoginResponse, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, _, _, _ := MetaData(ctx)

	trustToken := func() string {
		if trustResponse != nil {
			return trustResponse.TrustToken
		}
		return ""
	}()

	sessionToken := func() string {
		if trustResponse != nil {
			return trustResponse.SessionToken
		}
		return signinResponse.SessionToken
	}()
	var data = strings.NewReader(fmt.Sprintf(
		`{"accountCountryCode": "%s", "dsWebAuthToken": "%s", "extended_login": true, "trustToken": "%s"}`,
		signinResponse.AccountCountry,
		sessionToken,
		trustToken,
	))

	req := util.MustRequest(
		ctx,
		http.MethodPost,
		"https://setup.icloud.com/setup/ws/1/accountLogin",
		data,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "*/*",
			"Connection":   "keep-alive",
			"Origin":       "https://www.icloud.com",
			"Referer":      "https://www.icloud.com/",
			"User-Agent":   util.UserAgent,
		},
	)

	respMap, err := util.HttpDoJSON[map[string]any](httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("failed to accountLogin request: %w", err)
	}

	return &AccountLoginResponse{
		Required2fa:                (*respMap)["hsaChallengeRequired"].(bool),
		WebServiceSckdatabasewsUrl: (*respMap)["webservices"].(map[string]any)["ckdatabasews"].(map[string]any)["url"].(string),
	}, nil
}

func (a *authService) login2fa(ctx context.Context, clientId, code string, signinResp *SigninResponse) error {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, config, _, _ := MetaData(ctx)

	req := util.MustRequest(
		ctx,
		http.MethodPost,
		"https://idmsa.apple.com/appleauth/auth/verify/trusteddevice/securitycode",
		bytes.NewBuffer(util.MustMarshal(map[string]any{
			"securityCode": map[string]string{
				"code": code,
			},
		})),
		map[string]string{
			"Accept":                           "*/*",
			"Content-Type":                     "application/json",
			"X-Apple-OAuth-Client-Id":          config.OauthClientId,
			"X-Apple-OAuth-Client-Type":        "firstPartyAuth",
			"X-Apple-OAuth-Redirect-URI":       "https://www.icloud.com",
			"X-Apple-OAuth-Require-Grant-Code": "true",
			"X-Apple-OAuth-Response-Mode":      "web_message",
			"X-Apple-OAuth-Response-Type":      "code",
			"X-Apple-OAuth-State":              clientId,
			"X-Apple-Widget-Key":               config.OauthClientId,
			"scnt":                             signinResp.Scnt,
			"X-Apple-ID-Session-Id":            signinResp.SessionId,
			"User-Agent":                       util.UserAgent,
		},
	)

	if _, err := httpClient.Do(req); err != nil {
		return fmt.Errorf("failed to 2fa request: %w", err)
	}

	return nil
}

func (a *authService) trustSession(ctx context.Context, clientId string, signinResp *SigninResponse) (*TrustResponse, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, config, _, _ := MetaData(ctx)

	req := util.MustRequest(
		ctx,
		http.MethodGet,
		"https://idmsa.apple.com/appleauth/auth/2sv/trust",
		nil,
		map[string]string{
			"Accept":                           "*/*",
			"Content-Type":                     "application/json",
			"X-Apple-OAuth-Client-Id":          config.OauthClientId,
			"X-Apple-OAuth-Client-Type":        "firstPartyAuth",
			"X-Apple-OAuth-Redirect-URI":       "https://www.icloud.com",
			"X-Apple-OAuth-Require-Grant-Code": "true",
			"X-Apple-OAuth-Response-Mode":      "web_message",
			"X-Apple-OAuth-Response-Type":      "code",
			"X-Apple-OAuth-State":              clientId,
			"X-Apple-Widget-Key":               config.OauthClientId,
			"scnt":                             signinResp.Scnt,
			"X-Apple-ID-Session-Id":            signinResp.SessionId,
			"User-Agent":                       util.UserAgent,
		},
	)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to trust request: %w", err)
	}
	if !util.HttpCheck2XX(resp) {
		return nil, errors.New("invalid request")
	}

	return &TrustResponse{
		TrustToken:   resp.Header.Get("X-Apple-Twosv-Trust-Token"),
		SessionToken: resp.Header.Get("X-Apple-Session-Token"),
	}, nil
}

func (a *authService) validateCookie(ctx context.Context) error {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, _, _, _ := MetaData(ctx)

	req := util.MustRequest(
		ctx,
		http.MethodGet,
		"https://setup.icloud.com/setup/ws/1/validate",
		nil,
		map[string]string{
			"Accept":     "*/*",
			"Connection": "keep-alive",
			"Origin":     "https://www.icloud.com",
			"Referer":    "https://www.icloud.com/",
			"User-Agent": util.UserAgent,
		},
	)

	if resp, err := httpClient.Do(req); err != nil {
		return fmt.Errorf("failed to trust request: %w", err)
	} else if !util.HttpCheck2XX(resp) {
		return errors.New("invalid cookie")
	}

	return nil
}
