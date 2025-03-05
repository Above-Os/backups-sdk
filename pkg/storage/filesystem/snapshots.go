package filesystem

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

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
