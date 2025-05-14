package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

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

type SnapshotsOption struct {
	Basedir      string
	Password     string
	Operator     string
	Logger       *zap.SugaredLogger
	Space        *options.SpaceSnapshotsOption
	Aws          *options.AwsSnapshotsOption
	TencentCloud *options.TencentCloudSnapshotsOption
	Filesystem   *options.FilesystemSnapshotsOption
}

type SnapshotsService struct {
	password string
	option   *SnapshotsOption
}

func NewSnapshotsService(option *SnapshotsOption) *SnapshotsService {

	var snapshotsService = &SnapshotsService{
		password: option.Password,
		option:   option,
	}

	return snapshotsService
}

func (s *SnapshotsService) Snapshots() (*restic.SnapshotList, error) {
	var password = s.password
	var err error
	if password == "" {
		password, err = utils.InputPasswordWithConfirm(true)
		if err != nil {
			panic(err)
		}
	}

	var service Location
	if s.option.Space != nil {
		service = &space.Space{
			RepoId:         s.option.Space.RepoId,
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
			RepoId:          s.option.Aws.RepoId,
			RepoName:        s.option.Aws.RepoName,
			Endpoint:        s.option.Aws.Endpoint,
			AccessKey:       s.option.Aws.AccessKey,
			SecretAccessKey: s.option.Aws.SecretAccessKey,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if s.option.TencentCloud != nil {
		service = &cos.TencentCloud{
			RepoId:          s.option.TencentCloud.RepoId,
			RepoName:        s.option.TencentCloud.RepoName,
			Endpoint:        s.option.TencentCloud.Endpoint,
			AccessKey:       s.option.TencentCloud.AccessKey,
			SecretAccessKey: s.option.TencentCloud.SecretAccessKey,
			Password:        password,
			BaseHandler:     &BaseHandler{},
		}
	} else if s.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoId:      s.option.Filesystem.RepoId,
			RepoName:    s.option.Filesystem.RepoName,
			Endpoint:    s.option.Filesystem.Endpoint,
			Password:    password,
			BaseHandler: &BaseHandler{},
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
		return nil, fmt.Errorf("There is no suitable recovery method.")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := service.Snapshots(ctx)
	if err != nil {
		logger.Errorf("List Spanshots error: %v", err)
		return nil, err
	}
	return result, nil
}
