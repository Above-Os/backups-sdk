package storage

import (
	"context"
	"strings"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"go.uber.org/zap"
)

type SnapshotsOption struct {
	Basedir      string
	Password     string
	Logger       *zap.SugaredLogger
	Space        *options.SpaceSnapshotsOption
	Aws          *options.AwsSnapshotsOption
	TencentCloud *options.TencentCloudSnapshotsOption
	Filesystem   *options.FilesystemSnapshotsOption
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
	password, err := utils.InputPasswordWithConfirm(false)
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
			CloudName:      strings.ToLower(s.option.Space.CloudName),
			RegionId:       strings.ToLower(s.option.Space.RegionId),
			CloudApiMirror: s.option.Space.CloudApiMirror,
			Password:       password,
			StsToken:       &space.StsToken{},
		}
	} else if s.option.Aws != nil {
		service = &s3.Aws{
			RepoName:        s.option.Aws.RepoName,
			Endpoint:        s.option.Aws.Endpoint,
			AccessKey:       s.option.Aws.AccessKey,
			SecretAccessKey: s.option.Aws.SecretAccessKey,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if s.option.TencentCloud != nil {
		service = &cos.TencentCloud{
			RepoName:        s.option.TencentCloud.RepoName,
			Endpoint:        s.option.TencentCloud.Endpoint,
			AccessKey:       s.option.TencentCloud.AccessKey,
			SecretAccessKey: s.option.TencentCloud.SecretAccessKey,
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

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := service.Snapshots(ctx); err != nil {
		logger.Errorf("List Spanshots error: %v", err)
	}
}
