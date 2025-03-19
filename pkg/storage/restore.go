package storage

import (
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type RestoreOption struct {
	Space        *options.SpaceRestoreOption
	Aws          *options.AwsRestoreOption
	TencentCloud *options.TencentCloudRestoreOption
	Filesystem   *options.FilesystemRestoreOption
}

type RestoreService struct {
	option *RestoreOption
}

func NewRestoreService(option *RestoreOption) *RestoreService {
	var restoreService = &RestoreService{
		option: option,
	}

	return restoreService
}

func (r *RestoreService) Restore() {
	password, err := utils.InputPasswordWithConfirm(false)
	if err != nil {
		panic(err)
	}

	var service Location

	if r.option.Space != nil {
		service = &space.Space{
			RepoName:          r.option.Space.RepoName,
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
		}
	} else if r.option.Aws != nil {
		service = &s3.Aws{
			RepoName:          r.option.Aws.RepoName,
			SnapshotId:        r.option.Aws.SnapshotId,
			Endpoint:          r.option.Aws.Endpoint,
			AccessKey:         r.option.Aws.AccessKey,
			SecretAccessKey:   r.option.Aws.SecretAccessKey,
			Path:              r.option.Aws.Path,
			LimitDownloadRate: r.option.Aws.LimitDownloadRate,
			Password:          password,
			BaseHandler:       &BaseHandler{},
		}
	} else if r.option.TencentCloud != nil {
		service = &cos.TencentCloud{
			RepoName:          r.option.TencentCloud.RepoName,
			SnapshotId:        r.option.TencentCloud.SnapshotId,
			Endpoint:          r.option.TencentCloud.Endpoint,
			AccessKey:         r.option.TencentCloud.AccessKey,
			SecretAccessKey:   r.option.TencentCloud.SecretAccessKey,
			Path:              r.option.TencentCloud.Path,
			LimitDownloadRate: r.option.TencentCloud.LimitDownloadRate,
			Password:          password,
			BaseHandler:       &BaseHandler{},
		}

	} else if r.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:    r.option.Filesystem.RepoName,
			SnapshotId:  r.option.Filesystem.SnapshotId,
			Endpoint:    r.option.Filesystem.Endpoint,
			Path:        r.option.Filesystem.Path,
			Password:    password,
			BaseHandler: &BaseHandler{},
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	if err := service.Restore(); err != nil {
		logger.Errorf("Restore error: %v", err)
	}
}
