package s3

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (s *S3) Snapshots() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}

	var resticEnv = s.GetEnv(repository)
	logger.Debugf("s3 snapshots env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), s.RepoName, "", resticEnv.ToMap(), nil)
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
