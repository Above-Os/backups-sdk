package cos

import (
	"errors"
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
)

const (
	TencentDomain = "myqcloud.com"
)

type Cos struct {
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

func (c *Cos) Regions() error {
	return nil
}

func (c *Cos) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		AWS_ACCESS_KEY_ID:     c.AccessKey,
		AWS_SECRET_ACCESS_KEY: c.SecretAccessKey,
		RESTIC_REPOSITORY:     repository,
		RESTIC_PASSWORD:       c.Password,
	}
	return envs
}

func (c *Cos) FormatRepository() (repository string, err error) {
	if c.Endpoint == "" {
		err = errors.New("cos endpoint is required")
		return
	}

	var endpoint = strings.TrimPrefix(c.Endpoint, "https://")
	endpoint = strings.TrimRight(endpoint, "/")
	if strings.EqualFold(endpoint, "") {
		err = fmt.Errorf("cos endpoint %s is invalid", endpoint)
		return
	}

	var repoSplit = strings.Split(endpoint, "/")
	if repoSplit == nil || len(repoSplit) < 2 {
		err = fmt.Errorf("cos endpoint %v is invalid", repoSplit)
		return
	}

	var repoBase = repoSplit[0]
	var repoBucket = repoSplit[1]
	var repoPrefix = ""
	if len(repoSplit) > 2 { // todo unittest
		repoPrefix = fmt.Sprintf("%s/", strings.Join(repoSplit[2:], "/"))
	}

	var repoBaseSplit = strings.SplitN(repoBase, ".", 3)
	if repoBaseSplit == nil || len(repoBaseSplit) != 3 {
		err = fmt.Errorf("cos endpoint %v is invalid", repoBaseSplit)
		return
	}
	if repoBaseSplit[0] != "cos" || repoBaseSplit[2] != TencentDomain {
		err = fmt.Errorf("cos endpoint %v is not myqcloud.com", repoBaseSplit)
		return
	}
	var repoRegion = repoBaseSplit[1]

	repository = fmt.Sprintf("s3:https://cos.%s.%s/%s/%s%s", repoRegion, TencentDomain, repoBucket, repoPrefix, c.RepoName)

	return
}
