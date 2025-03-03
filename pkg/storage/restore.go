package storage

import (
	"context"
	"fmt"

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

type RestoreOption struct {
	Basedir    string
	Space      *options.SpaceRestoreOption
	S3         *options.S3RestoreOption
	Cos        *options.CosRestoreOption
	Filesystem *options.FilesystemRestoreOption
}

type RestoreService struct {
	baseDir string
	option  *RestoreOption
}

func NewRestoreService(option *RestoreOption) *RestoreService {
	baseDir := util.GetBaseDir(option.Basedir, common.DefaultBaseDir)

	var restoreService = &RestoreService{
		baseDir: baseDir,
		option:  option,
	}

	InitLog(baseDir, common.Restore)

	return restoreService
}

func (r *RestoreService) Restore() {
	password, err := InputPasswordWithConfirm(common.Restore)
	if err != nil {
		panic(err)
	}

	var service Location

	if r.option.Space != nil {
		service = &space.Space{
			RepoName:       r.option.Space.RepoName,
			SnapshotId:     r.option.Space.SnapshotId,
			Path:           r.option.Space.Path,
			OlaresId:       r.option.Space.OlaresId,
			CloudApiMirror: r.option.Space.CloudApiMirror,
			BaseDir:        r.baseDir,
			Password:       password,
			UserToken:      &space.UserToken{},
			SpaceToken:     &space.SpaceToken{},
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
		}

	} else if r.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:   r.option.Filesystem.RepoName,
			SnapshotId: r.option.Filesystem.SnapshotId,
			Endpoint:   r.option.Filesystem.Endpoint,
			Path:       r.option.Filesystem.Path,
			Password:   password,
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	if err := r.startRestore(service); err != nil {
		logger.Errorf("Restore from Space error: %v", err)
	}
}

func (r *RestoreService) startRestore(service Location) error {
	restoreFromLocation := service.GetLocation()
	if restoreFromLocation == "" {
		return fmt.Errorf("There is no suitable recovery method.")
	}

	if restoreFromLocation == common.LocationSpace {
		return r.restoreFromSpace(service)
	}

	if restoreFromLocation == common.LocationS3 ||
		restoreFromLocation == common.LocationCos ||
		restoreFromLocation == common.LocationFileSystem {
		return r.restoreFromCloud(service)
	}

	return nil
}

func (r *RestoreService) restoreFromSpace(service Location) error {
	return service.Restore()
}

func (r *RestoreService) restoreFromCloud(service Location) error {
	repository, err := service.FormatRepository()
	if err != nil {
		return err
	}
	envs := service.GetEnv(repository)

	location := service.GetLocation()
	repoName := service.GetRepoName()
	restorePath := service.GetPath()
	snapshotId := service.GetSnapshotId()

	logger.Debugf("%s restore env vars: %s", location, util.Base64encode([]byte(envs.ToString())))

	re, err := restic.NewRestic(context.Background(), repoName, envs, nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := re.GetSnapshot(snapshotId)
	if err != nil {
		return err
	}

	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("%s restore spanshot %s detail: %s", location, snapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = re.Restore(snapshotId, uploadPath, restorePath)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore from %s successful, data: %s", location, util.ToJSON(summary))
	}

	return nil
}
