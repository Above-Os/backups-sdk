package cos

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
	Operator          string
}

func (c *TencentCloud) Backup(ctx context.Context, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error) {
	storageInfo, err = c.FormatRepository()
	if err != nil {
		return
	}

	var envs = c.GetEnv(storageInfo.Url) // cos backup
	var opts = &restic.ResticOptions{
		RepoName:        c.RepoName,
		CloudName:       c.CloudName,
		RegionId:        c.RegionId,
		Path:            c.Path,
		LimitUploadRate: c.LimitUploadRate,
		RepoEnvs:        envs,
	}

	logger.Debugf("cos backup env vars: %s", utils.Base64encode([]byte(envs.String())))

	c.BaseHandler.SetOptions(opts)

	backupSummary, err = c.BaseHandler.Backup(ctx, progressCallback)
	return backupSummary, storageInfo, err
}

func (c *TencentCloud) Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error) {
	storageInfo, err := c.FormatRepository()
	if err != nil {
		return
	}

	var envs = c.GetEnv(storageInfo.Url) // cos restore
	var opts = &restic.ResticOptions{
		RepoName:          c.RepoName,
		CloudName:         c.CloudName,
		RegionId:          c.RegionId,
		SnapshotId:        c.SnapshotId,
		RepoEnvs:          envs,
		Path:              c.Path,
		LimitDownloadRate: c.LimitDownloadRate,
	}

	logger.Debugf("cos restore env vars: %s", utils.Base64encode([]byte(envs.String())))

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Restore(ctx, progressCallback)
}

func (c *TencentCloud) Snapshots(ctx context.Context) error {
	storageInfo, err := c.FormatRepository()
	if err != nil {
		return err
	}

	var envs = c.GetEnv(storageInfo.Url) // cos snapshot
	var opts = &restic.ResticOptions{
		RepoName:  c.RepoName,
		CloudName: c.CloudName,
		RegionId:  c.RegionId,
		RepoEnvs:  envs,
	}

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Snapshots(ctx)
}

func (c *TencentCloud) Stats(ctx context.Context) (*restic.StatsContainer, error) {
	storageInfo, err := c.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = c.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoName:  c.RepoName,
		CloudName: c.CloudName,
		RegionId:  c.RegionId,
		RepoEnvs:  envs,
	}

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Stats(ctx)
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

func (c *TencentCloud) FormatRepository() (storageInfo *model.StorageInfo, err error) {
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

	var repository = fmt.Sprintf("s3:https://cos.%s.%s/%s/%s%s", repoRegion, domainName, repoBucket, repoPrefix, c.RepoName)

	c.CloudName = constants.CloudTencentName
	c.RegionId = repoRegion

	storageInfo = &model.StorageInfo{
		Location:  "tencentcloud",
		Url:       repository,
		CloudName: constants.CloudTencentName,
		RegionId:  repoRegion,
		Bucket:    repoBucket,
		Prefix:    repoPrefix,
	}

	return
}
