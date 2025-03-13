package infraicloud

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/take0244/go-icloud-photo-gui/aop"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	ifICloud struct {
		*authService
		*photoService
		options    ifICloudOptions
		codeCh     chan string
		clientId   string
		signinResp *SigninResponse
		appleId    string
	}
	ifICloudOption  func(*ifICloudOptions)
	ifICloudOptions struct {
		metaDir       string
		photoParallel int
	}
	ifICloudConfigFile struct {
		WebServiceSckdatabasewsUrl string
		AppleId                    string
	}
)

func MetaDirPathOption(cookiePath string) ifICloudOption {
	return func(iio *ifICloudOptions) {
		iio.metaDir = cookiePath
	}
}
func PhotoDownloadParallel(cnt int) ifICloudOption {
	return func(iio *ifICloudOptions) {
		iio.photoParallel = cnt
	}
}

func NewICloud(oauthClientId string, optArgs ...ifICloudOption) usecase.ICloudService {
	return newICloud(oauthClientId, optArgs...)
}

func newICloud(oauthClientId string, optArgs ...ifICloudOption) *ifICloud {
	options := ifICloudOptions{metaDir: filepath.Join(os.TempDir(), "goicloudgui")}
	for _, opt := range optArgs {
		opt(&options)
	}

	fmt.Println(options)

	httpClient := util.HttpClientWithJarCookie()
	instance := &ifICloud{
		authService:  &authService{httpClient: httpClient, oauthClientId: oauthClientId},
		photoService: &photoService{httpClient: httpClient, parallelCnt: options.photoParallel},
		options:      options,
		codeCh:       make(chan string, 1),
	}

	return instance
}

func (i *ifICloud) Code2fa(code string) error {
	aop.Logger().Debug("ifICloud.Code2fa.login2fa")
	if err := i.login2fa(i.clientId, code, i.signinResp); err != nil {
		return err
	}

	aop.Logger().Debug("ifICloud.Code2fa.trustSession")
	trustResponse, err := i.trustSession(i.clientId, i.signinResp)
	if err != nil {
		return err
	}

	aop.Logger().Debug("ifICloud.Code2fa.accountLogin")
	if _, err := i.accountLogin(i.signinResp, trustResponse); err != nil {
		return err
	}

	aop.Logger().Debug("ifICloud.Code2fa.saveMeta")
	i.saveMeta(
		ifICloudConfigFile{
			WebServiceSckdatabasewsUrl: i.photoService.webServiceSckdatabasewsUrl,
			AppleId:                    i.appleId,
		},
	)

	return nil
}

func (i *ifICloud) Clear() {
	if err := i.clearMeta(); err != nil {
		aop.Logger().Error(err.Error())
	}
	newInstance := newICloud(i.authService.oauthClientId)
	newInstance.options = i.options
	i = newInstance
}

func (i *ifICloud) Login(accountName, password string) (*usecase.LoginResult, error) {
	aop.Logger().Debug("ifICloud.Login")

	if i.loadMeta(accountName) {
		aop.Logger().Debug("ifICloud.Login.loadMeta")
		return &usecase.LoginResult{Required2fa: false}, nil
	} else {
		aop.Logger().Debug("ifICloud.Login.DeleteFiles")
		i.clearMeta()
	}

	i.appleId = accountName
	i.clientId = "auth-" + util.MustUUID()

	aop.Logger().Debug("ifICloud.Login.signin")
	var err error
	i.signinResp, err = i.signin(i.clientId, accountName, password)
	if err != nil {
		return nil, err
	}

	aop.Logger().Debug("ifICloud.Login.accountLogin")
	accountResp, err := i.accountLogin(i.signinResp, nil)
	if err != nil {
		return nil, err
	}

	i.photoService.webServiceSckdatabasewsUrl = accountResp.WebServiceSckdatabasewsUrl

	if !accountResp.Required2fa {
		aop.Logger().Debug("ifICloud.Login.saveMeta")
		i.saveMeta(ifICloudConfigFile{
			WebServiceSckdatabasewsUrl: accountResp.WebServiceSckdatabasewsUrl,
			AppleId:                    i.appleId,
		})
	}

	return &usecase.LoginResult{Required2fa: accountResp.Required2fa}, nil
}

func (i *ifICloud) clearMeta() error {
	cookiePath := filepath.Join(i.options.metaDir, "cookie.json")
	metaDataPath := filepath.Join(i.options.metaDir, "meta.json")
	if err := os.Remove(cookiePath); err != nil {
		return err
	}
	if err := os.Remove(metaDataPath); err != nil {
		return err
	}
	return nil
}

func (i *ifICloud) saveMeta(config ifICloudConfigFile) {
	cookiePath := filepath.Join(i.options.metaDir, "cookie.json")
	metaDataPath := filepath.Join(i.options.metaDir, "meta.json")

	if err := util.SaveCookies(cookiePath, i.authService.httpClient.Jar); err != nil {
		aop.Logger().Warn(err.Error())
	}
	if err := util.WriteJson(metaDataPath, config); err != nil {
		aop.Logger().Warn(err.Error())
	}
}

func (i *ifICloud) loadMeta(appleId string) bool {
	cookiePath := filepath.Join(i.options.metaDir, "cookie.json")
	metaDataPath := filepath.Join(i.options.metaDir, "meta.json")
	if err := os.MkdirAll(i.options.metaDir, os.ModePerm); err != nil {
		aop.Logger().Warn(err.Error())
		return false
	}
	if err := util.LoadCookies(cookiePath, i.authService.httpClient.Jar); err != nil {
		aop.Logger().Warn(err.Error())
		return false
	}

	if err := i.validateCookie(); err != nil {
		aop.Logger().Warn(err.Error())
		return false
	}

	meta, err := util.LoadJson[ifICloudConfigFile](metaDataPath)
	if err != nil {
		aop.Logger().Error(err.Error())
		return false
	}

	if meta.WebServiceSckdatabasewsUrl != "" &&
		meta.AppleId == appleId {
		i.appleId = meta.AppleId
		i.photoService.webServiceSckdatabasewsUrl = (*meta).WebServiceSckdatabasewsUrl
		return true
	}

	return false
}

func (i *ifICloud) PhotoService() usecase.PhotoService {
	return i.photoService
}
