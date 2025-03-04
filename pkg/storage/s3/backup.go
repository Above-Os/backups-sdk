package s3

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (s *S3) Backup() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}

	envs := s.GetEnv(repository)

	r, err := restic.NewRestic(context.Background(), s.RepoName, envs, &restic.Option{LimitUploadRate: s.LimitUploadRate})
	if err != nil {
		return err
	}

	_, initRepo, err := r.Init()
	if err != nil {
		return err
	}

	if !initRepo {
		if err = r.Repair(); err != nil {
			return err
		}
	}

	backupResult, err := r.Backup(s.Path, "")
	if err != nil {
		return err
	}
	logger.Infof("Backup to S3 success, result id: %s", util.ToJSON(backupResult))

	return nil

}
