package space

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

func (s *Space) Snapshots() error {
	var repoName = s.RepoName

	if err := s.getStsToken(); err != nil {
		return errors.WithStack(err)
	}

	var envs = s.GetEnv(repoName)
	var opts = &restic.ResticOptions{
		RepoName:        s.RepoName,
		RepoEnvs:        envs,
		LimitUploadRate: s.LimitUploadRate,
	}
	logger.Debugf("space snapshots env vars: %s", utils.Base64encode([]byte(envs.String())))

	r, err := restic.NewRestic(context.Background(), opts)
	if err != nil {
		return err
	}

	snapshots, err := r.GetSnapshots(nil)
	if err != nil {
		return errors.WithStack(err)
	}
	snapshots.PrintTable()

	return nil
}
