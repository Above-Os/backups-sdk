package storage

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
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
	baseDir string
	option  *SnapshotsOption
}

func NewSnapshotsService(option *SnapshotsOption) *SnapshotsService {
	baseDir := util.GetBaseDir(option.Basedir, common.DefaultBaseDir)

	var snapshotsService = &SnapshotsService{
		baseDir: baseDir,
		option:  option,
	}

	InitLog(baseDir, common.Snapshots)

	return snapshotsService
}

func (s *SnapshotsService) Snapshots() {
	password, err := InputPasswordWithConfirm(common.Snapshots)
	if err != nil {
		panic(err)
	}

	var service Location

	if s.option.Space != nil {
		service = &space.Space{
			RepoName:       s.option.Space.RepoName,
			OlaresId:       s.option.Space.OlaresId,
			CloudApiMirror: s.option.Space.CloudApiMirror,
			BaseDir:        s.baseDir,
			Password:       password,
			UserToken:      &space.UserToken{},
			SpaceToken:     &space.SpaceToken{},
		}
	} else if s.option.S3 != nil {
		service = &s3.S3{
			RepoName:        s.option.S3.RepoName,
			Endpoint:        s.option.S3.Endpoint,
			AccessKey:       s.option.S3.AccessKey,
			SecretAccessKey: s.option.S3.SecretAccessKey,
			Password:        password,
		}
	} else if s.option.Cos != nil {
		service = &cos.Cos{
			RepoName:        s.option.Cos.RepoName,
			Endpoint:        s.option.Cos.Endpoint,
			AccessKey:       s.option.Cos.AccessKey,
			SecretAccessKey: s.option.Cos.SecretAccessKey,
			Password:        password,
		}
	} else if s.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName: s.option.Filesystem.RepoName,
			Endpoint: s.option.Filesystem.Endpoint,
			Password: password,
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
		return
	}

	if err := s.querySnapshots(service); err != nil {
		logger.Errorf("List Spanshots error: %v", err)
	}
}

func (s *SnapshotsService) querySnapshots(service Location) error {
	queryLocation := service.GetLocation()
	if queryLocation == "" {
		return fmt.Errorf("There is no suitable recovery method.")
	}

	if queryLocation == common.LocationSpace {
		return s.querySnapshotsFromSpace(service)
	}

	if queryLocation == common.LocationS3 ||
		queryLocation == common.LocationCos ||
		queryLocation == common.LocationFileSystem {
		return s.querySnapshotsFromCloud(service)
	}

	return nil

}

func (s *SnapshotsService) querySnapshotsFromSpace(service Location) error {
	return service.Snapshots()
}

func (s *SnapshotsService) querySnapshotsFromCloud(service Location) error {
	repository, err := service.FormatRepository()
	if err != nil {
		return err
	}

	envs := service.GetEnv(repository)
	location := service.GetLocation()
	repoName := service.GetRepoName()

	logger.Debugf("%s snapshots env vars: %s", location, util.Base64encode([]byte(envs.ToString())))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := restic.NewRestic(ctx, repoName, envs, nil)
	if err != nil {
		return err
	}

	snapshots, err := r.GetSnapshots()
	if err != nil {
		return err
	}
	snapshots.PrintTable()

	return nil
}
