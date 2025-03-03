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
	baseDir := util.GetBaseDir(option.Basedir, common.DefaultBaseDir)

	var backupService = &BackupService{
		baseDir: baseDir,
		option:  option,
	}

	InitLog(baseDir, common.Backup)

	return backupService
}

func (b *BackupService) Backup() {
	password, err := InputPasswordWithConfirm(common.Backup)
	if err != nil {
		panic(err)
	}

	var service Location
	if b.option.Space != nil {
		service = &space.Space{
			RepoName:        b.option.Space.RepoName,
			OlaresId:        b.option.Space.OlaresId,
			Path:            b.option.Space.Path,
			CloudApiMirror:  b.option.Space.CloudApiMirror,
			LimitUploadRate: b.option.Space.LimitUploadRate,
			BaseDir:         b.option.Space.BaseDir,
			Password:        password,
			UserToken:       &space.UserToken{},
			SpaceToken:      &space.SpaceToken{},
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
		}
	} else if b.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName: b.option.Filesystem.RepoName,
			Endpoint: b.option.Filesystem.Endpoint,
			Path:     b.option.Filesystem.Path,
			Password: password,
		}
	} else { // todo unittest
		logger.Fatalf("There is no suitable recovery method.")
	}

	if err := b.startBackup(service); err != nil {
		logger.Errorf("Backup to Space error: %v", err)
	}
}

func (b *BackupService) startBackup(service Location) error {
	backupToLocation := service.GetLocation()
	if backupToLocation == "" {
		return fmt.Errorf("There is no suitable recovery method.")
	}

	if backupToLocation == common.LocationSpace {
		return b.backupToSpace(service)
	}

	if backupToLocation == common.LocationS3 ||
		backupToLocation == common.LocationCos ||
		backupToLocation == common.LocationFileSystem {
		return b.backupToCloud(service)
	}

	return nil
}

func (b *BackupService) backupToSpace(service Location) error {
	return service.Backup()
}

func (b *BackupService) backupToCloud(service Location) error {
	repository, err := service.FormatRepository()
	if err != nil {
		return err
	}

	envs := service.GetEnv(repository)

	repoName := service.GetRepoName()
	backupPath := service.GetPath()

	r, err := restic.NewRestic(context.Background(), repoName, envs, &restic.Option{LimitUploadRate: service.GetLimitUploadRate()})
	if err != nil {
		return err
	}

	_, initRepo, err := r.Init()
	if err != nil {
		return err
	}

	if !initRepo {
		if err = r.Repair(); err != nil {
			return err
		}
	}

	backupResult, err := r.Backup(backupPath, "")
	if err != nil {
		return err
	}
	logger.Infof("Backup to %s success, result id: %s", service.GetLocation(), util.ToJSON(backupResult))

	return nil
}
