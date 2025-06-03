package space

import (
	"context"

	"github.com/pkg/errors"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/restic"
	"olares.com/backups-sdk/pkg/utils"
)

func (s *Space) Snapshots(ctx context.Context) (*restic.SnapshotList, error) {
	if err := s.getStsToken(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	storageInfo, err := s.FormatRepository()
	if err != nil {
		return nil, err
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:          s.RepoId,
		RepoName:        s.RepoName,
		RepoEnvs:        envs,
		LimitUploadRate: s.LimitUploadRate,
	}
	logger.Debugf("space snapshots env vars: %s", utils.Base64encode([]byte(envs.String())))

	r, err := restic.NewRestic(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	snapshots, err := r.GetSnapshots(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	snapshots.PrintTable()

	return snapshots, nil
}
