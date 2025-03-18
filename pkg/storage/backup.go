package storage

import (
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"go.uber.org/zap"
)

type BackupOption struct {
	Basedir    string
	Password   string
	Logger     *zap.SugaredLogger
	Space      *options.SpaceBackupOption
	S3         *options.S3BackupOption
	Cos        *options.CosBackupOption
	Filesystem *options.FilesystemBackupOption
}

type BackupService struct {
	baseDir  string
	password string
	option   *BackupOption
}

func NewBackupService(option *BackupOption) *BackupService {
	var backupService = &BackupService{
		password: option.Password,
		option:   option,
	}

	return backupService
}

func (b *BackupService) Backup() (*restic.SummaryOutput, string, error) {
	var password = b.password
	var err error
	if password == "" {
		password, err = utils.InputPasswordWithConfirm(true)
		if err != nil {
			panic(err)
		}
	}

	var service Location
	if b.option.Space != nil {
		service = &space.Space{
			RepoName:        b.option.Space.RepoName,
			OlaresDid:       b.option.Space.OlaresDid,
			AccessToken:     b.option.Space.AccessToken,
			ClusterId:       b.option.Space.ClusterId,
			CloudName:       strings.ToLower(b.option.Space.CloudName),
			RegionId:        strings.ToLower(b.option.Space.RegionId),
			Path:            b.option.Space.Path,
			CloudApiMirror:  b.option.Space.CloudApiMirror,
			LimitUploadRate: b.option.Space.LimitUploadRate,
			Password:        password,
			StsToken:        &space.StsToken{},
		}
	} else if b.option.S3 != nil {
		service = &s3.S3{
			RepoName:        b.option.S3.RepoName,
			Endpoint:        b.option.S3.Endpoint,
			AccessKey:       b.option.S3.AccessKey,
			SecretAccessKey: b.option.S3.SecretAccessKey,
			Path:            b.option.S3.Path,
			LimitUploadRate: b.option.S3.LimitUploadRate,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if b.option.Cos != nil {
		service = &cos.Cos{
			RepoName:        b.option.Cos.RepoName,
			Endpoint:        b.option.Cos.Endpoint,
			AccessKey:       b.option.Cos.AccessKey,
			SecretAccessKey: b.option.Cos.SecretAccessKey,
			Path:            b.option.Cos.Path,
			LimitUploadRate: b.option.Cos.LimitUploadRate,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if b.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:    b.option.Filesystem.RepoName,
			Endpoint:    b.option.Filesystem.Endpoint,
			Path:        b.option.Filesystem.Path,
			Password:    password,
			BaseHandler: &BaseHandler{},
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	return service.Backup()
}
