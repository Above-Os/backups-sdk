package filesystem

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (f *Filesystem) Restore() error {
	repository, err := f.FormatRepository()
	if err != nil {
		return err
	}
	envs := f.GetEnv(repository)

	logger.Debugf("restore from filesystem env vars: %s", util.Base64encode([]byte(envs.ToString())))

	re, err := restic.NewRestic(context.Background(), f.RepoName, envs, nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := re.GetSnapshot(f.SnapshotId)
	if err != nil {
		return err
	}

	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("restore from filesystem spanshot %s detail: %s", f.SnapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = re.Restore(f.SnapshotId, uploadPath, f.Path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore from filesystem successful, data: %s", util.ToJSON(summary))
	}

	return nil
}
