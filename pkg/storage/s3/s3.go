package s3

import (
	"errors"
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

const (
	awsDomain = "amazonaws.com"
)

type S3 struct {
	RepoName        string
	SnapshotId      string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Password        string
	Path            string
}

func (s *S3) GetEnv(repository string) *restic.ResticEnv {
	var envs = &restic.ResticEnv{
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
	if repoBaseSplit[2] != awsDomain {
		err = fmt.Errorf("s3 endpoint %s is not %s", repoBaseSplit[2], awsDomain)
		return
	}
	var bucket = repoBaseSplit[0]
	var region = repoBaseSplit[1]

	repository = fmt.Sprintf("s3:s3.%s.%s/%s/%s%s", region, awsDomain, bucket, repoPrefix, s.RepoName)

	return
}
