package s3

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (s *S3) Restore() error {
	repository, err := s.formatS3Repository()
	if err != nil {
		return err
	}

	var resticEnv = s.getEnv(repository)

	logger.Debugf("s3 restore env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), s.RepoName, "", resticEnv.ToMap(), nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := r.GetSnapshot(s.SnapshotId)
	if err != nil {
		return err
	}
	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("s3 restore spanshot %s detail: %s", s.SnapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = r.Restore(s.SnapshotId, uploadPath, s.Path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore s3 successful, data: %s", util.ToJSON(summary))
	}

	return nil
}
