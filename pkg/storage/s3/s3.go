package s3

import (
	"errors"
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
)

type S3 struct {
	RepoName        string
	SnapshotId      string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Password        string
	LimitUploadRate string
	Path            string
	BaseHandler     base.Interface
}

func (s *S3) Backup() (err error) {
	repository, err := s.FormatRepository()
	if err != nil {
		return
	}

	var envs = s.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName:        s.RepoName,
		Path:            s.Path,
		LimitUploadRate: s.LimitUploadRate,
		RepoEnvs:        envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Backup()
}

func (s *S3) Restore() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}
	var envs = s.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Restore()
}

func (s *S3) Snapshots() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}

	var envs = s.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Snapshots()
}

func (s *S3) Regions() error {
	return nil
}

func (s *S3) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		AWS_ACCESS_KEY_ID:     s.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.SecretAccessKey,
		RESTIC_REPOSITORY:     repository,
		RESTIC_PASSWORD:       s.Password,
	}
	return envs
}

func (s *S3) FormatRepository() (repository string, err error) {
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
		return "", fmt.Errorf("s3 endpoint %v is invalid", repoSplit)
	}
	var repoBase = repoSplit[0]
	var repoPrefix = ""
	if len(repoSplit) >= 2 {
		repoPrefix = fmt.Sprintf("%s/", repoSplit[1])
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

	repository = fmt.Sprintf("s3:s3.%s.%s/%s/%s%s", region, domainName, bucket, repoPrefix, s.RepoName)

	return
}
