package s3

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (s *S3) Restore() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}
	envs := s.GetEnv(repository)

	logger.Debugf("restore from S3 env vars: %s", util.Base64encode([]byte(envs.ToString())))

	re, err := restic.NewRestic(context.Background(), s.RepoName, envs, nil)
	if err != nil {
		return err
	}
	snapshotSummary, err := re.GetSnapshot(s.SnapshotId)
	if err != nil {
		return err
	}

	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("restore from S3 spanshot %s detail: %s", s.SnapshotId, util.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = re.Restore(s.SnapshotId, uploadPath, s.Path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore from S3 successful, data: %s", util.ToJSON(summary))
	}

	return nil

}
