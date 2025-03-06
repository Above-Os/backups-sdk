package storage

import (
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type RestoreOption struct {
	Space      *options.SpaceRestoreOption
	S3         *options.S3RestoreOption
	Cos        *options.CosRestoreOption
	Filesystem *options.FilesystemRestoreOption
}

type RestoreService struct {
	option *RestoreOption
}

func NewRestoreService(option *RestoreOption) *RestoreService {
	var restoreService = &RestoreService{
		option: option,
	}

	return restoreService
}

func (r *RestoreService) Restore() {
	password, err := utils.InputPasswordWithConfirm(false)
	if err != nil {
		panic(err)
	}

	var service Location

	if r.option.Space != nil {
		service = &space.Space{
			RepoName:       r.option.Space.RepoName,
			SnapshotId:     r.option.Space.SnapshotId,
			Path:           r.option.Space.Path,
			OlaresDid:      r.option.Space.OlaresDid,
			AccessToken:    r.option.Space.AccessToken,
			ClusterId:      r.option.Space.ClusterId,
			CloudName:      strings.ToLower(r.option.Space.CloudName),
			RegionId:       strings.ToLower(r.option.Space.RegionId),
			CloudApiMirror: r.option.Space.CloudApiMirror,
			Password:       password,
			StsToken:       &space.StsToken{},
		}
	} else if r.option.S3 != nil {
		service = &s3.S3{
			RepoName:        r.option.S3.RepoName,
			SnapshotId:      r.option.S3.SnapshotId,
			Endpoint:        r.option.S3.Endpoint,
			AccessKey:       r.option.S3.AccessKey,
			SecretAccessKey: r.option.S3.SecretAccessKey,
			Path:            r.option.S3.Path,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if r.option.Cos != nil {
		service = &cos.Cos{
			RepoName:        r.option.Cos.RepoName,
			SnapshotId:      r.option.Cos.SnapshotId,
			Endpoint:        r.option.Cos.Endpoint,
			AccessKey:       r.option.Cos.AccessKey,
			SecretAccessKey: r.option.Cos.SecretAccessKey,
			Path:            r.option.Cos.Path,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}

	} else if r.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:    r.option.Filesystem.RepoName,
			SnapshotId:  r.option.Filesystem.SnapshotId,
			Endpoint:    r.option.Filesystem.Endpoint,
			Path:        r.option.Filesystem.Path,
			Password:    password,
			BaseHandler: &BaseHandler{},
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	if err := service.Restore(); err != nil {
		logger.Errorf("Restore error: %v", err)
	}
}
