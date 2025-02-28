package filesystem

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (f *Filesystem) Backup() error {
	f.setRepoDir()
	var resticEnv = f.getEnv()
	var repoName = f.RepoName
	var path = f.Path

	logger.Debugf("fs backup env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), repoName, "", resticEnv.ToMap(), nil)
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

	backupResult, err := r.Backup(repoName, path, "")
	if err != nil {
		return err
	}
	logger.Infof("Backup to fs success, result id: %s", util.ToJSON(backupResult))

	return nil
}
