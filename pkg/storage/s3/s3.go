package s3

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type Aws struct {
	RepoId            string
	RepoName          string
	SnapshotId        string
	Endpoint          string
	AccessKey         string
	SecretAccessKey   string
	Password          string
	LimitUploadRate   string
	LimitDownloadRate string
	Path              string
	Files             []string
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
		RepoId:          s.RepoId,
		RepoName:        s.RepoName,
		Path:            s.Path,
		Files:           s.Files,
		LimitUploadRate: s.LimitUploadRate,
		Operator:        s.Operator,
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
		RepoId:            s.RepoId,
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

func (s *Aws) Snapshots(ctx context.Context) (*restic.SnapshotList, error) {
	storageInfo, err := s.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:   s.RepoId,
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	logger.Debugf("s3 snapshots env vars: %s", utils.Base64encode([]byte(envs.String())))

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
		RepoId:   s.RepoId,
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	logger.Debugf("s3 stats env vars: %s", utils.Base64encode([]byte(envs.String())))

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

	var endpoint = strings.TrimRight(s.Endpoint, "/")

	s3UrlInfo, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var host = s3UrlInfo.Host
	var hosts = strings.Split(host, ".")
	if len(hosts) < 4 {
		return nil, fmt.Errorf("host invalid, host: %s", host)
	}

	if !strings.Contains(host, constants.StorageS3Domain) {
		return nil, fmt.Errorf("host is not s3 format, host: %s", host)
	}

	var region, bucket, prefix string
	path := strings.TrimLeft(s3UrlInfo.Path, "/")
	paths := []string{}
	if path != "" {
		paths = strings.Split(path, "/")
	}

	if hosts[0] == "s3" {
		region = hosts[1]
		if len(paths) == 0 {
			return nil, fmt.Errorf("bucket not exists in path: %s", path)
		}
		bucket = paths[0]
		if len(paths) > 1 {
			prefix = strings.Join(paths[1:], "/")
		}
	} else {
		bucket = hosts[0]
		region = hosts[1]

		prefix = path
	}

	if prefix == "" {
		prefix = constants.OlaresStorageDefaultPrefix
	} else {
		prefix = fmt.Sprintf("%s/%s", prefix, constants.OlaresStorageDefaultPrefix)
	}

	repository := fmt.Sprintf("s3:%s://s3.%s.amazonaws.com/%s/%s/%s-%s", s3UrlInfo.Scheme, region, bucket, prefix, utils.EncodeURLPart(s.RepoName), s.RepoId)

	storageInfo = &model.StorageInfo{
		Location:  "awss3",
		Url:       repository,
		CloudName: constants.CloudAWSName,
		RegionId:  region,
		Bucket:    bucket,
		Prefix:    prefix,
	}

	return
}
