package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type Aws struct {
	RepoName          string
	SnapshotId        string
	Endpoint          string
	AccessKey         string
	SecretAccessKey   string
	Password          string
	LimitUploadRate   string
	LimitDownloadRate string
	Path              string
	BaseHandler       base.Interface
	Operator          string
}

func (s *Aws) Backup(ctx context.Context, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error) {
	storageInfo, err = s.FormatRepository()
	if err != nil {
		return
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoName:        s.RepoName,
		Path:            s.Path,
		LimitUploadRate: s.LimitUploadRate,
		RepoEnvs:        envs,
	}

	logger.Debugf("s3 backup env vars: %s", utils.Base64encode([]byte(envs.String())))

	s.BaseHandler.SetOptions(opts)

	backupSummary, err = s.BaseHandler.Backup(ctx, progressCallback)
	return backupSummary, storageInfo, err
}

func (s *Aws) Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error) {
	storageInfo, err := s.FormatRepository()
	if err != nil {
		return
	}
	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoName:          s.RepoName,
		SnapshotId:        s.SnapshotId,
		RepoEnvs:          envs,
		Path:              s.Path,
		LimitDownloadRate: s.LimitDownloadRate,
	}

	logger.Debugf("s3 restore env vars: %s", utils.Base64encode([]byte(envs.String())))

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Restore(ctx, progressCallback)
}

func (s *Aws) Snapshots(ctx context.Context) error {
	storageInfo, err := s.FormatRepository()
	if err != nil {
		return err
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Snapshots(ctx)
}

func (s *Aws) Stats(ctx context.Context) (*restic.StatsContainer, error) {
	storageInfo, err := s.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Stats(ctx)
}

func (s *Aws) Regions() ([]map[string]string, error) {
	return nil, nil
}

func (s *Aws) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		AWS_ACCESS_KEY_ID:     s.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.SecretAccessKey,
		RESTIC_REPOSITORY:     repository,
		RESTIC_PASSWORD:       s.Password,
	}
	return envs
}

func (s *Aws) FormatRepository() (storageInfo *model.StorageInfo, err error) {
	if s.Endpoint == "" {
		err = errors.New("s3 endpoint is required")
		return
	}

	var domainName = constants.StorageS3Domain
	var endpoint = strings.TrimPrefix(s.Endpoint, "https://")
	endpoint = strings.TrimRight(endpoint, "/")
	if strings.EqualFold(endpoint, "") {
		err = fmt.Errorf("s3 endpoint %s is invalid", endpoint)
		return
	}

	var repoSplit = strings.SplitN(endpoint, "/", 2)
	if repoSplit == nil || len(repoSplit) < 1 {
		return nil, fmt.Errorf("s3 endpoint %v is invalid", repoSplit)
	}
	var repoBase = repoSplit[0]
	var repoPrefix = constants.OlaresStorageDefaultPrefix
	if len(repoSplit) >= 2 {
		repoPrefix = fmt.Sprintf("%s/%s", repoSplit[1], constants.OlaresStorageDefaultPrefix)
	}

	var repoBaseSplit = strings.SplitN(repoBase, ".", 3)
	if len(repoBaseSplit) != 3 {
		err = fmt.Errorf("s3 endpoint %v is invalid", repoBaseSplit)
		return
	}
	if repoBaseSplit[2] != domainName {
		err = fmt.Errorf("s3 endpoint %s is not %s", repoBaseSplit[2], domainName)
		return
	}
	var bucket = repoBaseSplit[0]
	var region = repoBaseSplit[1]

	var repository = fmt.Sprintf("s3:https://s3.%s.%s/%s/%s%s", region, domainName, bucket, repoPrefix, s.RepoName)

	storageInfo = &model.StorageInfo{
		Location:  "awss3",
		Url:       repository,
		CloudName: constants.CloudAWSName,
		RegionId:  region,
		Bucket:    bucket,
		Prefix:    repoPrefix,
	}

	return
}
