package storage

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/restic"
	"olares.com/backups-sdk/pkg/storage/cos"
	"olares.com/backups-sdk/pkg/storage/filesystem"
	"olares.com/backups-sdk/pkg/storage/model"
	"olares.com/backups-sdk/pkg/storage/s3"
	"olares.com/backups-sdk/pkg/storage/space"
	"olares.com/backups-sdk/pkg/utils"
)

type BackupOption struct {
	Basedir                  string
	Password                 string
	Operator                 string
	BackupType               string // file / app
	BackupAppTypeName        string // if app
	BackupFileTypeSourcePath string // if file
	Ctx                      context.Context
	Logger                   *zap.SugaredLogger
	Space                    *options.SpaceBackupOption
	Aws                      *options.AwsBackupOption
	TencentCloud             *options.TencentCloudBackupOption
	Filesystem               *options.FilesystemBackupOption
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

func (b *BackupService) Backup(dryRun bool, progressCallback func(percentDone float64)) (*restic.SummaryOutput, *model.StorageInfo, error) {
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
			RepoId:                   b.option.Space.RepoId,
			RepoName:                 b.option.Space.RepoName,
			OlaresDid:                b.option.Space.OlaresDid,
			AccessToken:              b.option.Space.AccessToken,
			ClusterId:                b.option.Space.ClusterId,
			CloudName:                strings.ToLower(b.option.Space.CloudName),
			RegionId:                 strings.ToLower(b.option.Space.RegionId),
			Path:                     b.option.Space.Path,
			Files:                    b.option.Space.Files,
			FilesPrefixPath:          b.option.Space.FilesPrefixPath,
			Metadata:                 b.option.Space.Metadata,
			CloudApiMirror:           b.option.Space.CloudApiMirror,
			LimitUploadRate:          b.option.Space.LimitUploadRate,
			Password:                 password,
			StsToken:                 &space.StsToken{},
			Operator:                 b.option.Operator,
			BackupType:               b.option.BackupType,
			BackupAppTypeName:        b.option.BackupAppTypeName,
			BackupFileTypeSourcePath: b.option.BackupFileTypeSourcePath,
		}
	} else if b.option.Aws != nil {
		service = &s3.Aws{
			RepoId:                   b.option.Aws.RepoId,
			RepoName:                 b.option.Aws.RepoName,
			Endpoint:                 b.option.Aws.Endpoint,
			AccessKey:                b.option.Aws.AccessKey,
			SecretAccessKey:          b.option.Aws.SecretAccessKey,
			Path:                     b.option.Aws.Path,
			Files:                    b.option.Aws.Files,
			FilesPrefixPath:          b.option.Aws.FilesPrefixPath,
			Metadata:                 b.option.Aws.Metadata,
			LimitUploadRate:          b.option.Aws.LimitUploadRate,
			Password:                 password,
			BaseHandler:              &BaseHandler{},
			Operator:                 b.option.Operator,
			BackupType:               b.option.BackupType,
			BackupAppTypeName:        b.option.BackupAppTypeName,
			BackupFileTypeSourcePath: b.option.BackupFileTypeSourcePath,
		}
	} else if b.option.TencentCloud != nil {
		service = &cos.TencentCloud{
			RepoId:                   b.option.TencentCloud.RepoId,
			RepoName:                 b.option.TencentCloud.RepoName,
			Endpoint:                 b.option.TencentCloud.Endpoint,
			AccessKey:                b.option.TencentCloud.AccessKey,
			SecretAccessKey:          b.option.TencentCloud.SecretAccessKey,
			Path:                     b.option.TencentCloud.Path,
			Files:                    b.option.TencentCloud.Files,
			FilesPrefixPath:          b.option.TencentCloud.FilesPrefixPath,
			Metadata:                 b.option.TencentCloud.Metadata,
			LimitUploadRate:          b.option.TencentCloud.LimitUploadRate,
			Password:                 password,
			BaseHandler:              &BaseHandler{},
			Operator:                 b.option.Operator,
			BackupType:               b.option.BackupType,
			BackupAppTypeName:        b.option.BackupAppTypeName,
			BackupFileTypeSourcePath: b.option.BackupFileTypeSourcePath,
		}
	} else if b.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoId:                   b.option.Filesystem.RepoId,
			RepoName:                 b.option.Filesystem.RepoName,
			Endpoint:                 b.option.Filesystem.Endpoint,
			Path:                     b.option.Filesystem.Path,
			Files:                    b.option.Filesystem.Files,
			FilesPrefixPath:          b.option.Filesystem.FilesPrefixPath,
			Metadata:                 b.option.Filesystem.Metadata,
			Password:                 password,
			BaseHandler:              &BaseHandler{},
			Operator:                 b.option.Operator,
			BackupType:               b.option.BackupType,
			BackupAppTypeName:        b.option.BackupAppTypeName,
			BackupFileTypeSourcePath: b.option.BackupFileTypeSourcePath,
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	summaryOutput, storageInfo, err := service.Backup(b.option.Ctx, dryRun, progressCallback)

	if err != nil {
		fmt.Printf("Backup error: %v\n", err)
	}
	return summaryOutput, storageInfo, err
}
