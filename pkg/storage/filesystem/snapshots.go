package filesystem

import (
	"context"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (f *Filesystem) Snapshots() error {
	repository, err := f.FormatRepository()
	if err != nil {
		return err
	}

	envs := f.GetEnv(repository)

	logger.Debugf("snapshots from filesystem env vars: %s", util.Base64encode([]byte(envs.ToString())))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := restic.NewRestic(ctx, f.RepoName, envs, nil)
	if err != nil {
		return err
	}

	snapshots, err := r.GetSnapshots()
	if err != nil {
		return err
	}
	snapshots.PrintTable()

	return nil
}
