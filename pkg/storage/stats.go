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
)

type StatsOption struct {
	Basedir      string
	Password     string
	Space        *options.SpaceSnapshotsOption
	Aws          *options.AwsSnapshotsOption
	TencentCloud *options.TencentCloudSnapshotsOption
	Filesystem   *options.FilesystemSnapshotsOption
}

type StatsService struct {
	password string
	option   *SnapshotsOption
}

func NewStatsService(option *SnapshotsOption) *StatsService {

	var statsService = &StatsService{
		password: option.Password,
		option:   option,
	}

	return statsService
}

func (s *StatsService) Stats() (*restic.StatsContainer, error) {
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
		return nil, err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := service.Stats(ctx)
	if err != nil {
		logger.Errorf("get stats error: %v", err)
		return nil, err
	}

	fmt.Printf("%s", utils.ToJSON(result))

	return result, nil
}
