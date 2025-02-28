package filesystem

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (f *Filesystem) Restore() error {
	var resticEnv = f.getEnv()

	logger.Debugf("fs restore env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), f.RepoName, "", resticEnv.ToMap(), nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := r.GetSnapshot(f.SnapshotId)
	if err != nil {
		return err
	}
	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("fs restore spanshot %s detail: %s", f.SnapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = r.Restore(f.SnapshotId, uploadPath, f.Path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore fs successful, data: %s", util.ToJSON(summary))
	}

	return nil
}
