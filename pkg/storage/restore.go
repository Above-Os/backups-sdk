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
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"go.uber.org/zap"
)

type RestoreOption struct {
	Password     string
	Operator     string `json:"operator"`
	BackupType   string `json:"backup_type"` // file / app
	Ctx          context.Context
	Logger       *zap.SugaredLogger
	Space        *options.SpaceRestoreOption        `json:"space,omitempty"`
	Aws          *options.AwsRestoreOption          `json:"aws,omitempty"`
	TencentCloud *options.TencentCloudRestoreOption `json:"tencentcloud,omitempty"`
	Filesystem   *options.FilesystemRestoreOption   `json:"filesystem,omitempty"`
}

type RestoreService struct {
	password string
	option   *RestoreOption
}

func NewRestoreService(option *RestoreOption) *RestoreService {
	var restoreService = &RestoreService{
		password: option.Password,
		option:   option,
	}

	return restoreService
}

func (r *RestoreService) Restore(progressCallback func(percentDone float64)) (restoreSummary map[string]*restic.RestoreSummaryOutput, metadata string, totalBytes uint64, err error) {
	var password = r.password
	if password == "" {
		password, err = utils.InputPasswordWithConfirm(false)
		if err != nil {
			panic(err)
		}
	}

	var service Location

	if r.option.Space != nil {
		service = &space.Space{
			RepoId:   r.option.Space.RepoId,
			RepoName: r.option.Space.RepoName,
			// When restoring from BackupURL on a new machine, it is necessary to extract the Suffix from the Prefix of the backup in the BackupURL
			RepoSuffix:        r.option.Space.RepoSuffix,
			SnapshotId:        r.option.Space.SnapshotId,
			Path:              r.option.Space.Path,
			OlaresDid:         r.option.Space.OlaresDid,
			AccessToken:       r.option.Space.AccessToken,
			ClusterId:         r.option.Space.ClusterId,
			CloudName:         strings.ToLower(r.option.Space.CloudName),
			RegionId:          strings.ToLower(r.option.Space.RegionId),
			CloudApiMirror:    r.option.Space.CloudApiMirror,
			Password:          password,
			LimitDownloadRate: r.option.Space.LimitDownloadRate,
			StsToken:          &space.StsToken{},
			Operator:          r.option.Operator,
			BackupType:        r.option.BackupType,
		}
	} else if r.option.Aws != nil {
		service = &s3.Aws{
			RepoId:            r.option.Aws.RepoId,
			RepoName:          r.option.Aws.RepoName,
			SnapshotId:        r.option.Aws.SnapshotId,
			Endpoint:          r.option.Aws.Endpoint,
			AccessKey:         r.option.Aws.AccessKey,
			SecretAccessKey:   r.option.Aws.SecretAccessKey,
			Path:              r.option.Aws.Path,
			LimitDownloadRate: r.option.Aws.LimitDownloadRate,
			Password:          password,
			BaseHandler:       &BaseHandler{},
			Operator:          r.option.Operator,
			BackupType:        r.option.BackupType,
		}
	} else if r.option.TencentCloud != nil {
		service = &cos.TencentCloud{
			RepoId:            r.option.TencentCloud.RepoId,
			RepoName:          r.option.TencentCloud.RepoName,
			SnapshotId:        r.option.TencentCloud.SnapshotId,
			Endpoint:          r.option.TencentCloud.Endpoint,
			AccessKey:         r.option.TencentCloud.AccessKey,
			SecretAccessKey:   r.option.TencentCloud.SecretAccessKey,
			Path:              r.option.TencentCloud.Path,
			LimitDownloadRate: r.option.TencentCloud.LimitDownloadRate,
			Password:          password,
			BaseHandler:       &BaseHandler{},
			Operator:          r.option.Operator,
			BackupType:        r.option.BackupType,
		}

	} else if r.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoId:      r.option.Filesystem.RepoId,
			RepoName:    r.option.Filesystem.RepoName,
			SnapshotId:  r.option.Filesystem.SnapshotId,
			Endpoint:    r.option.Filesystem.Endpoint,
			Path:        r.option.Filesystem.Path,
			Password:    password,
			BaseHandler: &BaseHandler{},
			Operator:    r.option.Operator,
			BackupType:  r.option.BackupType,
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	restoreOutput, metadata, totalBytes, err := service.Restore(r.option.Ctx, progressCallback)
	if err != nil {
		fmt.Printf("Restore error: %v", err)
	}

	return restoreOutput, metadata, totalBytes, err
}
