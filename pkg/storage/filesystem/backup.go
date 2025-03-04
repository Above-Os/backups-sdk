package filesystem

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (f *Filesystem) Backup() error {
	repository, err := f.FormatRepository()
	if err != nil {
		return err
	}

	envs := f.GetEnv(repository)

	r, err := restic.NewRestic(context.Background(), f.RepoName, envs, nil)
	if err != nil {
		return err
	}

	_, initRepo, err := r.Init()
	if err != nil {
		return err
	}

	if !initRepo {
		if err = r.Repair(); err != nil {
			return err
		}
	}

	backupResult, err := r.Backup(f.Path, "")
	if err != nil {
		return err
	}
	logger.Infof("Backup to filesystem success, result id: %s", util.ToJSON(backupResult))

	return nil
}
