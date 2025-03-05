package s3

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

func (s *S3) Backup() (err error) {
	repository, err := s.FormatRepository()
	if err != nil {
		return
	}

	var envs = s.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName:        s.RepoName,
		Path:            s.Path,
		LimitUploadRate: s.LimitUploadRate,
		RepoEnvs:        envs,
	}

	s.BaseHandler.SetOptions(opts)
	return s.BaseHandler.Backup()
}
