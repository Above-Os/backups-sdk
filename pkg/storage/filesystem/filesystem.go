package filesystem

import (
	"path"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type Filesystem struct {
	RepoName    string
	SnapshotId  string
	Endpoint    string
	Password    string
	Path        string
	BaseHandler base.Interface
}

func (f *Filesystem) Backup() (backupSummary *restic.SummaryOutput, repo string, err error) {
	repository, err := f.FormatRepository()
	if err != nil {
		return
	}

	var envs = f.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName: f.RepoName,
		RepoEnvs: envs,
	}

	f.BaseHandler.SetOptions(opts)
	return f.BaseHandler.Backup()
}

func (f *Filesystem) Restore() error {
	repository, err := f.FormatRepository()
	if err != nil {
		return err
	}
	var envs = f.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName: f.RepoName,
		RepoEnvs: envs,
	}

	f.BaseHandler.SetOptions(opts)
	return f.BaseHandler.Restore()
}

func (f *Filesystem) Snapshots() error {
	repository, err := f.FormatRepository()
	if err != nil {
		return err
	}

	var envs = f.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName: f.RepoName,
		RepoEnvs: envs,
	}

	f.BaseHandler.SetOptions(opts)
	return f.BaseHandler.Snapshots()
}

func (f *Filesystem) Regions() ([]map[string]string, error) {
	return nil, nil
}

func (f *Filesystem) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		RESTIC_REPOSITORY: path.Join(f.Endpoint, f.RepoName),
		RESTIC_PASSWORD:   f.Password,
	}
	return envs
}

func (f *Filesystem) FormatRepository() (repository string, err error) {
	if err := f.setRepoDir(); err != nil {
		return "", err
	}
	return f.Endpoint, nil
}

func (f *Filesystem) setRepoDir() error {
	var p = path.Join(f.Endpoint, f.RepoName)
	if !utils.IsExist(p) {
		if err := utils.CreateDir(p); err != nil {
			return err
		}
		return nil
	}
	return nil
}
