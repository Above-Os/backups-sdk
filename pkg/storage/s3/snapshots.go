package s3

import (
	"context"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (s *S3) Snapshots() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}

	envs := s.GetEnv(repository)

	logger.Debugf("snapshots from S3 env vars: %s", util.Base64encode([]byte(envs.ToString())))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := restic.NewRestic(ctx, s.RepoName, envs, nil)
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
