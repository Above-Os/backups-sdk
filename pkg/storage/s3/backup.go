package s3

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (s *S3) Backup() error {
	repository, err := s.formatS3Repository()
	if err != nil {
		return err
	}

	var resticEnv = s.getEnv(repository)

	logger.Debugf("s3 backup env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), s.RepoName, "", resticEnv.ToMap(), nil)
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

	backupResult, err := r.Backup(s.RepoName, s.Path, "")
	if err != nil {
		return err
	}
	logger.Infof("Backup to S3 success, result id: %s", util.ToJSON(backupResult))

	return nil
}
