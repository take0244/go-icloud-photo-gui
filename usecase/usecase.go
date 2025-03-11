package usecase

type (
	UseCase interface {
		ICloudService() ICloudService
	}
	useCase struct {
		iCloudService ICloudService
	}
)

type (
	LoginResult struct {
		Required2fa bool
	}
	ICloudService interface {
		PhotoService() PhotoService
		Login(username, password string) (*LoginResult, error)
		Code2fa(code string) error
		Clear()
	}
	PhotoService interface {
		DownloadAllPhotos(dir string) error
	}
)

func NewUseCase(rep1 ICloudService) UseCase {
	return &useCase{
		iCloudService: rep1,
	}
}

func (u *useCase) ICloudService() ICloudService {
	return u.iCloudService
}
