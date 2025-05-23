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

// {bucket}.{region}.amazonaws.com/{prefix}
// {bucket}.s3.{region}.amazonaws.com/{prefix}
// s3.{region}.amazonaws.com/{bucket}/{prefix}
func (s *Aws) FormatRepository() (storageInfo *model.StorageInfo, err error) {
	if s.Endpoint == "" {
		err = errors.New("s3 endpoint is required")
		return
	}

	b, r, p, ep, err := s3format(s.Endpoint, s.RepoName, s.RepoId)
	if err != nil {
		return nil, err
	}

	storageInfo = &model.StorageInfo{
		Location:  "awss3",
		Url:       ep,
		CloudName: constants.CloudAWSName,
		RegionId:  r,
		Bucket:    b,
		Prefix:    p,
	}

	return
}

func s3format(rawurl string, repoName, repoId string) (bucket, region, prefix, endpoint string, err error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", "", "", "", err
	}

	host := u.Host
	path := strings.TrimPrefix(u.Path, "/")

	parts := strings.Split(host, ".")

	if len(parts) < 3 || parts[len(parts)-2] != "amazonaws" || parts[len(parts)-1] != "com" {
		return "", "", "", "", errors.New("host is not a valid amazonaws.com domain")
	}

	switch {
	case strings.HasPrefix(host, "s3."):
		if len(parts) < 4 {
			return "", "", "", "", errors.New("host format invalid for s3.region.amazonaws.com")
		}
		region = parts[1]
		pathParts := strings.SplitN(path, "/", 2)
		if len(pathParts) < 1 || pathParts[0] == "" {
			return "", "", "", "", errors.New("bucket not found in path")
		}
		bucket = pathParts[0]
		if len(pathParts) == 2 {
			prefix = pathParts[1]
		} else {
			prefix = ""
		}

	case len(parts) >= 5 && parts[1] == "s3":
		bucket = parts[0]
		region = parts[2]
		prefix = path

	case len(parts) >= 4:
		bucket = parts[0]
		region = parts[1]
		prefix = path

	default:
		return "", "", "", "", errors.New("host format not recognized")
	}

	if prefix != "" {
		endpoint = fmt.Sprintf("s3:https://s3.%s.amazonaws.com/%s/%s/%s/%s", region, bucket, prefix, constants.OlaresStorageDefaultPrefix, utils.JoinName(utils.EncodeURLPart(repoName), repoId))
	} else {
		endpoint = fmt.Sprintf("s3:https://s3.%s.amazonaws.com/%s/%s/%s", region, bucket, constants.OlaresStorageDefaultPrefix, utils.JoinName(utils.EncodeURLPart(repoName), repoId))
	}

	return bucket, region, prefix, endpoint, nil
}
