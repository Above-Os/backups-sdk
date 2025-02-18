package backupssdk

import (
	"fmt"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/backups-sdk/pkg/common"
	storageprovider "bytetrade.io/web3os/backups-sdk/pkg/storage"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func init() {
	_, err := util.GetCommand("restic")
	if err != nil {
		panic(err)
	}
}

type BackupClient struct {
	option storageprovider.BackupOption
}

type BackupClientOption struct {
	RepoName        string
	OlaresId        string
	StorageLocation string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string

	Path            string
	CloudApiMirror  string
	LimitUploadRate string
	BaseDir         string
	Version         string
	Logger          *zap.SugaredLogger
}

// + backup

func NewBackupClient(opt *BackupClientOption,
) *BackupClient {
	var o = storageprovider.BackupOption{
		RepoName:        opt.RepoName,
		OlaresId:        opt.OlaresId,
		StorageLocation: opt.StorageLocation,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		UploadPath:      opt.Path,
		CloudApiMirror:  opt.CloudApiMirror,
		LimitUploadRate: opt.LimitUploadRate,
		BaseDir:         opt.BaseDir,
		Version:         opt.Version,
	}

	var client = &BackupClient{
		option: o,
	}

	client.setLogger(opt.BaseDir, opt.Version, opt.Logger)

	return client
}

func (c *BackupClient) Backup() error {
	if !util.IsExist(c.option.UploadPath) {
		return errors.WithStack(fmt.Errorf("backup path not exist: %s", c.option.UploadPath))
	}

	if c.option.StorageLocation == common.StorageLocationFilesystem && !util.IsExist(c.option.Endpoint) {
		return errors.WithStack(fmt.Errorf("backup to filesystem repository not exist: %s", c.option.Endpoint))
	}

	u := &storageprovider.Backup{}
	return u.Backup(c.option)
}

func (c *BackupClient) setLogger(baseDir string, version string, log *zap.SugaredLogger) {
	if log != nil {
		logger.SetLogger(log)
		return
	}

	installerPath := filepath.Join(baseDir, "versions", fmt.Sprintf("v%s", version))
	if err := util.CreateDir(installerPath); err != nil {
		panic(err)
	}

	jsonLogDir := path.Join(baseDir, "logs")
	consoleLogDir := path.Join(installerPath, "logs", "backups_backup.log")
	logger.InitLog(jsonLogDir, consoleLogDir, true)
}

// + restore

type RestoreClient struct {
	option storageprovider.RestoreOption
}

type RestoreClientOption struct {
	RepoName          string
	SnapshotId        string
	OlaresId          string
	StorageLocation   string
	Endpoint          string
	AccessKey         string
	SecretAccessKey   string
	TargetPath        string
	CloudApiMirror    string
	LimitDownloadRate string
	BaseDir           string
	Version           string
	Logger            *zap.SugaredLogger
}

func NewRestoreClient(opt *RestoreClientOption) *RestoreClient {
	var o = storageprovider.RestoreOption{
		RepoName:          opt.RepoName,
		SnapshotId:        opt.SnapshotId,
		OlaresId:          opt.OlaresId,
		StorageLocation:   opt.StorageLocation,
		Endpoint:          opt.Endpoint,
		AccessKey:         opt.AccessKey,
		SecretAccessKey:   opt.SecretAccessKey,
		TargetPath:        opt.TargetPath,
		CloudApiMirror:    opt.CloudApiMirror,
		LimitDownloadRate: opt.LimitDownloadRate,
		BaseDir:           opt.BaseDir,
		Version:           opt.Version,
	}

	var client = &RestoreClient{
		option: o,
	}

	client.setLogger(opt.BaseDir, opt.Version, opt.Logger)

	return client
}

func (c *RestoreClient) Restore() error {
	if c.option.SnapshotId == "" {
		return errors.WithStack(fmt.Errorf("snapshot-id is empty"))
	}

	if !util.IsExist(c.option.TargetPath) {
		return errors.WithStack(fmt.Errorf("restore path not found"))
	}

	d := &storageprovider.Restore{}

	return d.Restore(c.option)
}

func (c *RestoreClient) setLogger(baseDir string, version string, log *zap.SugaredLogger) {
	if log != nil {
		logger.SetLogger(log)
		return
	}

	installerPath := filepath.Join(baseDir, "versions", fmt.Sprintf("v%s", version))
	if err := util.CreateDir(installerPath); err != nil {
		panic(err)
	}

	jsonLogDir := path.Join(baseDir, "logs")
	consoleLogDir := path.Join(installerPath, "logs", "backups_restore.log")
	logger.InitLog(jsonLogDir, consoleLogDir, true)
}

// + snapshots
type SnapshotsClient struct {
	option storageprovider.SnapshotsOption
}

type SnapshotsClientOption struct {
	RepoName        string
	OlaresId        string
	StorageLocation string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	CloudApiMirror  string
	BaseDir         string
	Version         string
	Logger          *zap.SugaredLogger
}

func NewSnapshotsClient(opt *SnapshotsClientOption) *SnapshotsClient {
	var o = storageprovider.SnapshotsOption{
		RepoName:        opt.RepoName,
		OlaresId:        opt.OlaresId,
		StorageLocation: opt.StorageLocation,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		CloudApiMirror:  opt.CloudApiMirror,
		BaseDir:         opt.BaseDir,
		Version:         opt.Version,
	}

	var client = &SnapshotsClient{
		option: o,
	}

	client.setLogger(opt.BaseDir, opt.Version, opt.Logger)

	logger.Debugf("snapshots options: %s", util.ToJSON(o))

	return client
}

func (c *SnapshotsClient) Snapshots() error {
	u := &storageprovider.Snapshots{}
	return u.Snapshots(c.option)
}

func (c *SnapshotsClient) setLogger(baseDir string, version string, log *zap.SugaredLogger) {
	if log != nil {
		logger.SetLogger(log)
		return
	}

	installerPath := filepath.Join(baseDir, "versions", fmt.Sprintf("v%s", version))
	if err := util.CreateDir(installerPath); err != nil {
		panic(err)
	}

	jsonLogDir := path.Join(baseDir, "logs")
	consoleLogDir := path.Join(installerPath, "logs", "backups_snapshots.log")
	logger.InitLog(jsonLogDir, consoleLogDir, true)
}
