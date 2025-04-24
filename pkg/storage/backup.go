package storage

import (
	"context"
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"go.uber.org/zap"
)

type BackupOption struct {
	Basedir      string
	Password     string
	Operator     string
	Ctx          context.Context
	Logger       *zap.SugaredLogger
	Space        *options.SpaceBackupOption
	Aws          *options.AwsBackupOption
	TencentCloud *options.TencentCloudBackupOption
	Filesystem   *options.FilesystemBackupOption
}

type BackupService struct {
	baseDir  string
	password string
	option   *BackupOption
}

func NewBackupService(option *BackupOption) *BackupService {
	var backupService = &BackupService{
		password: option.Password,
		option:   option,
	}

	return backupService
}

func (b *BackupService) Backup(progressCallback func(percentDone float64)) (*restic.SummaryOutput, *model.StorageInfo, error) {
	var password = b.password
	var err error
	if password == "" {
		password, err = utils.InputPasswordWithConfirm(true)
		if err != nil {
			panic(err)
		}
	}

	var service Location
	if b.option.Space != nil {
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
			Operator:        b.option.Operator,
		}
	} else if b.option.Aws != nil {
		service = &s3.Aws{
			RepoName:        b.option.Aws.RepoName,
			Endpoint:        b.option.Aws.Endpoint,
			AccessKey:       b.option.Aws.AccessKey,
			SecretAccessKey: b.option.Aws.SecretAccessKey,
			Path:            b.option.Aws.Path,
			LimitUploadRate: b.option.Aws.LimitUploadRate,
			Password:        password,
			BaseHandler:     &BaseHandler{},
			Operator:        b.option.Operator,
		}
	} else if b.option.TencentCloud != nil {
		service = &cos.TencentCloud{
			RepoName:        b.option.TencentCloud.RepoName,
			Endpoint:        b.option.TencentCloud.Endpoint,
			AccessKey:       b.option.TencentCloud.AccessKey,
			SecretAccessKey: b.option.TencentCloud.SecretAccessKey,
			Path:            b.option.TencentCloud.Path,
			LimitUploadRate: b.option.TencentCloud.LimitUploadRate,
			Password:        password,
			BaseHandler:     &BaseHandler{},
			Operator:        b.option.Operator,
		}
	} else if b.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:    b.option.Filesystem.RepoName,
			Endpoint:    b.option.Filesystem.Endpoint,
			Path:        b.option.Filesystem.Path,
			Password:    password,
			BaseHandler: &BaseHandler{},
			Operator:    b.option.Operator,
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	summaryOutput, storageInfo, err := service.Backup(b.option.Ctx, progressCallback)

	if err != nil {
		fmt.Printf("Backup error: %v\n", err)
	}
	return summaryOutput, storageInfo, err
}
