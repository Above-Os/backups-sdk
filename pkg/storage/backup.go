package storage

import (
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type BackupOption struct {
	Basedir    string
	Space      *options.SpaceBackupOption
	S3         *options.S3BackupOption
	Cos        *options.CosBackupOption
	Filesystem *options.FilesystemBackupOption
}

type BackupService struct {
	baseDir string
	option  *BackupOption
}

func NewBackupService(option *BackupOption) *BackupService {
	var backupService = &BackupService{
		option: option,
	}

	return backupService
}

func (b *BackupService) Backup() {
	password, err := utils.InputPasswordWithConfirm(true)
	if err != nil {
		panic(err)
	}

	var service Location
	if b.option.Space != nil {
		if utils.ContainsPathSeparator(b.option.Space.RepoName) {
			panic(fmt.Errorf("repo name contains path separator: '\\' or '/', name: %s", b.option.Space.RepoName))
		}
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
		if utils.ContainsPathSeparator(b.option.S3.RepoName) {
			panic(fmt.Errorf("repo name contains path separator: '\\' or '/', name: %s", b.option.S3.RepoName))
		}
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
		if utils.ContainsPathSeparator(b.option.Cos.RepoName) {
			panic(fmt.Errorf("repo name contains path separator: '\\' or '/', name: %s", b.option.Cos.RepoName))
		}
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
		if utils.ContainsPathSeparator(b.option.Filesystem.RepoName) {
			panic(fmt.Errorf("repo name contains path separator: '\\' or '/', name: %s", b.option.Filesystem.RepoName))
		}
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

	if err := service.Backup(); err != nil {
		logger.Errorf("Backup error: %v", err)
	}
}
