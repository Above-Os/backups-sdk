package filesystem

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

type Filesystem struct {
	RepoId      string
	RepoName    string
	SnapshotId  string
	Endpoint    string
	Password    string
	Path        string
	Files       []string
	BaseHandler base.Interface
	Operator    string
}

func (f *Filesystem) Backup(ctx context.Context, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error) {
	storageInfo, err = f.FormatRepository()
	if err != nil {
		return
	}

	var envs = f.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:   f.RepoId,
		RepoName: f.RepoName,
		Path:     f.Path,
		Files:    f.Files,
		Operator: f.Operator,
		RepoEnvs: envs,
	}

	logger.Debugf("fs backup env vars: %s", utils.Base64encode([]byte(envs.String())))

	f.BaseHandler.SetOptions(opts)
	backupSummary, err = f.BaseHandler.Backup(ctx, progressCallback)

	if err != nil {
		files, e := f.getTmpFiles(storageInfo.Url)
		if e != nil {
			err = errors.Wrap(err, e.Error())
		} else if files != nil && len(files) > 0 {
			for _, fn := range files {
				if deleterr := utils.DeleteFile(fn); deleterr != nil {
					logger.Errorf("fs backup delete tmp file error: %v", deleterr)
				} else {
					logger.Debugf("fs backup delete tmp file successful: %s", fn)
				}
			}
		}
	}

	return backupSummary, storageInfo, err
}

func (f *Filesystem) Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error) {
	storageInfo, err := f.FormatRepository()
	if err != nil {
		return
	}
	var envs = f.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:     f.RepoId,
		RepoName:   f.RepoName,
		SnapshotId: f.SnapshotId,
		RepoEnvs:   envs,
		Path:       f.Path,
	}

	logger.Debugf("fs restore env vars: %s", utils.Base64encode([]byte(envs.String())))

	f.BaseHandler.SetOptions(opts)
	return f.BaseHandler.Restore(ctx, progressCallback)
}

func (f *Filesystem) Snapshots(ctx context.Context) (*restic.SnapshotList, error) {
	storageInfo, err := f.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = f.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:   f.RepoId,
		RepoName: f.RepoName,
		RepoEnvs: envs,
	}

	logger.Debugf("fs snapshots env vars: %s", utils.Base64encode([]byte(envs.String())))

	f.BaseHandler.SetOptions(opts)
	return f.BaseHandler.Snapshots(ctx)
}

func (f *Filesystem) Stats(ctx context.Context) (*restic.StatsContainer, error) {
	storageInfo, err := f.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = f.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:   f.RepoId,
		RepoName: f.RepoName,
		RepoEnvs: envs,
	}

	logger.Debugf("fs stats env vars: %s", utils.Base64encode([]byte(envs.String())))

	f.BaseHandler.SetOptions(opts)
	return f.BaseHandler.Stats(ctx)
}

func (f *Filesystem) Regions() ([]map[string]string, error) {
	return nil, nil
}

func (f *Filesystem) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		RESTIC_REPOSITORY: repository,
		RESTIC_PASSWORD:   f.Password,
	}
	return envs
}

func (f *Filesystem) FormatRepository() (storageInfo *model.StorageInfo, err error) {
	if err := f.setRepoDir(); err != nil {
		return nil, err
	}

	storageInfo = &model.StorageInfo{
		Location:  "filesystem",
		Url:       path.Join(f.Endpoint, constants.OlaresStorageDefaultPrefix, utils.JoinName(f.RepoName, f.RepoId)),
		CloudName: constants.CloudFilesystemName,
		RegionId:  "",
		Bucket:    "",
		Prefix:    "",
	}

	return storageInfo, nil
}

func (f *Filesystem) setRepoDir() error {
	var p = path.Join(f.Endpoint, constants.OlaresStorageDefaultPrefix, fmt.Sprintf("%s-%s", f.RepoName, f.RepoId))
	if !utils.IsExist(p) {
		if err := utils.CreateDir(p); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (f *Filesystem) getTmpFiles(root string) ([]string, error) {
	var files []string

	pattern := filepath.Join(root, "data", "*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.Contains(info.Name(), "-tmp-") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
