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

	repoSuffix, err := utils.GetSuffix(s.StsToken.Prefix, "-")
	if err != nil {
		return nil, err
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:     s.RepoId,
		RepoName:   s.RepoName,
		RepoSuffix: repoSuffix,
		CloudName:  s.CloudName,
		RegionId:   s.RegionId,
		Operator:   s.Operator,
		RepoEnvs:   envs,
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

func (s *Space) GetSnapshot(ctx context.Context, snapshotId string) (*restic.SnapshotList, error) {
	if err := s.getStsToken(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	storageInfo, err := s.FormatRepository()
	if err != nil {
		return nil, err
	}

	repoSuffix, err := utils.GetSuffix(s.StsToken.Prefix, "-")
	if err != nil {
		return nil, err
	}

	var envs = s.GetEnv(storageInfo.Url)
	var opts = &restic.ResticOptions{
		RepoId:     s.RepoId,
		RepoName:   s.RepoName,
		RepoSuffix: repoSuffix,
		CloudName:  s.CloudName,
		RegionId:   s.RegionId,
		Operator:   s.Operator,
		RepoEnvs:   envs,
	}
	logger.Debugf("space snapshot env vars: %s", utils.Base64encode([]byte(envs.String())))

	r, err := restic.NewRestic(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	snapshot, err := r.GetSnapshot(snapshotId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var list restic.SnapshotList
	list = append(list, snapshot)

	return &list, nil
}
