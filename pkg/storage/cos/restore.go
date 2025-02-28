package cos

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (c *Cos) Restore() error {
	repository, err := c.formatCosRepository()
	if err != nil {
		return err
	}

	var resticEnv = c.getEnv(repository)
	logger.Debugf("cos restore env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), c.RepoName, "", resticEnv.ToMap(), nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := r.GetSnapshot(c.SnapshotId)
	if err != nil {
		return err
	}
	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("cos restore spanshot %s detail: %s", c.SnapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = r.Restore(c.SnapshotId, uploadPath, c.Path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore cos successful, data: %s", util.ToJSON(summary))
	}

	return nil
}
