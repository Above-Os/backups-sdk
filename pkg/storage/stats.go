package storage

import (
	"context"
	"strings"
	"time"

	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/restic"
	"olares.com/backups-sdk/pkg/storage/cos"
	"olares.com/backups-sdk/pkg/storage/filesystem"
	"olares.com/backups-sdk/pkg/storage/s3"
	"olares.com/backups-sdk/pkg/storage/space"
	"olares.com/backups-sdk/pkg/utils"
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
			CloudName:       constants.CloudTencentName,
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
		return nil, err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := service.Stats(ctx)
	if err != nil {
		logger.Errorf("get stats error: %v", err)
		return nil, err
	}

	return result, nil
}
