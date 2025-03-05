package storage

import (
	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

type SnapshotsOption struct {
	Basedir    string
	Space      *options.SpaceSnapshotsOption
	S3         *options.S3SnapshotsOption
	Cos        *options.CosSnapshotsOption
	Filesystem *options.FilesystemSnapshotsOption
}

type SnapshotsService struct {
	option *SnapshotsOption
}

func NewSnapshotsService(option *SnapshotsOption) *SnapshotsService {

	var snapshotsService = &SnapshotsService{
		option: option,
	}

	return snapshotsService
}

func (s *SnapshotsService) Snapshots() {
	password, err := util.InputPasswordWithConfirm(false)
	if err != nil {
		panic(err)
	}

	var service Location
	if s.option.Space != nil {
		service = &space.Space{
			RepoName:       s.option.Space.RepoName,
			OlaresDid:      s.option.Space.OlaresDid,
			AccessToken:    s.option.Space.AccessToken,
			ClusterId:      s.option.Space.ClusterId,
			CloudName:      s.option.Space.CloudName,
			RegionId:       s.option.Space.RegionId,
			CloudApiMirror: s.option.Space.CloudApiMirror,
			Password:       password,
			StsToken:       &space.StsToken{},
		}
	} else if s.option.S3 != nil {
		service = &s3.S3{
			RepoName:        s.option.S3.RepoName,
			Endpoint:        s.option.S3.Endpoint,
			AccessKey:       s.option.S3.AccessKey,
			SecretAccessKey: s.option.S3.SecretAccessKey,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if s.option.Cos != nil {
		service = &cos.Cos{
			RepoName:        s.option.Cos.RepoName,
			Endpoint:        s.option.Cos.Endpoint,
			AccessKey:       s.option.Cos.AccessKey,
			SecretAccessKey: s.option.Cos.SecretAccessKey,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if s.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:    s.option.Filesystem.RepoName,
			Endpoint:    s.option.Filesystem.Endpoint,
			Password:    password,
			BaseHandler: &BaseHandler{},
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
		return
	}

	if err := service.Snapshots(); err != nil {
		logger.Errorf("List Spanshots error: %v", err)
	}
}
