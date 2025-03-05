package filesystem

import (
	"path"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
)

type Filesystem struct {
	RepoName    string
	SnapshotId  string
	Endpoint    string
	Password    string
	Path        string
	BaseHandler base.Interface
}

func (f *Filesystem) Regions() error {
	return nil
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
	if !util.IsExist(p) {
		if err := util.CreateDir(p); err != nil {
			return err
		}
		return nil
	}
	return nil
}
