package cos

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/restic"
	"olares.com/backups-sdk/pkg/storage/base"
	"olares.com/backups-sdk/pkg/storage/model"
	"olares.com/backups-sdk/pkg/utils"
)

type TencentCloud struct {
	RepoId                   string
	RepoName                 string
	SnapshotId               string
	Endpoint                 string
	AccessKey                string
	SecretAccessKey          string
	Password                 string
	CloudName                string
	RegionId                 string
	LimitUploadRate          string
	LimitDownloadRate        string
	Path                     string
	Files                    []string
	FilesPrefixPath          string
	Metadata                 string
	BaseHandler              base.Interface
	Operator                 string
	BackupType               string
	BackupAppTypeName        string
	BackupFileTypeSourcePath string
}

func (c *TencentCloud) Backup(ctx context.Context, dryRun bool, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error) {
	storageInfo, err = c.FormatRepository()
	if err != nil {
		return
	}

	var envs = c.GetEnv(storageInfo.Url) // cos backup
	var opts = &restic.ResticOptions{
		RepoId:                   c.RepoId,
		RepoName:                 c.RepoName,
		CloudName:                c.CloudName,
		RegionId:                 c.RegionId,
		Path:                     c.Path,
		Files:                    c.Files,
		FilesPrefixPath:          c.FilesPrefixPath,
		Metadata:                 c.Metadata,
		LimitUploadRate:          c.LimitUploadRate,
		Operator:                 c.Operator,
		BackupType:               c.BackupType,
		BackupAppTypeName:        c.BackupAppTypeName,
		BackupFileTypeSourcePath: c.BackupFileTypeSourcePath,
		RepoEnvs:                 envs,
	}

	logger.Debugf("cos backup env vars: %s", utils.Base64encode([]byte(envs.String())))

	c.BaseHandler.SetOptions(opts)

	backupSummary, err = c.BaseHandler.Backup(ctx, dryRun, progressCallback)
	return backupSummary, storageInfo, err
}

func (c *TencentCloud) Restore(ctx context.Context, progressCallback func(percentDone float64)) (map[string]*restic.RestoreSummaryOutput, string, uint64, error) {
	storageInfo, err := c.FormatRepository()
	if err != nil {
		return nil, "", 0, err
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

func (c *TencentCloud) GetSnapshot(ctx context.Context, snapshotId string) (*restic.SnapshotList, error) {
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
	return c.BaseHandler.GetSnapshot(ctx, snapshotId)
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

	var endpoint = c.Endpoint

	cosUrlInfo, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var host = cosUrlInfo.Host
	var bucket, region, prefix string

	if !strings.Contains(host, constants.StorageTencentDoman) {
		return nil, fmt.Errorf("host invalid, host: %s", host)
	}

	var hosts = strings.Split(host, ".")
	if len(hosts) != 4 && len(hosts) != 5 {
		return nil, fmt.Errorf("host invalid, host: %s, support format like: https://cos.MY_REGION.myqcloud.com/MY_BUCKET_NAME/PREFIX_PATH, https://MY_BUCKET_NAME.cos.MY_REGION.myqcloud.com/PREFIX_PATH", host)
	}

	var path = strings.Trim(cosUrlInfo.Path, "/")
	var paths = strings.Split(path, "/")
	if len(paths) < 1 {
		return nil, fmt.Errorf("bucket not exists, path: %s", path)
	}

	if len(hosts) == 4 { // cos.MY_REGION.myqcloud.com/MY_BUCKET_NAME
		bucket = paths[0]
		region = hosts[1]
		if len(paths) > 1 {
			prefix = fmt.Sprintf("%s/%s", strings.Join(paths[1:], "/"), constants.OlaresStorageDefaultPrefix)
		} else {
			prefix = constants.OlaresStorageDefaultPrefix
		}
	} else { // MY_BUCKET_NAME.cos.MY_REGION.myqcloud.com
		bucket = hosts[0]
		region = hosts[2]
		var s = strings.Join(paths[0:], "/")
		if s == "" {
			prefix = constants.OlaresStorageDefaultPrefix
		} else {
			prefix = fmt.Sprintf("%s/%s", s, constants.OlaresStorageDefaultPrefix)
		}
	}

	var repository = fmt.Sprintf("s3:%s://cos.%s.%s/%s/%s/%s",
		cosUrlInfo.Scheme, region, constants.StorageTencentDoman,
		bucket, prefix,
		utils.JoinName(utils.EncodeURLPart(c.RepoName), c.RepoId),
	)

	c.CloudName = constants.CloudTencentName
	c.RegionId = region

	storageInfo = &model.StorageInfo{
		Location:  constants.CloudTencentName,
		Url:       repository,
		CloudName: constants.CloudTencentName,
		RegionId:  region,
		Bucket:    bucket,
		Prefix:    prefix,
	}

	return
}
