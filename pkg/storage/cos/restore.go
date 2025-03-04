package cos

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (c *Cos) Restore() error {
	repository, err := c.FormatRepository()
	if err != nil {
		return err
	}
	envs := c.GetEnv(repository)

	logger.Debugf("restore from Tencent COS env vars: %s", util.Base64encode([]byte(envs.ToString())))

	re, err := restic.NewRestic(context.Background(), c.RepoName, envs, nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := re.GetSnapshot(c.SnapshotId)
	if err != nil {
		return err
	}

	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("restore from Tencent COS spanshot %s detail: %s", c.SnapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = re.Restore(c.SnapshotId, uploadPath, c.Path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore from Tencent COS successful, data: %s", util.ToJSON(summary))
	}

	return nil

}
