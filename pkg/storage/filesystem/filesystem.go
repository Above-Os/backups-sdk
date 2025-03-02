package filesystem

import (
	"path"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
)

type Filesystem struct {
	RepoName   string
	SnapshotId string
	Endpoint   string
	Password   string
	Path       string
}

func (f *Filesystem) GetEnv(repository string) *restic.ResticEnv {
	var envs = &restic.ResticEnv{
		RESTIC_REPOSITORY: path.Join(f.Endpoint, f.RepoName),
		RESTIC_PASSWORD:   f.Password,
	}
	return envs
}

func (f *Filesystem) FormatRepository() (repository string, err error) {
	return "", nil
}

func (f *Filesystem) setRepoDir() {
	var p = path.Join(f.Endpoint, f.RepoName)
	if !util.IsExist(p) {
		util.CreateDir(p)
	}
}
