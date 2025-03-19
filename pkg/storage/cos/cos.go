package cos

import (
	"errors"
	"fmt"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
)

type TencentCloud struct {
	RepoName          string
	SnapshotId        string
	Endpoint          string
	AccessKey         string
	SecretAccessKey   string
	Password          string
	CloudName         string
	RegionId          string
	LimitUploadRate   string
	LimitDownloadRate string
	Path              string
	BaseHandler       base.Interface
}

func (c *TencentCloud) Backup() (backupSummary *restic.SummaryOutput, repo string, err error) {
	repository, err := c.FormatRepository()
	if err != nil {
		return
	}

	var envs = c.GetEnv(repository) // cos backup
	var opts = &restic.ResticOptions{
		RepoName:        c.RepoName,
		CloudName:       c.CloudName,
		RegionId:        c.RegionId,
		RepoEnvs:        envs,
		LimitUploadRate: c.LimitUploadRate,
	}

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Backup()
}

func (c *TencentCloud) Restore() error {
	repository, err := c.FormatRepository()
	if err != nil {
		return err
	}
	var envs = c.GetEnv(repository) // cos restore
	var opts = &restic.ResticOptions{
		RepoName:          c.RepoName,
		RepoEnvs:          envs,
		LimitDownloadRate: c.LimitDownloadRate,
	}

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Restore()
}

func (c *TencentCloud) Snapshots() error {
	repository, err := c.FormatRepository()
	if err != nil {
		return err
	}

	var envs = c.GetEnv(repository) // cos snapshot
	var opts = &restic.ResticOptions{
		RepoName:        c.RepoName,
		RepoEnvs:        envs,
		LimitUploadRate: c.LimitUploadRate,
	}

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Snapshots()
}

func (c *TencentCloud) Regions() ([]map[string]string, error) {
	return nil, nil
}

func (c *TencentCloud) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		AWS_ACCESS_KEY_ID:     c.AccessKey,
		AWS_SECRET_ACCESS_KEY: c.SecretAccessKey,
		RESTIC_REPOSITORY:     repository,
		RESTIC_PASSWORD:       c.Password,
	}
	return envs
}

func (c *TencentCloud) FormatRepository() (repository string, err error) {
	if c.Endpoint == "" {
		err = errors.New("cos endpoint is required")
		return
	}

	var domainName = constants.StorageTencentDoman
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
	if repoBaseSplit[0] != "cos" || repoBaseSplit[2] != domainName {
		err = fmt.Errorf("cos endpoint %v is not %s", repoBaseSplit, domainName)
		return
	}
	var repoRegion = repoBaseSplit[1]

	repository = fmt.Sprintf("s3:https://cos.%s.%s/%s/%s%s", repoRegion, domainName, repoBucket, repoPrefix, c.RepoName)

	c.CloudName = constants.CloudTencentName
	c.RegionId = repoRegion

	return
}
