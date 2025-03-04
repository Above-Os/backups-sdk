package space

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

func (s *Space) Snapshots() error {
	var repoName = s.RepoName

	if err := s.getStsToken(DefaultLocation, DefaultRegion); err != nil {
		return errors.WithStack(err)
	}

	envs := s.GetEnv(repoName)
	logger.Debugf("space snapshots env vars: %s", util.Base64encode([]byte(envs.ToString())))

	r, err := restic.NewRestic(context.Background(), repoName, envs, nil)
	if err != nil {
		return err
	}

	snapshots, err := r.GetSnapshots()
	if err != nil {
		return errors.WithStack(err)
	}
	snapshots.PrintTable()

	return nil
}
