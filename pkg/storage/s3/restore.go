package s3

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

func (s *S3) Restore() error {
	repository, err := s.FormatRepository()
	if err != nil {
		return err
	}
	var envs = s.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName: s.RepoName,
		RepoEnvs: envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Restore()

}
