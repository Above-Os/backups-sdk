package filesystem

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (f *Filesystem) Snapshots() error {
	var resticEnv = f.getEnv()
	logger.Debugf("fs snapshots env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), f.RepoName, "", resticEnv.ToMap(), nil)
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
