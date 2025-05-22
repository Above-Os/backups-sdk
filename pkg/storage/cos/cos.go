package cos

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

type TencentCloud struct {
	RepoId            string
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
	Files             []string
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
		RepoId:          c.RepoId,
		RepoName:        c.RepoName,
		CloudName:       c.CloudName,
		RegionId:        c.RegionId,
		Path:            c.Path,
		Files:           c.Files,
		LimitUploadRate: c.LimitUploadRate,
		Operator:        c.Operator,
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
		RepoId:            c.RepoId,
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

func (c *TencentCloud) Snapshots(ctx context.Context) (*restic.SnapshotList, error) {
	storageInfo, err := c.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = c.GetEnv(storageInfo.Url) // cos snapshot
	var opts = &restic.ResticOptions{
		RepoId:    c.RepoId,
		RepoName:  c.RepoName,
		CloudName: c.CloudName,
		RegionId:  c.RegionId,
		RepoEnvs:  envs,
	}

	logger.Debugf("cos snapshots env vars: %s", utils.Base64encode([]byte(envs.String())))

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
		RepoId:    c.RepoId,
		RepoName:  c.RepoName,
		CloudName: c.CloudName,
		RegionId:  c.RegionId,
		RepoEnvs:  envs,
	}

	logger.Debugf("cos stats env vars: %s", utils.Base64encode([]byte(envs.String())))

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

	var endpoint = strings.TrimRight(c.Endpoint, "/")

	cosUrlInfo, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var host = cosUrlInfo.Host
	var hosts = strings.Split(host, ".")
	if len(hosts) != 4 {
		return nil, fmt.Errorf("host invalid, host: %s", host)
	}

	if !strings.Contains(host, constants.StorageTencentDoman) {
		return nil, fmt.Errorf("host is not cos format, host: %s", host)
	}

	var region = hosts[1]

	var path = strings.TrimLeft(cosUrlInfo.Path, "/")
	var paths = strings.Split(path, "/")
	if len(paths) == 0 {
		return nil, fmt.Errorf("bucket not exists, path: %s", path)
	}

	var bucket = paths[0]
	var prefix string = constants.OlaresStorageDefaultPrefix
	if len(paths) > 1 {
		prefix = fmt.Sprintf("%s/%s", strings.Join(paths[1:], "/"), constants.OlaresStorageDefaultPrefix)
	}

	var repository = fmt.Sprintf("s3:%s://%s/%s/%s/%s", cosUrlInfo.Scheme, cosUrlInfo.Host, bucket, prefix, utils.JoinName(utils.EncodeURLPart(c.RepoName), c.RepoId))

	c.CloudName = constants.CloudTencentName
	c.RegionId = region

	storageInfo = &model.StorageInfo{
		Location:  "tencentcloud",
		Url:       repository,
		CloudName: constants.CloudTencentName,
		RegionId:  region,
		Bucket:    bucket,
		Prefix:    prefix,
	}

	return
}
