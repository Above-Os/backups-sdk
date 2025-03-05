package cos

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

func (c *Cos) Snapshots() error {
	repository, err := c.FormatRepository()
	if err != nil {
		return err
	}

	var envs = c.GetEnv(repository)
	var opts = &restic.ResticOptions{
		RepoName:        c.RepoName,
		RepoEnvs:        envs,
		LimitUploadRate: c.LimitUploadRate,
	}

	c.BaseHandler.SetOptions(opts)
	return c.BaseHandler.Snapshots()
}
